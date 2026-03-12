package main

import (
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/logger"
	"github.com/arnoldreis/resiliq/internal/queue"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

func main() {
	// inicializa o logger estruturado
	logger.InitLogger()
	log := logger.GetLogger()
	defer log.Sync()

	// inicializa conexão com o banco
	db, err := database.NewConnection("localhost", 5432, "user", "password", "resiliq")
	if err != nil {
		log.Fatal("Falha na conexão com o banco", zap.Error(err))
	}

	consumer := queue.NewConsumer(db)

	// contexto para graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// inicia servidor de métricas para o consumidor (porta 9091)
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		log.Info("Servidor de métricas do consumidor rodando na porta 9091")
		if err := http.ListenAndServe(":9091", mux); err != nil {
			log.Error("Falha ao iniciar servidor de métricas", zap.Error(err))
		}
	}()

	// captura sinais do SO
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Info("Sinal recebido, iniciando shutdown...", zap.String("signal", sig.String()))
		cancel()
	}()

	// inicia o consumidor (bloqueante)
	consumer.Start(ctx)
	
	log.Info("Consumidor encerrado com sucesso")
}
