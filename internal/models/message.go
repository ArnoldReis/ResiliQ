package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type MessageStatus string

const (
	StatusPending    MessageStatus = "PENDING"
	StatusProcessing MessageStatus = "PROCESSING"
	StatusCompleted  MessageStatus = "COMPLETED"
	StatusFailed     MessageStatus = "FAILED"
)

/**
 * Message representa uma entrada na fila de mensagens
 */
type Message struct {
	ID             uuid.UUID       `json:"id"`
	Payload        json.RawMessage `json:"payload"`
	IdempotencyKey *string         `json:"idempotency_key,omitempty"`
	Status         MessageStatus   `json:"status"`
	RetryCount     int             `json:"retry_count"`
	NextRetryAt *time.Time      `json:"next_retry_at"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}
