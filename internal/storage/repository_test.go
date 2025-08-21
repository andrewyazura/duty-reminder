package storage

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
)

func setupTestDatabase(t *testing.T) (Querier, func()) {
	t.Helper()

	connection, err := pgx.Connect(context.Background(), os.Getenv("TEST_POSTGRES_URL"))
	if err != nil {
		t.Fatalf("couldn't setup the test database: %v", err)
	}

	transaction, err := connection.Begin(context.Background())
	if err != nil {
		t.Fatalf("couldn't start a transaction: %v", err)
	}

	teardownFunc := func() {
		err := transaction.Rollback(context.Background())

		if err != nil {
			t.Fatalf("failed to rollback: %v", err)
		}
	}

	return transaction, teardownFunc
}

func TestFindByID(t *testing.T) {
	querier, teardownFunc := setupTestDatabase(t)
	defer teardownFunc()

	repo := PostgresHouseholdRepository{db: querier}

	t.Run("", func(t *testing.T) {
		repo.FindByID(context.Background(), 1)
	})
}
