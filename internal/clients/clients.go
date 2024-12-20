package clients

import (
	"chatapp/internal/config"
	"chatapp/internal/logger"
	"context"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	ErrDbConnection = errors.New("cannot connect to DB")
	ErrDbPing       = errors.New("cannot ping DB")
)

type Clients struct {
	Postgres *sqlx.DB
}

func NewClients(ctx context.Context, cfg *config.Config, logger logger.Logger) (*Clients, error) {
	db, err := connectToDB(ctx, cfg)
	if err != nil {
		return nil, err
	}
	logger.Info(ctx, "Connected to DB")
	return &Clients{
		Postgres: db,
	}, nil
}

func connectToDB(ctx context.Context, cfg *config.Config) (*sqlx.DB, error) {
	connStr := cfg.DatabaseConnectionString()
	db, err := sqlx.ConnectContext(ctx, "postgres", connStr)
	if err != nil {
		return nil, errors.Join(ErrDbConnection, err)
	}
	if err := db.PingContext(ctx); err != nil {
		return nil, errors.Join(ErrDbPing, err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}
