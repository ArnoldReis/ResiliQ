package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/logger"
	"github.com/arnoldreis/resiliq/internal/queue"
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

	// captura sinais do SO
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Sinal recebido: %v, iniciando shutdown...", sig)
		cancel()
	}()

	// inicia o consumidor (bloqueante)
	consumer.Start(ctx)
	
	log.Println("Consumidor encerrado com sucesso")
}
