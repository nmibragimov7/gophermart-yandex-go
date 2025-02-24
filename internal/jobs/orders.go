package jobs

import (
	"errors"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/models/entity"
	"go-musthave-diploma-tpl/internal/models/response"
	"go-musthave-diploma-tpl/internal/repository"
	"sync"
	"time"

	"go.uber.org/zap"
)

type JobProvider struct {
	Sugar      *zap.SugaredLogger
	Config     *config.Config
	Channel    chan *entity.AccrualWithUserID
	Repository *repository.RepositoryProvider
}

const (
	jobsCount = 4
)

func (p *JobProvider) Flush() {
	ticker := time.NewTicker(5 * time.Second)

	var updates []*entity.AccrualWithUserID

	for {
		select {
		case update := <-p.Channel:
			updates = append(updates, update)
		case <-ticker.C:
			if len(updates) == 0 {
				continue
			}

			err := p.Repository.UpdateOrderBatches(updates)
			if err != nil {
				p.Sugar.Errorw("Failed to update orders",
					"error", err,
				)
			}

			grouped := groupOrders(updates)
			err = p.Repository.UpdateBalanceBatches(grouped)
			if err != nil {
				p.Sugar.Errorw("Failed to update balances",
					"error", err,
				)
			}

			updates = nil
		}
	}
}

func (p *JobProvider) Run(initialInterval time.Duration) {
	interval := initialInterval
	timer := time.NewTimer(interval)
	defer timer.Stop()

	for {
		orders, err := p.Repository.GetNewOrders(jobsCount)
		if err != nil {
			p.Sugar.Errorw("Failed to get new orders",
				"error", err,
			)
			<-timer.C
			timer.Reset(interval)
			continue
		}
		if len(orders) == 0 {
			p.Sugar.Infow("No new orders")
			<-timer.C
			timer.Reset(interval)
			continue
		}

		var once *sync.Once
		doneCh := make(chan struct{})
		errorCh := make(chan error, len(orders))
		responsesCh := p.fanOut(once, doneCh, errorCh, orders)
		p.fanIn(doneCh, responsesCh...)

		select {
		case <-timer.C:
			interval = initialInterval
			timer.Reset(interval)
		case err := <-errorCh:
			var tooManyRequests *response.TooManyRequestsError
			if errors.As(err, &tooManyRequests) {
				interval = time.Duration(tooManyRequests.RetryAfter) * time.Second
			}

			timer.Reset(interval)
		}

		//once.Do(func() { close(doneCh) })
		//
		//if !timer.Stop() {
		//	<-timer.C
		//}
		//interval = initialInterval
		//timer.Reset(interval)
	}
}
func (p *JobProvider) fanOut(once *sync.Once, doneCh chan struct{}, errorCh chan error, orders []*entity.OrderWithUserID) []chan *entity.AccrualWithUserID {
	channels := make([]chan *entity.AccrualWithUserID, len(orders))

	for i, order := range orders {
		channels[i] = p.sendRequest(once, doneCh, errorCh, order)
	}

	return channels
}
func (p *JobProvider) fanIn(doneCh chan struct{}, responsesCh ...chan *entity.AccrualWithUserID) {
	var wg sync.WaitGroup

	for _, ch := range responsesCh {
		closureCh := ch
		wg.Add(1)

		go func(ch chan *entity.AccrualWithUserID) {
			defer wg.Done()

			for data := range closureCh {
				select {
				case <-doneCh:
					return
				case p.Channel <- data:
				}
			}
		}(ch)
	}
	go func() {
		wg.Wait()
	}()
}
func (p *JobProvider) sendRequest(once *sync.Once, doneCh chan struct{}, errorCh chan error, order *entity.OrderWithUserID) chan *entity.AccrualWithUserID {
	channel := make(chan *entity.AccrualWithUserID)

	go func() {
		defer close(channel)

		ord := order
		select {
		case <-doneCh:
			return
		default:
			res, err := p.getOrderStatus(ord.Number) // запрос на другой сервис по http
			if err != nil {
				errorCh <- err
				once.Do(func() { close(doneCh) })
				return
			}
			channel <- &entity.AccrualWithUserID{
				UserID:  ord.UserID,
				Order:   res.Order,
				Status:  res.Status,
				Accrual: res.Accrual,
			}
		}
	}()

	return channel
}
func groupOrders(data []*entity.AccrualWithUserID) map[int64]float64 {
	grouped := make(map[int64]float64)
	for _, item := range data {
		if item.Accrual == nil {
			continue
		}
		grouped[item.UserID] += *item.Accrual
	}

	return grouped
}
