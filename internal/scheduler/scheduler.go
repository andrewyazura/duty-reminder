// Package scheduler
package scheduler

import (
	"context"
	"log/slog"

	"github.com/andrewyazura/duty-reminder/internal/domain"
	"github.com/andrewyazura/duty-reminder/internal/eventbus"
	"github.com/andrewyazura/duty-reminder/internal/services"
	"github.com/andrewyazura/duty-reminder/internal/storage"
	"github.com/go-co-op/gocron/v2"
)

type NotificationScheduler struct {
	eventBus      *eventbus.EventBus
	logger        *slog.Logger
	scheduler     gocron.Scheduler
	householdJobs map[int64]gocron.Job
}

func New(
	bus *eventbus.EventBus,
	logger *slog.Logger,
	uow services.UnitOfWork,
) (*NotificationScheduler, error) {
	s, err := gocron.NewScheduler()
	if err != nil {
		return nil, err
	}

	n := &NotificationScheduler{
		eventBus:      bus,
		logger:        logger,
		scheduler:     s,
		householdJobs: make(map[int64]gocron.Job),
	}

	err = n.registerJobs(uow)
	if err != nil {
		return nil, err
	}

	bus.Subscribe("HouseholdCreated", n.createHouseholdJob)
	bus.Subscribe("HouseholdCrontabUpdated", n.updateHouseholdJob)
	bus.Subscribe("HouseholdDeleted", n.deleteHouseholdJob)

	return n, nil
}

func (n *NotificationScheduler) Start() {
	n.scheduler.Start()
	n.logger.Info("scheduler started")
}

func (n *NotificationScheduler) Shutdown() {
	err := n.scheduler.Shutdown()

	if err != nil {
		n.logger.Error("failed to shutdown the scheduler", "error", err)
		return
	}

	n.logger.Info("scheduler shutdown")
}

func (n *NotificationScheduler) registerJobs(uow services.UnitOfWork) error {
	var households []*domain.Household

	err := uow.Execute(context.Background(), func(repo storage.HouseholdRepository) error {
		var err error
		households, err = repo.GetSchedules(context.Background())
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	for _, h := range households {
		job, err := n.createJob(h)

		if err != nil {
			return err
		}

		n.householdJobs[h.TelegramID] = job
	}

	return nil
}

func (n *NotificationScheduler) createHouseholdJob(ctx context.Context, event eventbus.Event) {
	h := event.(*domain.Household)

	job, err := n.createJob(h)
	if err != nil {
		n.logger.Error(
			"failed to register a new household job",
			"telegram_id", h.TelegramID,
			"error", err,
		)
	}

	n.logger.Info("created a job", "household", h.TelegramID)

	n.householdJobs[h.TelegramID] = job
}

func (n *NotificationScheduler) updateHouseholdJob(ctx context.Context, event eventbus.Event) {
	h := event.(*domain.Household)

	job, ok := n.householdJobs[h.TelegramID]
	if !ok {
		return
	}

	err := n.scheduler.RemoveJob(job.ID())
	if err != nil {
		n.logger.Error(
			"failed to remove old household job",
			"telegram_id", h.TelegramID,
			"error", err,
		)
	}

	newJob, err := n.createJob(h)
	if err != nil {
		n.logger.Error(
			"failed to register a new household job",
			"telegram_id", h.TelegramID,
			"error", err,
		)
	}

	n.logger.Info("updated a job", "household", h.TelegramID)

	n.householdJobs[h.TelegramID] = newJob
}

func (n *NotificationScheduler) createJob(h *domain.Household) (gocron.Job, error) {
	return n.scheduler.NewJob(
		gocron.CronJob(h.Crontab, false),
		gocron.NewTask(
			func(ctx context.Context, h *domain.Household) {
				n.eventBus.Publish(ctx, "NotifyHousehold", h)
			},
			h,
		),
	)
}

func (n *NotificationScheduler) deleteHouseholdJob(ctx context.Context, event eventbus.Event) {
	h := event.(*domain.Household)

	job, ok := n.householdJobs[h.TelegramID]
	if !ok {
		return
	}

	err := n.scheduler.RemoveJob(job.ID())
	if err != nil {
		n.logger.Error(
			"failed to remove old household job",
			"telegram_id", h.TelegramID,
			"error", err,
		)
	}

	delete(n.householdJobs, h.TelegramID)
	n.logger.Info("deleted a job", "household", h.TelegramID)
}
