// Package storage
package storage

import (
	"context"
	"database/sql"

	"github.com/andrewyazura/duty-reminder/internal/domain"
)

type HouseholdRepository interface {
	Create(h *domain.Household) error
	FindByID(id int) (*domain.Household, error)
}

type PostgresHouseholdRepository struct{ db *sql.DB }

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

	row := repo.db.QueryRowContext(ctx, householdQuery, telegramID)
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

	rows, err := repo.db.QueryContext(ctx, membersQuery, telegramID)

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
