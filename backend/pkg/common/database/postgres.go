package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBOptions struct {
	MaxRetries     int
	RetryInterval  time.Duration
	MaxConnections int32
}

func NewPostgresPool(ctx context.Context, dsn string, opts DBOptions) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	config.MaxConns = opts.MaxConnections

	var pool *pgxpool.Pool
	for i := 1; i <= opts.MaxRetries; i++ {
		log.Printf("Connecting to Postgres (attempt %d/%d)...", i, opts.MaxRetries)

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			err = pool.Ping(ctx)
			if err == nil {
				log.Println("Sucessfuly connected to Postgres")
				return pool, nil
			}
		}

		if i < opts.MaxRetries {
			log.Printf("Postgres not ready: %v. Retrying in %v...", err, opts.RetryInterval)

			select {
			case <-time.After(opts.RetryInterval):
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during db connection: %w", ctx.Err())
			}
		}
	}

	return nil, fmt.Errorf("could not connect to postgres after %d attempts: %w", opts.MaxRetries, err)
}
