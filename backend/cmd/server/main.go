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

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/database"
	"media-sequencer-backend/internal/handlers"
	"media-sequencer-backend/internal/repository"
	"media-sequencer-backend/internal/services"
	"media-sequencer-backend/internal/websocket"
)

func main() {
	log.Println("[SERVER_INIT] Starting Multi-Window Media Sequencer Backend...")

	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize database connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Printf("[DATABASE_ERROR] Warning: Could not connect to database (%v). Server will start in degraded mode.", err)
	} else {
		defer db.Close()

		// Run migrations
		migrationCtx, mCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer mCancel()
		if err := db.RunMigrations(migrationCtx); err != nil {
			log.Fatalf("[DATABASE_ERROR] Failed to run database migrations: %v", err)
		}

		// Seed initial data if needed
		seedCtx, sCancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer sCancel()
		if err := db.SeedInitialData(seedCtx, false); err != nil {
			log.Printf("[DATABASE_ERROR] Error seeding initial data: %v", err)
		}
	}

	// 3. Initialize WebSocket Hub
	wsHub := websocket.NewHub()
	go wsHub.Run()

	// 4. Initialize Repositories
	var windowRepo repository.WindowRepository
	var mediaRepo repository.MediaRepository
	var playlistRepo repository.PlaylistRepository
	var syncRepo repository.SyncRepository

	if db != nil {
		windowRepo = repository.NewWindowRepository(db.Pool)
		mediaRepo = repository.NewMediaRepository(db.Pool)
		playlistRepo = repository.NewPlaylistRepository(db.Pool)
		syncRepo = repository.NewSyncRepository(db.Pool)
	}

	// 5. Initialize Services
	playbackService := services.NewPlaybackService(cfg, syncRepo)
	syncService := services.NewSyncService(cfg, syncRepo, mediaRepo, wsHub)
	windowService := services.NewWindowService(windowRepo, playlistRepo, playbackService)
	mediaService := services.NewMediaService(mediaRepo)
	playlistService := services.NewPlaylistService(playlistRepo, windowRepo, mediaRepo, wsHub)

	// 6. Initialize Handlers
	windowHandler := handlers.NewWindowHandler(windowService)
	mediaHandler := handlers.NewMediaHandler(mediaService)
	playlistHandler := handlers.NewPlaylistHandler(playlistService)
	syncHandler := handlers.NewSyncHandler(syncService)
	wsHandler := handlers.NewWebSocketHandler(wsHub)

	// 7. Setup Router
	router := handlers.SetupRouter(handlers.RouterDependencies{
		Config:          cfg,
		WindowHandler:   windowHandler,
		MediaHandler:    mediaHandler,
		PlaylistHandler: playlistHandler,
		SyncHandler:     syncHandler,
		WSHandler:       wsHandler,
	})

	// 8. Configure HTTP Server
	serverAddr := fmt.Sprintf(":%s", cfg.Port)
	srv := &http.Server{
		Addr:         serverAddr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// 9. Start Server in Goroutine
	go func() {
		log.Printf("[SERVER_STARTED] HTTP & WebSocket Server running at http://0.0.0.0%s", serverAddr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[SERVER_ERROR] Listen error: %v", err)
		}
	}()

	// 10. Wait for interrupt signal for Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[SERVER_SHUTDOWN] Shutting down server gracefully...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("[SERVER_SHUTDOWN] Server forced to shutdown: %v", err)
	}

	log.Println("[SERVER_SHUTDOWN] Server exiting cleanly")
}
