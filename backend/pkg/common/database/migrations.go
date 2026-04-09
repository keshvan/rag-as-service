package database

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type MigrateOptions struct {
	ConnectTimeout time.Duration
}

func (o *MigrateOptions) setDefaults() {
	if o.ConnectTimeout <= 0 {
		o.ConnectTimeout = 5 * time.Second
	}
}

func RunMigrations(ctx context.Context, dsn, migrationsDir string, opts MigrateOptions) error {
	opts.setDefaults()

	absDir, err := filepath.Abs(migrationsDir)
	if err != nil {
		return fmt.Errorf("resolve migrations dir: %w", err)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("open sql connection: %w", err)
	}
	defer func() { _ = db.Close() }()

	pingCtx, cancel := context.WithTimeout(ctx, opts.ConnectTimeout)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("init postgres migrate driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+filepath.ToSlash(absDir),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("init migrate instance: %w", err)
	}
	defer func() {
		_, _ = m.Close()
	}()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			return nil
		}
		return fmt.Errorf("apply up migrations: %w", err)
	}

	return nil
}
