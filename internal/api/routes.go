package api

import (
	"github.com/gin-gonic/gin"
)

func SetupRoutes(h *Handler) *gin.Engine {
	router := gin.Default()

	router.POST("/wallet", h.CreateWallet)
	router.GET("/wallets", h.ListWallets)
	router.GET("/sign", h.SignData)

	return router
}
