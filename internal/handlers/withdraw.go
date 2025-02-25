package handlers

import (
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/models/request"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/session"
	"go-musthave-diploma-tpl/internal/utils/moonChecker"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *HandlerProvider) WithdrawHandler(c *gin.Context) {
	ssp := session.SessionProvider{
		Config: p.Config,
	}

	userID, err := ssp.ParseToken(c)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	var body request.Withdraw
	bytes, err := c.GetRawData()
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}
	if err := json.Unmarshal(bytes, &body); err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusBadRequest, err)
		return
	}

	if body.Sum <= 0 {
		sendErrorResponse(c, p.Sugar, http.StatusBadRequest, errors.New("sum must be positive"))
	}

	if body.Order == "" || !moonChecker.MoonChecker(body.Order) {
		sendErrorResponse(c, p.Sugar, http.StatusUnprocessableEntity, err)
		return
	}

	err = p.Repository.BalanceWithdraw(userID, &body)
	if err != nil {
		var shouldPositiveErr *repository.ShouldBePositiveError
		if errors.As(err, &shouldPositiveErr) {
			sendErrorResponse(c, p.Sugar, http.StatusPaymentRequired, err)
			return
		}

		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
func (p *HandlerProvider) WithdrawalsHandler(c *gin.Context) {
	ssp := session.SessionProvider{
		Config: p.Config,
	}

	userID, err := ssp.ParseToken(c)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	withdraws, err := p.Repository.GetWithdraws(userID)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	if len(withdraws) == 0 {
		sendErrorResponse(c, p.Sugar, http.StatusNoContent, errors.New("no withdraws found"))
		return
	}

	c.JSON(http.StatusOK, withdraws)
}
