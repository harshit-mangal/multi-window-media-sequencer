package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/config"
	"media-sequencer-backend/internal/middleware"
)

type RouterDependencies struct {
	Config          *config.Config
	WindowHandler   *WindowHandler
	MediaHandler    *MediaHandler
	PlaylistHandler *PlaylistHandler
	SyncHandler     *SyncHandler
	WSHandler       *WebSocketHandler
}

// SetupRouter initializes Gin engine with all routes, middleware, and handlers
func SetupRouter(deps RouterDependencies) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Global Middlewares
	router.Use(gin.Recovery())
	router.Use(gin.Logger())
	router.Use(middleware.CORSMiddleware(deps.Config.AllowedOrigins))

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "healthy",
			"service": "media-sequencer-backend",
		})
	})

	// WebSocket endpoint
	router.GET("/ws", deps.WSHandler.HandleWS)

	// API route group
	api := router.Group("/api")
	{
		// Windows routes
		api.GET("/windows", deps.WindowHandler.GetAllWindows)
		api.GET("/windows/:id", deps.WindowHandler.GetWindowByID)

		// Playlist routes for windows
		api.GET("/windows/:id/playlist", deps.PlaylistHandler.GetPlaylist)
		api.POST("/windows/:id/playlist", deps.PlaylistHandler.AddItem)
		api.PUT("/windows/:id/playlist/:itemId", deps.PlaylistHandler.UpdateItem)
		api.DELETE("/windows/:id/playlist/:itemId", deps.PlaylistHandler.RemoveItem)

		// Media routes
		api.GET("/media", deps.MediaHandler.GetAll)
		api.GET("/media/:id", deps.MediaHandler.GetByID)
		api.POST("/media", deps.MediaHandler.Create)
		api.DELETE("/media/:id", deps.MediaHandler.Delete)

		// Synchronization routes
		api.POST("/sync", deps.SyncHandler.StartSync)
		api.GET("/sync/current", deps.SyncHandler.GetCurrentSync)
	}

	return router
}
