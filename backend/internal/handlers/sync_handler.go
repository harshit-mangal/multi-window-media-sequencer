package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
)

type SyncHandler struct {
	syncService services.SyncService
}

func NewSyncHandler(syncService services.SyncService) *SyncHandler {
	return &SyncHandler{
		syncService: syncService,
	}
}

// StartSync triggers a synchronized media playback override across all windows
// POST /api/sync
func (h *SyncHandler) StartSync(c *gin.Context) {
	var req models.StartSyncRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid sync request payload", err.Error())
		return
	}

	event, err := h.syncService.StartSync(c.Request.Context(), req)
	if err != nil {
		if err.Error() == "media item not found" {
			JSONError(c, http.StatusNotFound, "MEDIA_NOT_FOUND", "Selected sync media does not exist", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "SYNC_START_FAILED", "Failed to start synchronization", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, event)
}

// GetCurrentSync returns the currently active synchronization event, if any
// GET /api/sync/current
func (h *SyncHandler) GetCurrentSync(c *gin.Context) {
	resp, err := h.syncService.GetCurrentSync(c.Request.Context())
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get current sync state", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, resp)
}
