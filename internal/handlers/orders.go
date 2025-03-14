package handlers

import (
	"errors"
	"go-musthave-diploma-tpl/internal/models/entity"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/session"
	"go-musthave-diploma-tpl/internal/utils/moonchecker"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *HandlerProvider) OrderSaveHandler(c *gin.Context) {
	ssp := session.SessionProvider{
		Config: p.Config,
	}

	userID, err := ssp.ParseToken(c)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusBadRequest, err)
		return
	}
	orderNumber := string(body)
	if orderNumber == "" || !moonchecker.MoonChecker(orderNumber) {
		sendErrorResponse(c, p.Sugar, http.StatusUnprocessableEntity, err)
		return
	}

	order, err := p.Repository.GetOrderWithUserID(orderNumber)
	if err != nil {
		var notFoundError *repository.NotFoundError
		if !errors.As(err, &notFoundError) {
			sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
			return
		}

		err = p.Repository.SaveOrder(&entity.OrderWithUserID{
			Number: orderNumber,
			UserID: userID,
		})
		if err != nil {
			var duplicateError *repository.DuplicateError
			if errors.As(err, &duplicateError) {
				sendErrorResponse(c, p.Sugar, http.StatusConflict, err)
				return
			}

			sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
			return
		}

		c.Status(http.StatusAccepted)
		return
	}

	if userID != order.UserID {
		sendErrorResponse(c, p.Sugar, http.StatusConflict, err)
		return
	}

	c.Status(http.StatusOK)
}
func (p *HandlerProvider) OrdersHandler(c *gin.Context) {
	ssp := session.SessionProvider{
		Config: p.Config,
	}

	userID, err := ssp.ParseToken(c)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	orders, err := p.Repository.GetOrders(userID)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	if len(orders) == 0 {
		sendErrorResponse(c, p.Sugar, http.StatusNoContent, err)
		return
	}

	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, orders)
}
