package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	MessagesEnqueued = promauto.NewCounter(prometheus.CounterOpts{
		Name: "resiliq_messages_enqueued_total",
		Help: "O número total de mensagens enfileiradas",
	})

	MessagesProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "resiliq_messages_processed_total",
		Help: "O número total de mensagens processadas, rotuladas por status",
	}, []string{"status"})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "resiliq_processing_duration_seconds",
		Help:    "Tempo gasto processando cada mensagem",
		Buckets: prometheus.DefBuckets,
	})
)
