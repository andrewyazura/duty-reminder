// Package storage
package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/andrewyazura/duty-reminder/internal/domain"
)

type HouseholdRepository interface {
	Create(h *domain.Household) error
	FindByID(id int) (*domain.Household, error)
}

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (commandTag pgconn.CommandTag, err error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

type PostgresHouseholdRepository struct {
	db Querier
}

func (repo PostgresHouseholdRepository) Create(h *domain.Household) error {
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
			name
		FROM members
		WHERE
			household_telegram_id = $1
	`

	rows, err := repo.db.Query(ctx, membersQuery, telegramID)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	h.Members = []*domain.Member{}
	for rows.Next() {
		member := &domain.Member{}
		if err := rows.Scan(&member.TelegramID, &member.Name); err != nil {
			return nil, err
		}

		h.Members = append(h.Members, member)
	}

	return h, nil
}
