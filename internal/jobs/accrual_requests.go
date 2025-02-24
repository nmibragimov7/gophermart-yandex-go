package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go-musthave-diploma-tpl/internal/models/response"
	"net/http"
	"strconv"

	"github.com/go-resty/resty/v2"
)

func (p *JobProvider) getOrderStatus(ctx context.Context, order string) (*response.Accrual, error) {
	client := resty.New()
	fmt.Println("order", order)
	uri := *p.Config.Accrual + "/api/orders/" + order
	resp, err := client.R().SetContext(ctx).Get(uri)
	fmt.Println("resp", resp)
	fmt.Println("err", err)
	if err != nil {
		return nil, fmt.Errorf("failed to get order status: %w", err)
	}

	switch resp.StatusCode() {
	case http.StatusOK:
		fmt.Println("StatusOK")
		var res response.Accrual
		if err := json.Unmarshal(resp.Body(), &res); err != nil {
			return nil, fmt.Errorf("failed to unmarshal accrual response: %w", err)
		}
		return &res, nil
	case http.StatusTooManyRequests:
		fmt.Println("StatusTooManyRequests")
		retryAfter := resp.Header().Get("Retry-After")
		if retryAfter == "" {
			return nil, errors.New("no retry-after header")
		}

		if seconds, err := strconv.Atoi(retryAfter); err == nil {
			return nil, fmt.Errorf("retry after %d seconds", response.TooManyRequestsError{RetryAfter: seconds})
		}

		return nil, errors.New("invalid retry-after header")
	default:
		fmt.Println("default", resp.Status())
		return nil, errors.New("unexpected status code" + resp.Status())
	}
}
