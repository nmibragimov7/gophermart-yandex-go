package response

import (
	"fmt"
	"go-musthave-diploma-tpl/internal/models/entity"
)

type Response struct {
	Message string `json:"message"`
}

type Accrual struct {
	Order   string             `json:"order"`
	Status  entity.OrderStatus `json:"status"`
	Accrual *float64           `json:"accrual,omitempty"`
}

type TooManyRequestsError struct {
	RetryAfter int
}

func (e *TooManyRequestsError) Error() string {
	return fmt.Sprintf("too many requests: %d", e.RetryAfter)
}
