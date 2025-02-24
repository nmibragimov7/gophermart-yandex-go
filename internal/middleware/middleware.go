package middleware

import (
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/session"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func AuthMiddleware(sgr *zap.SugaredLogger, cnf *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		ssp := session.SessionProvider{
			Config: cnf,
		}
		err := ssp.CheckToken(c)
		if err != nil {
			sgr.Errorw(
				"failed to verify token",
				"info", "incorrect access token",
			)

			//message := response.Response{
			//	Message: "Невалидный токен",
			//}

			//c.Header("Content-Type", "application/json")
			//c.JSON(http.StatusUnauthorized, message)
			c.Status(http.StatusUnauthorized)
			c.Abort()
			return
		}

		c.Next()
	}
}
