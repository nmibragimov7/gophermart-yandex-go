package entity

import (
	"time"
)

const (
	OrderStatusNew        string = "NEW"
	OrderStatusProcessed  string = "PROCESSED"
	OderStatusInvalid     string = "INVALID"
	OrderStatusProcessing string = "PROCESSING"
)

type User struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Password  string `json:"password"`
	CreatedAt string `json:"created_at"`
}

type OrderWithUserID struct {
	Number string `json:"number"`
	UserID int64  `json:"-"`
}

type Order struct {
	OrderWithUserID
	Accrual    float64   `json:"accrual,omitempty"`
	Status     string    `json:"status"`
	UploadedAt time.Time `json:"uploaded_at"`
}

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type Withdraw struct {
	Order       string    `json:"order"`
	Sum         int       `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type AccrualWithUserID struct {
	UserID  int64    `json:"user_id"`
	Order   string   `json:"order"`
	Status  string   `json:"status"`
	Accrual *float64 `json:"accrual,omitempty"`
}
