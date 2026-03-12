package queue

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/models"
)

/**
 * TestIntegration e2e-ish test for the producer-consumer flow
 */
func TestIntegration(t *testing.T) {
	// Skip if not running in an environment with DB
	// In a real scenario, we'd use testcontainers or a dedicated test DB
	db, err := database.NewConnection("localhost", 5432, "user", "password", "resiliq")
	if err != nil {
		t.Skip("Postgres não disponível, pulando teste de integração")
	}

	producer := NewProducer(db)
	consumer := NewConsumer(db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. Enfileira uma mensagem
	payload := json.RawMessage(`{"test": "integration"}`)
	err = producer.Enqueue(ctx, payload, nil)
	if err != nil {
		t.Fatalf("Erro ao enfileirar: %v", err)
	}

	// 2. Roda o processamento uma vez
	err = consumer.processNext(ctx)
	if err != nil {
		t.Fatalf("Erro ao processar: %v", err)
	}

	// 3. Verifica no banco se o status mudou para COMPLETED
	var status models.MessageStatus
	err = db.QueryRowContext(ctx, "SELECT status FROM messages ORDER BY created_at DESC LIMIT 1").Scan(&status)
	if err != nil {
		t.Fatalf("Erro ao consultar status: %v", err)
	}

	if status != models.StatusCompleted {
		t.Errorf("Esperado COMPLETED, obtido %s", status)
	}
}
