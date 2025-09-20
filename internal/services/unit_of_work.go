package services

import (
	"context"

	"github.com/andrewyazura/duty-reminder/internal/storage"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UnitOfWork interface {
	Execute(ctx context.Context, fn func(repo storage.HouseholdRepository) error) error
	ExecuteTransaction(ctx context.Context, fn func(repo storage.HouseholdRepository) error) error
}

type PostgresUnitOfWork struct {
	pool *pgxpool.Pool
}

func NewPostgresUnitOfWork(pool *pgxpool.Pool) *PostgresUnitOfWork {
	return &PostgresUnitOfWork{pool: pool}
}

func (uow PostgresUnitOfWork) Execute(ctx context.Context, fn func(storage.HouseholdRepository) error) error {
	conn, err := uow.pool.Acquire(ctx)
	if err != nil {
		return err
	}
	defer conn.Release()

	repo := storage.NewPostgresHouseholdRepository(conn)

	if err := fn(repo); err != nil {
		return err
	}

	return nil
}

func (uow PostgresUnitOfWork) ExecuteTransaction(ctx context.Context, fn func(storage.HouseholdRepository) error) error {
	transaction, err := uow.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer transaction.Rollback(ctx)

	repo := storage.NewPostgresHouseholdRepository(transaction)

	if err := fn(repo); err != nil {
		return err
	}

	if err := transaction.Commit(ctx); err != nil {
		return err
	}

	return nil
}
