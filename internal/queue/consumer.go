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
	query := `
		SELECT id, payload, idempotency_key, retry_count
		FROM messages
		WHERE status = 'PENDING'
		AND (next_retry_at IS NULL OR next_retry_at <= NOW())
		ORDER BY created_at ASC
		FOR UPDATE SKIP LOCKED
		LIMIT 1
	`

	var msg models.Message
	err = tx.QueryRowContext(ctx, query).Scan(&msg.ID, &msg.Payload, &msg.IdempotencyKey, &msg.RetryCount)
	if err != nil {
		return err // pode ser sql.ErrNoRows
	}

	// marca como PROCESSING
	_, err = tx.ExecContext(ctx, "UPDATE messages SET status = 'PROCESSING', updated_at = NOW() WHERE id = $1", msg.ID)
	if err != nil {
		return fmt.Errorf("erro ao atualizar para PROCESSING: %w", err)
	}

	// simulando processamento (aqui entraria a lógica de negócio)
	err = c.handleMessage(msg)

	if err != nil {
		log.Printf("Falha ao processar mensagem %s: %v", msg.ID, err)
		return c.handleFailure(ctx, tx, msg, err)
	}

	// marca como COMPLETED
	_, err = tx.ExecContext(ctx, "UPDATE messages SET status = 'COMPLETED', updated_at = NOW() WHERE id = $1", msg.ID)
	if err != nil {
		return fmt.Errorf("erro ao marcar como COMPLETED: %w", err)
	}

	return tx.Commit()
}

func (c *Consumer) handleMessage(msg models.Message) error {
	// simula uma falha aleatória para testar retries (opcional)
	// if time.Now().UnixNano()%2 == 0 {
	// 	return fmt.Errorf("erro temporário simulado")
	// }
	
	log.Printf("Processando mensagem %s: %s", msg.ID, string(msg.Payload))
	time.Sleep(200 * time.Millisecond)
	return nil
}

func (c *Consumer) handleFailure(ctx context.Context, tx *sql.Tx, msg models.Message, processErr error) error {
	maxRetries := 5 // limite definido na Fase 4

	if msg.RetryCount >= maxRetries {
		// move para a DLQ
		log.Printf("Mensagem %s atingiu limite de retries, movendo para DLQ", msg.ID)
		
		insertDlq := `INSERT INTO dead_letter_queue (id, payload, idempotency_key, error_message) VALUES ($1, $2, $3, $4)`
		_, err := tx.ExecContext(ctx, insertDlq, msg.ID, msg.Payload, msg.IdempotencyKey, processErr.Error())
		if err != nil {
			return fmt.Errorf("erro ao inserir na DLQ: %w", err)
		}

		_, err = tx.ExecContext(ctx, "DELETE FROM messages WHERE id = $1", msg.ID)
		if err != nil {
			return fmt.Errorf("erro ao deletar da messages após DLQ: %w", err)
		}
	} else {
		// agende o retry com Exponential Backoff (2^retry_count segundos)
		backoff := time.Duration(1<<uint(msg.RetryCount)) * time.Second
		nextRetry := time.Now().Add(backoff)
		
		log.Printf("Agendando retry para mensagem %s em %v", msg.ID, backoff)

		updateRetry := `
			UPDATE messages 
			SET status = 'PENDING', 
			    retry_count = retry_count + 1, 
			    next_retry_at = $2, 
			    updated_at = NOW() 
			WHERE id = $1
		`
		_, err := tx.ExecContext(ctx, updateRetry, msg.ID, nextRetry)
		if err != nil {
			return fmt.Errorf("erro ao atualizar retry: %w", err)
		}
	}

	return tx.Commit()
}
