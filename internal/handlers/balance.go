package handlers

import (
	"go-musthave-diploma-tpl/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *HandlerProvider) BalanceHandler(c *gin.Context) {
	ssp := session.SessionProvider{
		Config: p.Config,
	}

	userID, err := ssp.ParseToken(c)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	balance, err := p.Repository.GetBalance(userID)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, balance)
}
