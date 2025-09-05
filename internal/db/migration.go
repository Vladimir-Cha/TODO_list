package db

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

func GooseMigrationsWithPool(ctx context.Context, pool *pgxpool.Pool) error {
	// преобразование pgxpool.Pool в *sql.DB
	db := stdlib.OpenDBFromPool(pool)
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.UpContext(ctx, db, "migrations"); err != nil {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
