package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/arnoldreis/resiliq/internal/database"
	"github.com/arnoldreis/resiliq/internal/queue"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	// inicializa conexão com o banco
	db, err := database.NewConnection("localhost", 5432, "user", "password", "resiliq")
	if err != nil {
		log.Fatalf("Falha na conexão com o banco: %v", err)
	}

	producer := queue.NewProducer(db)

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	/**
	 * Endpoint para enfileirar novas mensagens
	 */
	r.Post("/enqueue", func(w http.ResponseWriter, r *http.Request) {
		var payload json.RawMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Payload inválido", http.StatusBadRequest)
			return
		}

		// extrai a chave de idempotência do header se existir
		var idKeyPtr *string
		if idKey := r.Header.Get("X-Idempotency-Key"); idKey != "" {
			idKeyPtr = &idKey
		}

		if err := producer.Enqueue(r.Context(), payload, idKeyPtr); err != nil {
			log.Printf("Erro ao enfileirar: %v", err)
			http.Error(w, "Erro interno", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"status": "accepted"})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Produtor rodando na porta %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}
