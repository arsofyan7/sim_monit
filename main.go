package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"sim_monit/server/config"
	"sim_monit/server/db"
	"sim_monit/server/routes"
	"sim_monit/server/worker"
)

func main() {
	log.Println("==================================================")
	log.Println("           SIM_MONIT MONITORING ENGINE           ")
	log.Println("==================================================")

	// 1. Load Configuration
	cfg := config.LoadConfig()

	// 2. Initialize SQLite Database with PRAGMA WAL
	database, err := db.InitDB(cfg.DBPath)
	if err != nil {
		log.Fatalf("Database initialization failed: %v", err)
	}
	defer database.Close()

	// 3. Initialize & Start Background Workers
	scheduler := worker.InitScheduler(database)
	scheduler.Start()

	aggregator := worker.InitAggregator(database)
	aggregator.Start()

	purger := worker.InitPurger(database)
	purger.Start()

	// 4. Setup Gin HTTP Router
	router := routes.SetupRouter(database, cfg)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Port),
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 5. Graceful Server Startup
	go func() {
		log.Printf("[Server] SIM_MONIT HTTP service listening on port %s (Mode: %s)\n", cfg.Port, cfg.GinMode)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[Server] Failed to listen: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[Server] Shutting down SIM_MONIT gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[Server] Server forced to shutdown: %v", err)
	}

	log.Println("[Server] SIM_MONIT terminated cleanly.")
}
