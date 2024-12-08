package database

import (
	"context"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewDatabaseConnection(logger *zap.SugaredLogger) *pgxpool.Pool {
	dsn, found := os.LookupEnv("DATABASE_DSN")
	if !found {
		logger.Fatalf("Missing DATABASE_DSN environment variable")
	}
	conn, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Fatalf("Unable to connect to database: %v\n", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		logger.Fatalf("Unable to ping database: %v\n", err)
	}

	return conn
}
