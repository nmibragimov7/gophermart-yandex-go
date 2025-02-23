package common

import "github.com/gin-gonic/gin"

type Handler interface {
	RegisterHandler(c *gin.Context)
	LoginHandler(c *gin.Context)
	OrderSaveHandler(c *gin.Context)
	OrdersHandler(c *gin.Context)
	BalanceHandler(c *gin.Context)
	WithdrawHandler(c *gin.Context)
	WithdrawalsHandler(c *gin.Context)
}
