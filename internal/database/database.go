package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/arnoldreis/resiliq/internal/logger"
	_ "github.com/lib/pq"
)

/**
 * DB holds the database connection pool
 */
type DB struct {
	*sql.DB
}

/**
 * NewConnection cria uma nova conexão com o PostgreSQL
 * @param host - Host do banco
 * @param port - Porta do banco
 * @param user - Usuário do banco
 * @param password - Senha do banco
 * @param dbname - Nome do banco
 * @returns Instância do DB ou erro
 */
func NewConnection(host string, port int, user, password, dbname string) (*DB, error) {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// abre a conexão sem testar imediatamente
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir conexão: %w", err)
	}

	// configurações recomendadas para pooling
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// testa a conexão de fato
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("erro ao pingar banco: %w", err)
	}

	logger.GetLogger().Info("Conexão com o banco de dados estabelecida com sucesso")
	return &DB{db}, nil
}
