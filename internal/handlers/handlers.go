package handlers

import (
	"encoding/json"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/models/response"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/session"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type HandlerProvider struct {
	Sugar      *zap.SugaredLogger
	Config     *config.Config
	Session    *session.SessionProvider
	Repository *repository.RepositoryProvider
}

const (
	logKeyURI       = "uri"
	logKeyIP        = "ip"
	contentType     = "Content-Type"
	contentLength   = "Content-Length"
	applicationJSON = "application/json"
)

func sendErrorResponse(c *gin.Context, sgr *zap.SugaredLogger, code int, err error) {
	sgr.With(
		logKeyURI, c.Request.URL.Path,
		logKeyIP, c.ClientIP(),
	).Error(
		err,
	)

	message := response.Response{
		Message: http.StatusText(code),
	}

	bytes, err := json.Marshal(message)
	if err != nil {
		sgr.With(
			logKeyURI, c.Request.URL.Path,
			logKeyIP, c.ClientIP(),
		).Error(
			err,
		)
		return
	}

	c.Header(contentType, applicationJSON)
	c.Header(contentLength, strconv.Itoa(len(bytes)))

	c.JSON(code, message)
}
