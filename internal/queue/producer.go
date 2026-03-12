package queue

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/metrics"
)

type Producer struct {
	db *database.DB
}

/**
 * NewProducer cria uma nova instância do produtor
 * @param db - Instância da conexão com o banco
 */
func NewProducer(db *database.DB) *Producer {
	return &Producer{db: db}
}

/**
 * Enqueue adiciona um payload à fila de mensagens com suporte opcional a idempotência
 * @param ctx - Contexto da requisição
 * @param payload - Os dados da mensagem em formato JSON
 * @param idempotencyKey - Chave única para evitar duplicidade (opcional)
 * @returns Erro se a inserção falhar
 */
func (p *Producer) Enqueue(ctx context.Context, payload json.RawMessage, idempotencyKey *string) error {
	query := `
		INSERT INTO messages (payload, idempotency_key, status)
		VALUES ($1, $2, 'PENDING')
		ON CONFLICT (idempotency_key) DO NOTHING
	`
	_, err := p.db.ExecContext(ctx, query, payload, idempotencyKey)
	if err != nil {
		return fmt.Errorf("erro ao enfileirar mensagem: %w", err)
	}

	// incrementa métrica de sucesso
	metrics.MessagesEnqueued.Inc()
	return nil
}
