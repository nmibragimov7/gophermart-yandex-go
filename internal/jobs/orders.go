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
	Config     *config.Config
	Sugar      *zap.SugaredLogger
	Repository *repository.RepositoryProvider
}

const (
	jobsCount = 4
)

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
		errCh := make(chan error, 1)
		responsesCh := p.fanOut(once, doneCh, errCh, orders)
		resultsCh := p.fanIn(doneCh, responsesCh...)

		result := make([]*entity.AccrualWithUserID, 0)

		select {
		case err := <-errCh:
			once.Do(func() { close(doneCh) })

			var tooManyRequests *response.TooManyRequestsError
			if errors.As(err, &tooManyRequests) {
				interval = time.Duration(tooManyRequests.RetryAfter) * time.Second
			}

			err = p.Repository.UpdateOrderBatches(result)
			if err != nil {
				p.Sugar.Errorw("Failed to update orders",
					"error", err,
				)
			}

			grouped := groupOrders(result)
			err = p.Repository.UpdateBalanceBatches(grouped)
			if err != nil {
				p.Sugar.Errorw("Failed to update balances",
					"error", err,
				)
			}

			continue
		default:
			for data := range resultsCh {
				result = append(result, data)
			}
		}

		err = p.Repository.UpdateOrderBatches(result)
		if err != nil {
			p.Sugar.Errorw("Failed to update orders",
				"error", err,
			)
		}

		grouped := groupOrders(result)
		err = p.Repository.UpdateBalanceBatches(grouped)
		if err != nil {
			p.Sugar.Errorw("Failed to update balances",
				"error", err,
			)
		}

		once.Do(func() { close(doneCh) })

		if !timer.Stop() {
			<-timer.C
		}
		timer.Reset(interval)
	}
}
func (p *JobProvider) fanOut(once *sync.Once, doneCh chan struct{}, errorCh chan error, orders []*entity.OrderWithUserID) []chan *entity.AccrualWithUserID {
	channels := make([]chan *entity.AccrualWithUserID, len(orders))

	for i, order := range orders {
		channels[i] = p.sendRequest(once, doneCh, errorCh, order)
	}

	return channels
}
func (p *JobProvider) fanIn(doneCh chan struct{}, responsesCh ...chan *entity.AccrualWithUserID) chan *entity.AccrualWithUserID {
	channel := make(chan *entity.AccrualWithUserID)
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
				case channel <- data:
				}
			}
		}(ch)
	}
	go func() {
		wg.Wait()
		close(channel)
	}()

	return channel
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
