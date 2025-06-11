package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ethereum-validator-api/config"

	"ethereum-validator-api/ethereum"
	"ethereum-validator-api/handlers"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	log.Printf("Starting server with configuration:")
	log.Printf("  Port: %s", cfg.Port)
	log.Printf("  ETH RPC URL: %s", cfg.EthRPCURL)

	// Initialize Ethereum client
	ethClient, err := ethereum.NewClient(cfg.EthRPCURL)
	if err != nil {
		log.Fatalf("Failed to initialize Ethereum client: %v", err)
	}

	// Setup Gin router
	router := gin.Default()

	// Initialize handlers
	h := handlers.New(ethClient)

	// Routes
	router.GET("/blockreward/:slot", h.GetBlockReward)
	router.GET("/syncduties/:slot", h.GetSyncDuties)

	// Server configuration
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	log.Printf("Server started on :%s", cfg.Port)

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}
