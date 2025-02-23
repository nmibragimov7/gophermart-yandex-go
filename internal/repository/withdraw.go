package repository

import (
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/models/entity"
	"go-musthave-diploma-tpl/internal/models/request"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

const (
	InsertIntoPayments = `INSERT INTO withdrawals (sum, user_id, "order") VALUES ($1, $2, $3)`
	SelectWithdraws    = `SELECT sum, "order", processed_at FROM withdrawals WHERE user_id = $1`
)

func (p *RepositoryProvider) BalanceWithdraw(userID int64, withdraw *request.Withdraw) error {
	tx, err := p.DB.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("failed to rollback transaction: %s", err.Error())
		}
	}()

	_, err = tx.Exec(UpdateBalance, withdraw.Sum, userID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.CheckViolation {
			return fmt.Errorf("failed to decreace current balance: %w", NewShouldBePositiveError(
				pgerrcode.UniqueViolation,
				err,
			))
		}

		return fmt.Errorf("failed to update balance: %w", err)
	}
	_, err = tx.Exec(InsertIntoPayments, withdraw.Sum, userID, withdraw.Order)
	if err != nil {
		return fmt.Errorf("failed to insert payment: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
func (p *RepositoryProvider) GetWithdraws(userID int64) ([]entity.Withdraw, error) {
	var records []entity.Withdraw

	rows, err := p.DB.Query(SelectWithdraws, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query records: %w", err)
	}
	defer func() {
		err := rows.Close()
		if err != nil {
			log.Printf("failed to close rows: %s", err.Error())
		}
	}()

	for rows.Next() {
		var record entity.Withdraw

		err := rows.Scan(&record.Sum, &record.Order, &record.ProcessedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}

		records = append(records, record)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to scan records: %w", err)
	}

	return records, nil
}
