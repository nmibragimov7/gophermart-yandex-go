package repository

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/models/entity"
	"log"
)

const (
	CreateBalance     = "INSERT INTO balance (user_id) VALUES ($1)"
	SelectUserBalance = "SELECT current, withdrawn FROM balance WHERE user_id = $1"
	UpdateBalance     = `UPDATE balance SET current = current - $1, withdrawn = withdrawn + $1 WHERE user_id = $2`
)

func (p *RepositoryProvider) GetBalance(userID int64) (*entity.Balance, error) {
	var record entity.Balance
	err := p.DB.QueryRow(SelectUserBalance, userID).Scan(&record.Current, &record.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to get record from database: %w", err)
	}

	return &record, nil
}
func (p *RepositoryProvider) UpdateBalanceBatches(record map[int64]float64) error {
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

	query := "UPDATE balance SET current = CASE"
	ids := make([]interface{}, 0, len(record))
	params := make([]interface{}, 0, len(record)*2)
	i := 1
	for userID, accrual := range record {
		query += fmt.Sprintf(" WHEN user_id = $%d THEN current + $%d", i, i+1)
		params = append(params, userID, accrual)
		ids = append(ids, userID)
		i += 2
	}
	query += " END WHERE user_id IN ("
	for j := range ids {
		query += fmt.Sprintf("$%d,", i+j)
		params = append(params, ids[j])
	}
	query = query[:len(query)-1] + ")"

	_, err = tx.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("failed to update records in database: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
