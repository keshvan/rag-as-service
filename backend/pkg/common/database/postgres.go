package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DBOptions struct {
	MaxRetries     int
	RetryInterval  time.Duration
	MaxConnections int32
}

func NewPostgresPool(ctx context.Context, dsn string, opts DBOptions, log *slog.Logger) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse url: %w", err)
	}

	config.MaxConns = opts.MaxConnections

	var pool *pgxpool.Pool
	for i := 1; i <= opts.MaxRetries; i++ {
		log.Info("connecting to postgres", "attempt", i)

		pool, err = pgxpool.NewWithConfig(ctx, config)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				log.Info("postgres connected")
				return pool, nil
			}
		}

		if pool != nil {
			pool.Close()
		}

		if i < opts.MaxRetries {
			log.Warn("postgres not ready",
				"err", err,
				"retry_in", opts.RetryInterval,
			)

			select {
			case <-time.After(opts.RetryInterval):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}

	return nil, fmt.Errorf("could not connect to postgres after %d attempts: %w", opts.MaxRetries, err)
}
