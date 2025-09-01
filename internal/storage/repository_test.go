package storage

import (
	"context"
	"io"
	"log/slog"
	"os"
	"reflect"
	"testing"

	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/testutils"
)

var tmpDB *testutils.TempPostgresInstance

func TestMain(m *testing.M) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

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

	ctx := context.Background()
	repo := PostgresHouseholdRepository{db: querier}

	t.Run("success", func(t *testing.T) {
		want := domain.NewHousehold()
		want.TelegramID = 0
		want.Checklist = append(want.Checklist, "point 1")

		_, err := querier.Exec(ctx, `
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

		got, err := repo.FindByID(ctx, want.TelegramID)

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

		want.AddMember(&domain.Member{Name: "test1", TelegramID: 1, Order: 1})
		want.AddMember(&domain.Member{Name: "test2", TelegramID: 2, Order: 2})
		want.AddMember(&domain.Member{Name: "test3", TelegramID: 3, Order: 3})

		_, err := querier.Exec(ctx, `
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
			_, err = querier.Exec(ctx, `
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
		}

		got, err := repo.FindByID(ctx, want.TelegramID)

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

func TestSaveWithMembers(t *testing.T) {
	querier, teardownFunc := setupTestDatabase(t)
	defer teardownFunc()

	ctx := context.Background()
	repo := PostgresHouseholdRepository{db: querier}

	t.Run("success", func(t *testing.T) {
		want := domain.NewHousehold()
		want.TelegramID = 1
		want.Checklist = append(want.Checklist, "point 1")

		want.AddMember(&domain.Member{Name: "test1", TelegramID: 1, Order: 1})

		_, err := querier.Exec(ctx, `
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
			_, err = querier.Exec(ctx, `
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
		}

		want.RemoveMember(1)
		want.AddMember(&domain.Member{Name: "test2", TelegramID: 2, Order: 2})
		want.AddMember(&domain.Member{Name: "test3", TelegramID: 3, Order: 3})

		err = repo.SaveWithMembers(ctx, &want)
		if err != nil {
			t.Fatalf("SaveWithMembers() failed: %v", err)
		}

		got, err := repo.FindByID(ctx, want.TelegramID)
		if err != nil {
			t.Fatalf("FindByID() failed: %v", err)
		}

		if len(got.Members) != 2 {
			t.Fatalf("got %d members, want %d", len(got.Members), 2)
		}

		if got.Members[0].Name != "test2" || got.Members[1].Name != "test3" {
			t.Errorf("members are not correct or not in the correct order")
		}
	})
}

func TestGetSchedules(t *testing.T) {
	querier, teardownFunc := setupTestDatabase(t)
	defer teardownFunc()

	ctx := context.Background()
	repo := PostgresHouseholdRepository{db: querier}

	t.Run("success", func(t *testing.T) {
		h1 := domain.NewHousehold()
		h1.TelegramID = 1

		h2 := domain.NewHousehold()
		h2.TelegramID = 2

		households := []*domain.Household{&h1, &h2}

		for _, h := range households {
			_, err := querier.Exec(ctx, `
			INSERT INTO households (
				telegram_id,
				checklist,
				crontab,
				current_member_index
			) VALUES ($1, $2, $3, $4)`,
				h.TelegramID,
				h.Checklist,
				h.Crontab,
				h.CurrentMember,
			)

			if err != nil {
				t.Fatalf("failed to insert test household into database: %v", err)
			}
		}

		households, err := repo.GetSchedules(ctx)
		if err != nil {
			t.Fatalf("GetSchedules() failed: %v", err)
		}

		if len(households) != 2 {
			t.Fatalf("got %d households, want %d", len(households), 2)
		}

		if households[0].Crontab != h1.Crontab {
			t.Errorf("got crontab string %s, want %s", households[0].Crontab, h1.Crontab)
		}

		if households[1].Crontab != h2.Crontab {
			t.Errorf("got crontab string %s, want %s", households[1].Crontab, h2.Crontab)
		}
	})
}
