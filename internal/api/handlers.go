package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	walletService WalletService
}

type WalletService interface {
	CreateWallet() (*Wallet, error)
	SignData(walletAddress string, data []byte) ([]byte, error)
	ListWallets() []string
}

func NewHandler(service WalletService) *Handler {
	return &Handler{
		walletService: service,
	}
}

func (h *Handler) CreateWallet(c *gin.Context) {
	wallet, err := h.walletService.CreateWallet()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"address": wallet.Address})
}

func (h *Handler) ListWallets(c *gin.Context) {
	wallets := h.walletService.ListWallets()
	c.JSON(http.StatusOK, gin.H{"wallets": wallets})
}

func (h *Handler) SignData(c *gin.Context) {
	wallet := c.Query("wallet")
	data := c.Query("data")

	if wallet == "" || data == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wallet and data parameters are required"})
		return
	}

	signature, err := h.walletService.SignData(wallet, []byte(data))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"signature": signature})
}
