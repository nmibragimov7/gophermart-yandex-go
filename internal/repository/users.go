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
	SelectUser = "SELECT id, password FROM users WHERE login = $1"
	InsertUser = "INSERT INTO users (login, password) VALUES ($1, $2) RETURNING id"
)

func (p *RepositoryProvider) GetUser(values *request.Login) (*entity.User, error) {
	var record entity.User
	err := p.DB.QueryRow(SelectUser, values.Login).Scan(&record.ID, &record.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to get record from database: %w", err)
	}

	return &record, nil
}
func (p *RepositoryProvider) SaveUser(values *request.Register) (int64, error) {
	tx, err := p.DB.Begin()
	if err != nil {
		return 0, fmt.Errorf("failed to start transaction: %w", err)
	}

	defer func() {
		err := tx.Rollback()
		if err != nil {
			log.Printf("failed to rollback transaction: %s", err.Error())
		}
	}()

	var userID int64
	err = tx.QueryRow(InsertUser, values.Login, values.Password).Scan(&userID)
	if err != nil {
		var pgErr *pq.Error
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return 0, fmt.Errorf("user already exists: %w", NewDuplicateError(
				pgerrcode.UniqueViolation,
				err,
			))
		}
		return 0, fmt.Errorf("failed to insert user: %w", err)
	}

	_, err = tx.Exec(CreateBalance, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to create balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return userID, nil
}
