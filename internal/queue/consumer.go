package queue

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/models"
)

type Consumer struct {
	db *database.DB
}

/**
 * NewConsumer cria uma nova instância do consumidor
 */
func NewConsumer(db *database.DB) *Consumer {
	return &Consumer{db: db}
}

/**
 * Start inicia o loop do worker para processar mensagens
 * @param ctx - Contexto para cancelamento do worker
 */
func (c *Consumer) Start(ctx context.Context) {
	log.Println("Consumidor iniciado, aguardando mensagens...")
	
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Encerrando consumidor...")
			return
		case <-ticker.C:
			if err := c.processNext(ctx); err != nil && err != sql.ErrNoRows {
				log.Printf("Erro ao processar mensagem: %v", err)
			}
		}
	}
}

func (c *Consumer) processNext(ctx context.Context) error {
	// inicia transação para garantir exclusividade
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// busca a próxima mensagem PENDING usando SELECT FOR UPDATE SKIP LOCKED
	// essa é a "mágica" para concorrência segura em sistemas de fila via RDBMS
	query := `
		SELECT id, payload, retry_count
		FROM messages
		WHERE status = 'PENDING'
		AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	var msg models.Message
	err = tx.QueryRowContext(ctx, query).Scan(&msg.ID, &msg.Payload, &msg.RetryCount)
	if err != nil {
		return err // pode ser sql.ErrNoRows
	}

	// marca como PROCESSING
	_, err = tx.ExecContext(ctx, "UPDATE messages SET status = 'PROCESSING', updated_at = NOW() WHERE id = $1", msg.ID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar para PROCESSING: %w", err)
	}

	// simula processamento
	log.Printf("Processando mensagem %s: %s", msg.ID, string(msg.Payload))
	time.Sleep(500 * time.Millisecond) // simulação de I/O

	// marca como COMPLETED
	_, err = tx.ExecContext(ctx, "UPDATE messages SET status = 'COMPLETED', updated_at = NOW() WHERE id = $1", msg.ID)
	if err != nil {
		return fmt.Errorf("erro ao marcar como COMPLETED: %w", err)
	}

	return tx.Commit()
}
