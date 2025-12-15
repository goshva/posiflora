package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool создаёт и возвращает пул соединений с PostgreSQL
func NewPool(ctx context.Context, dbURL string) (*pgxpool.Pool, error) {
	return pgxpool.New(ctx, dbURL)
}
