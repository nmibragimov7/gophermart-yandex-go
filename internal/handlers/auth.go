package handlers

import (
	"encoding/json"
	"errors"
	"go-musthave-diploma-tpl/internal/models/request"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
)

func (p *HandlerProvider) RegisterHandler(c *gin.Context) {
	var body request.Register
	bytes, err := c.GetRawData()
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}
	if err := json.Unmarshal(bytes, &body); err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusBadRequest, err)
		return
	}

	err = p.Repository.SaveUser(&body)
	if err != nil {
		var duplicateError *repository.DuplicateError
		if errors.As(err, &duplicateError) {
			sendErrorResponse(c, p.Sugar, http.StatusConflict, err)
			return
		}

		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	c.Status(http.StatusOK)
}
func (p *HandlerProvider) LoginHandler(c *gin.Context) {
	var body request.Login
	bytes, err := c.GetRawData()
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}
	if err := json.Unmarshal(bytes, &body); err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusBadRequest, err)
		return
	}

	user, err := p.Repository.GetUser(&body)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	ssp := &session.SessionProvider{
		Config: p.Config,
	}

	if !ssp.ComparePasswords(body.Password, user.Password) {
		sendErrorResponse(c, p.Sugar, http.StatusUnauthorized, err)
		return
	}

	token, err := ssp.CreateToken(user.ID)
	if err != nil {
		sendErrorResponse(c, p.Sugar, http.StatusInternalServerError, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}
