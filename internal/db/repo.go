package db

import (
	"context"

	"github.com/Vladimir-Cha/TODO_list/internal/config"
	"github.com/Vladimir-Cha/TODO_list/internal/errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/gommon/log"
)

// New создает и возвращает *pgxpool.Pool напрямую
func New(ctx context.Context, cfg config.AppConfig) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.Database.DBUrl)
	if err != nil {
		return nil, errors.ErrBadRequest
	}

	poolConfig.MaxConns = cfg.Database.DBMaxConns
	poolConfig.MinConns = cfg.Database.DBMinConns
	poolConfig.MaxConnLifetime = cfg.Database.DBConnLife

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, errors.ErrBadRequest
	}

	// Проверка подключения
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, errors.ErrBadRequest
	}

	log.Printf("PostgreSQL connected (Max: %d, Min: %d)", cfg.Database.DBMaxConns, cfg.Database.DBMinConns)
	return pool, nil
}

func InitNew(ctx context.Context, dbCfg config.AppConfig) (*pgxpool.Pool, error) {
	return New(ctx, dbCfg)
}
