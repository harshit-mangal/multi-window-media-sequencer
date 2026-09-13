package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/services"
)

type WindowHandler struct {
	windowService services.WindowService
}

func NewWindowHandler(windowService services.WindowService) *WindowHandler {
	return &WindowHandler{
		windowService: windowService,
	}
}

// GetAllWindows returns all display windows
// GET /api/windows
func (h *WindowHandler) GetAllWindows(c *gin.Context) {
	windows, err := h.windowService.GetAllWindows(c.Request.Context())
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve windows", err.Error())
		return
	}
	JSONSuccess(c, http.StatusOK, windows)
}

// GetWindowByID returns window details, its playlist, and its current computed playback state
// GET /api/windows/:id
func (h *WindowHandler) GetWindowByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ID", "Window ID must be a positive integer", nil)
		return
	}

	window, playbackState, err := h.windowService.GetWindowByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "window not found" {
			JSONError(c, http.StatusNotFound, "WINDOW_NOT_FOUND", "The requested window does not exist", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve window", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{
		"window":        window,
		"playbackState": playbackState,
	})
}
