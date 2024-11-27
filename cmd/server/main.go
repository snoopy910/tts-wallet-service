package main

import (
	"log"

	"github.com/snoopy910/tss-wallet-service/internal/api"
	"github.com/snoopy910/tss-wallet-service/internal/service"
	"github.com/snoopy910/tss-wallet-service/internal/storage"
)

func main() {
	// Initialize storage
	store := storage.NewMemoryStorage()

	// Initialize wallet service
	walletService := service.NewWalletService(store)

	// Initialize API handlers
	handler := api.NewHandler(walletService)

	// Setup and run server
	router := api.SetupRoutes(handler)

	log.Fatal(router.Run(":8080"))
}
