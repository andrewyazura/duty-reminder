// Package storage
package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/andrewyazura/duty-reminder/internal/domain"
)

type HouseholdRepository interface {
	Create(ctx context.Context, h *domain.Household) error
	Save(ctx context.Context, h *domain.Household) error
	SaveWithMembers(ctx context.Context, h *domain.Household) error
	FindByID(ctx context.Context, telegramID int) (*domain.Household, error)
	GetSchedules(ctx context.Context) ([]*domain.Household, error)
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row

	CopyFrom(ctx context.Context, tableName pgx.Identifier, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type PostgresHouseholdRepository struct {
	db Querier
}

func NewPostgresHouseholdRepository(querier Querier) *PostgresHouseholdRepository {
	return &PostgresHouseholdRepository{db: querier}
}

func (repo PostgresHouseholdRepository) Create(ctx context.Context, h *domain.Household) error {
	insertHouseholdQuery := `
		INSERT INTO households (
			telegram_id,
			checklist,
			crontab,
			current_member_index
		) VALUES ($1, $2, $3, $4)
	`

	_, err := repo.db.Exec(
		ctx,
		insertHouseholdQuery,
		h.TelegramID,
		h.Checklist,
		h.Crontab,
		h.CurrentMember,
	)

	if err != nil {
		return err
	}

	return nil
}

func (repo PostgresHouseholdRepository) Save(ctx context.Context, h *domain.Household) error {
	updateHouseholdQuery := `
		UPDATE households
		SET checklist = $1, crontab = $2, current_member_index = $3
		WHERE telegram_id = $4
	`

	_, err := repo.db.Exec(
		ctx,
		updateHouseholdQuery,
		h.Checklist,
		h.Crontab,
		h.CurrentMember,
		h.TelegramID,
	)

	if err != nil {
		return err
	}

	return nil
}

func (repo PostgresHouseholdRepository) SaveWithMembers(ctx context.Context, h *domain.Household) error {
	err := repo.Save(ctx, h)

	if err != nil {
		return err
	}

	deleteMembersQuery := `
		DELETE FROM members WHERE household_telegram_id = $1
	`

	if _, err := repo.db.Exec(ctx, deleteMembersQuery, h.TelegramID); err != nil {
		return err
	}

	if len(h.Members) == 0 {
		return nil
	}

	rows := make([][]any, len(h.Members))
	for i, m := range h.Members {
		rows[i] = []any{
			h.TelegramID,
			m.TelegramID,
			m.Name,
			m.Order,
		}
	}

	if _, err := repo.db.CopyFrom(
		ctx,
		pgx.Identifier{"members"},
		[]string{
			"household_telegram_id",
			"telegram_id",
			"name",
			"order",
		},
		pgx.CopyFromRows(rows),
	); err != nil {
		return err
	}

	return nil
}

func (repo PostgresHouseholdRepository) FindByID(ctx context.Context, telegramID int) (*domain.Household, error) {
	householdQuery := `
		SELECT 
			checklist,
			crontab,
			current_member_index
		FROM households
		WHERE telegram_id = $1
	`

	h := &domain.Household{TelegramID: telegramID}

	row := repo.db.QueryRow(ctx, householdQuery, telegramID)
	err := row.Scan(&h.Checklist, &h.Crontab, &h.CurrentMember)

	if err != nil {
		return nil, err
	}

	membersQuery := `
		SELECT
			telegram_id,
			name,
			"order"
		FROM members
		WHERE
			household_telegram_id = $1
		ORDER BY "order" ASC
	`

	rows, err := repo.db.Query(ctx, membersQuery, telegramID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	h.Members = []*domain.Member{}
	for rows.Next() {
		member := &domain.Member{}
		if err := rows.Scan(&member.TelegramID, &member.Name, &member.Order); err != nil {
			return nil, err
		}

		h.Members = append(h.Members, member)
	}

	return h, nil
}

func (repo PostgresHouseholdRepository) GetSchedules(ctx context.Context) ([]*domain.Household, error) {
	householdsQuery := `
		SELECT
			telegram_id,
			checklist,
			crontab
		FROM households
	`

	rows, err := repo.db.Query(ctx, householdsQuery)
	if err != nil {
		return nil, err
	}

	var households []*domain.Household
	for rows.Next() {
		h := &domain.Household{}
		err := rows.Scan(&h.TelegramID, &h.Checklist, &h.Crontab)

		if err != nil {
			return nil, err
		}

		households = append(households, h)
	}

	if err := rows.Err(); err != nil {
		return households, err
	}

	return households, nil
}
