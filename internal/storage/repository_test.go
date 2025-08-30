package storage

import (
	"context"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/testutils"
)

var tmpDB *testutils.TempPostgresInstance

func TestMain(m *testing.M) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	tmpDB = testutils.NewTempPostgresInstance(logger, "test-db", "test-user", 5434)

	err := tmpDB.Setup()
	if err != nil {
		logger.Error("failed to setup database", "error", err)
	}

	defer tmpDB.Cleanup()

	m.Run()
}

func setupTestDatabase(t *testing.T) (Querier, func()) {
	t.Helper()

	transaction, err := tmpDB.Connection.Begin(context.Background())
	if err != nil {
		t.Fatalf("failed to start a transaction: %v", err)
	}

	if err := applySchema(transaction); err != nil {
		t.Fatalf("failed to apply sql schema: %v", err)
	}

	teardownFunc := func() {
		err := transaction.Rollback(context.Background())

		if err != nil {
			t.Fatalf("failed to rollback: %v", err)
		}
	}

	return transaction, teardownFunc
}

func applySchema(db Querier) error {
	data, err := os.ReadFile("../../sql/schema.sql")
	if err != nil {
		return err
	}

	if _, err := db.Exec(context.Background(), string(data)); err != nil {
		return err
	}

	return nil
}

func TestFindByID(t *testing.T) {
	querier, teardownFunc := setupTestDatabase(t)
	defer teardownFunc()

	repo := PostgresHouseholdRepository{db: querier}

	t.Run("success", func(t *testing.T) {
		want := domain.NewHousehold()
		want.TelegramID = 0
		want.Checklist = append(want.Checklist, "point 1")

		_, err := querier.Exec(context.Background(), `
			INSERT INTO households (
				telegram_id,
				checklist,
				crontab,
				current_member_index
			) VALUES ($1, $2, $3, $4)`,
			want.TelegramID,
			want.Checklist,
			want.Crontab,
			want.CurrentMember,
		)

		if err != nil {
			t.Fatalf("failed to insert test household into database: %v", err)
		}

		got, err := repo.FindByID(context.Background(), want.TelegramID)

		if err != nil {
			t.Fatalf("FindByID returned an error: %v", err)
		}

		if got == nil {
			t.Fatalf("expected household with telegram_id %d to exist", 1)
		}

		if got.Crontab != want.Crontab {
			t.Errorf("crontab is %s, want %s", got.Crontab, want.Crontab)
		}

		if got.CurrentMember != want.CurrentMember {
			t.Errorf("current member index is %d, want %d", got.CurrentMember, want.CurrentMember)
		}

		if !reflect.DeepEqual(got.Checklist, want.Checklist) {
			t.Errorf("checklist is %v, want %v", got.Checklist, want.Checklist)
		}
	})

	t.Run("success with members", func(t *testing.T) {
		want := domain.NewHousehold()
		want.TelegramID = 1
		want.Checklist = append(want.Checklist, "point 1")

		want.AddMember(&domain.Member{Name: "test1", TelegramID: 1})
		want.AddMember(&domain.Member{Name: "test2", TelegramID: 2})
		want.AddMember(&domain.Member{Name: "test3", TelegramID: 3})

		_, err := querier.Exec(context.Background(), `
			INSERT INTO households (
				telegram_id,
				checklist,
				crontab,
				current_member_index
			) VALUES ($1, $2, $3, $4)`,
			want.TelegramID,
			want.Checklist,
			want.Crontab,
			want.CurrentMember,
		)

		if err != nil {
			t.Fatalf("failed to insert test household into database: %v", err)
		}

		for _, m := range want.Members {
			_, err = querier.Exec(context.Background(), `
				INSERT INTO members (
					household_telegram_id,
					name,
					"order",
					telegram_id
				) VALUES ($1, $2, $3, $4)`,
				want.TelegramID,
				m.Name,
				m.Order,
				m.TelegramID,
			)

			if err != nil {
				t.Fatalf("failed to insert test member into database: %v", err)
			}

			// FindByID does not select "order" column,
			// therefore it is expected to be 0 in the entity
			// and we only need it for order in DB
			m.Order = 0
		}

		got, err := repo.FindByID(context.Background(), want.TelegramID)

		if err != nil {
			t.Fatalf("FindByID returned an error: %v", err)
		}

		if got == nil {
			t.Fatalf("expected household with telegram_id %d to exist", 1)
		}

		if !reflect.DeepEqual(got.Members, want.Members) {
			t.Errorf("members list is %v, want %v", got.Members, want.Members)
		}
	})
}
