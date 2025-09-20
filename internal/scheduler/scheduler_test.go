package scheduler

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/storage"
)

type mockHouseholdRepo struct {
	households []*domain.Household
	err        error
}

func (repo *mockHouseholdRepo) Create(ctx context.Context, h *domain.Household) error { return nil }
func (repo *mockHouseholdRepo) Save(ctx context.Context, h *domain.Household) error   { return nil }
func (repo *mockHouseholdRepo) SaveWithMembers(ctx context.Context, h *domain.Household) error {
	return nil
}
func (repo *mockHouseholdRepo) FindByID(ctx context.Context, telegramID int64) (*domain.Household, error) {
	return nil, nil
}

func (repo *mockHouseholdRepo) GetSchedules(ctx context.Context) ([]*domain.Household, error) {
	return repo.households, repo.err
}

type mockUnitOfWork struct {
	repo *mockHouseholdRepo
}

func (m *mockUnitOfWork) Execute(ctx context.Context, fn func(repo storage.HouseholdRepository) error) error {
	return fn(m.repo)
}

func (m *mockUnitOfWork) ExecuteTransaction(ctx context.Context, fn func(repo storage.HouseholdRepository) error) error {
	return fn(m.repo)
}

func TestNew(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	t.Run("success", func(t *testing.T) {
		bus := eventbus.NewEventBus(logger)
		mockRepo := &mockHouseholdRepo{
			households: []*domain.Household{
				{TelegramID: 1, Crontab: "0 9 * * *"},
				{TelegramID: 2, Crontab: "0 10 * * *"},
			},
		}
		mockUOW := &mockUnitOfWork{repo: mockRepo}

		s, err := New(bus, logger, mockUOW)
		if err != nil {
			t.Fatalf("New() returned unexpected error: %v", err)
		}

		jobs := s.scheduler.Jobs()
		if len(jobs) != 2 {
			t.Errorf("got %d jobs in scheduler, want %d", len(jobs), 2)
		}

		if len(s.householdJobs) != 2 {
			t.Errorf("got %d entries in householdJobs, want %d", len(s.householdJobs), 2)
		}
	})

	t.Run("fail", func(t *testing.T) {
		bus := eventbus.NewEventBus(logger)

		want := errors.New("database dead")
		mockRepo := &mockHouseholdRepo{
			err: want,
		}
		mockUOW := &mockUnitOfWork{repo: mockRepo}

		_, err := New(bus, logger, mockUOW)

		if err == nil {
			t.Fatal("New() did not return an error")
		}

		if !errors.Is(err, want) {
			t.Errorf("got error %v, want %v", err, want)
		}
	})
}

func TestEvents(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	bus := eventbus.NewEventBus(logger)
	mockRepo := &mockHouseholdRepo{}
	mockUOW := &mockUnitOfWork{repo: mockRepo}

	s, err := New(bus, logger, mockUOW)
	if err != nil {
		t.Fatalf("failed to create scheduler: %v", err)
	}

	t.Run("HouseholdCreated", func(t *testing.T) {
		jobsBefore := len(s.scheduler.Jobs())
		h := &domain.Household{TelegramID: -1234567898765, Crontab: "0 9 * * *"}

		s.createHouseholdJob(context.Background(), h)

		if got := len(s.scheduler.Jobs()); got != jobsBefore+1 {
			t.Errorf("jobs before: %d, got: %d jobs, want: %d", jobsBefore, got, jobsBefore+1)
		}
	})

	t.Run("HouseholdUpdated", func(t *testing.T) {
		h := &domain.Household{TelegramID: -1234567898765, Crontab: "0 9 * * *"}
		s.createHouseholdJob(context.Background(), h)
		initialJob, ok := s.householdJobs[h.TelegramID]
		if !ok {
			t.Fatal("initial job not created")
		}

		h.Crontab = "0 2 * * *"
		s.updateHouseholdJob(context.Background(), h)

		updatedJob, ok := s.householdJobs[h.TelegramID]
		if !ok {
			t.Fatal("job was removed instead of updated")
		}

		if updatedJob.ID() == initialJob.ID() {
			t.Error("job was not updated, ID remained the same")
		}
	})

	t.Run("HouseholdDeleted", func(t *testing.T) {
		h := &domain.Household{TelegramID: -1234567898765, Crontab: "0 9 * * *"}
		s.createHouseholdJob(context.Background(), h)
		jobsBefore := len(s.scheduler.Jobs())

		if _, ok := s.householdJobs[h.TelegramID]; !ok {
			t.Fatal("initial job not created")
		}

		s.deleteHouseholdJob(context.Background(), h)

		if got := len(s.scheduler.Jobs()); got != jobsBefore-1 {
			t.Errorf("jobs before: %d, got: %d jobs, want: %d", jobsBefore, got, jobsBefore+1)
		}

		if _, ok := s.householdJobs[h.TelegramID]; ok {
			t.Errorf("job was not removed")
		}
	})
}
