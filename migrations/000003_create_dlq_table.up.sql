CREATE TABLE IF NOT EXISTS dead_letter_queue (
    id UUID PRIMARY KEY,
    payload JSONB NOT NULL,
    idempotency_key VARCHAR(255),
    error_message TEXT,
    failed_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);
