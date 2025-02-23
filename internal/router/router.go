package router

import (
	"go-musthave-diploma-tpl/internal/common"
	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/middleware"
	"go-musthave-diploma-tpl/internal/repository"
	"go-musthave-diploma-tpl/internal/session"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type RouterProvider struct {
	Sugar      *zap.SugaredLogger
	Config     *config.Config
	Handler    common.Handler
	Session    *session.SessionProvider
	Repository *repository.RepositoryProvider
}

func (p *RouterProvider) Router() *gin.Engine {
	r := gin.Default()
	sugarWithCtx := p.Sugar.With(
		"app", "gophermart",
		"service", "main",
		"func", "Router",
	)

	r.Use(middleware.GzipMiddleware(sugarWithCtx))
	r.POST("/api/user/register", p.Handler.RegisterHandler)
	r.POST("/api/user/login", p.Handler.LoginHandler)

	r.Use(middleware.AuthMiddleware(sugarWithCtx, p.Config))
	r.POST("/api/user/orders", p.Handler.OrderSaveHandler)
	r.GET("/api/user/orders", p.Handler.OrdersHandler)
	r.GET("/api/user/balance", p.Handler.BalanceHandler)
	r.POST("/api/user/balance/withdraw", p.Handler.WithdrawHandler)
	r.GET("/api/user/withdrawals", p.Handler.WithdrawalsHandler)

	return r
}
