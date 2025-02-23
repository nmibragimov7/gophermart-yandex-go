package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/models/entity"
	"log"

	"github.com/jackc/pgerrcode"
	"github.com/lib/pq"
)

const (
	SelectOrder                          = "SELECT number, user_id FROM orders WHERE number = $1"
	SelectAllOrders                      = "SELECT number, accrual, status, uploaded_at FROM orders WHERE user_id = $1"
	InsertOrder                          = "INSERT INTO orders (number, user_id) VALUES ($1, $2)"
	UpdateOrder                          = "UPDATE orders SET status = $1, accrual = $2 WHERE number = $3"
	SelectOrdersByNewAndProcessingStatus = `SELECT number, user_id FROM orders WHERE status in ('NEW','PROCESSING') ORDER BY uploaded_at LIMIT $1`
)

func (p *RepositoryProvider) GetOrders(userID int64) ([]entity.Order, error) {
	var records []entity.Order

	rows, err := p.DB.Query(SelectAllOrders, userID)
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
		var record entity.Order

		err := rows.Scan(&record.Number, &record.Accrual, &record.Status, &record.UploadedAt)
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
func (p *RepositoryProvider) GetOrderWithUserID(number string) (*entity.OrderWithUserID, error) {
	var record entity.OrderWithUserID
	err := p.DB.QueryRow(SelectOrder, number).Scan(&record.Number, &record.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("order not exists: %w", NewNotFoundError(number))
		}

		return nil, fmt.Errorf("failed to get record from database: %w", err)
	}

	return &record, nil
}
func (p *RepositoryProvider) SaveOrder(data *entity.OrderWithUserID) error {
	_, err := p.DB.Exec(InsertOrder, data.Number, data.UserID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("order already exists: %w", NewDuplicateError(
				pgerrcode.UniqueViolation,
				err,
			))
		}
		return fmt.Errorf("failed to insert order: %w", err)
	}

	return nil
}
func (p *RepositoryProvider) GetNewOrders(limit int) ([]*entity.OrderWithUserID, error) {
	var records []*entity.OrderWithUserID

	rows, err := p.DB.Query(SelectOrdersByNewAndProcessingStatus, limit)
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
		var number string
		var userID int64

		err := rows.Scan(&number, &userID)
		if err != nil {
			return nil, fmt.Errorf("failed to scan record: %w", err)
		}
		records = append(records, &entity.OrderWithUserID{Number: number, UserID: userID})
	}
	return records, nil
}
func (p *RepositoryProvider) UpdateOrderBatches(data []*entity.AccrualWithUserID) error {
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

	query := "UPDATE orders SET status = CASE number "
	params := make([]interface{}, 0, len(data)*3)
	orderNumbers := make([]string, 0, len(data))

	for i, record := range data {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", i*3+1, i*3+2)
		params = append(params, record.Order, record.Status)
		orderNumbers = append(orderNumbers, fmt.Sprintf("$%d", i*3+3))
	}

	query += "END, accrual = CASE number "
	for i, record := range data {
		query += fmt.Sprintf("WHEN $%d THEN $%d ", i*3+1, i*3+2+len(data))
		params = append(params, record.Order, record.Accrual)
	}

	query += fmt.Sprintf("END WHERE number IN (%s)", stringJoin(orderNumbers, ", "))

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

func stringJoin(elements []string, sep string) string {
	result := ""
	for i, elem := range elements {
		if i > 0 {
			result += sep
		}
		result += elem
	}
	return result
}
