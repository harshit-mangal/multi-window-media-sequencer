package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
)

type PlaylistHandler struct {
	playlistService services.PlaylistService
}

func NewPlaylistHandler(playlistService services.PlaylistService) *PlaylistHandler {
	return &PlaylistHandler{
		playlistService: playlistService,
	}
}

// GetPlaylist returns all playlist items for a window
// GET /api/windows/:id/playlist
func (h *PlaylistHandler) GetPlaylist(c *gin.Context) {
	windowID, err := strconv.Atoi(c.Param("id"))
	if err != nil || windowID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ID", "Window ID must be a positive integer", nil)
		return
	}

	items, err := h.playlistService.GetPlaylist(c.Request.Context(), windowID)
	if err != nil {
		if err.Error() == "window not found" {
			JSONError(c, http.StatusNotFound, "WINDOW_NOT_FOUND", "The requested window does not exist", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve playlist", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, items)
}

// AddItem adds a media item to a window's playlist
// POST /api/windows/:id/playlist
func (h *PlaylistHandler) AddItem(c *gin.Context) {
	windowID, err := strconv.Atoi(c.Param("id"))
	if err != nil || windowID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ID", "Window ID must be a positive integer", nil)
		return
	}

	var req models.AddToPlaylistRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid playlist item payload", err.Error())
		return
	}

	item, err := h.playlistService.AddItem(c.Request.Context(), windowID, req)
	if err != nil {
		if err.Error() == "window not found" {
			JSONError(c, http.StatusNotFound, "WINDOW_NOT_FOUND", "Window not found", nil)
			return
		}
		if err.Error() == "media not found" {
			JSONError(c, http.StatusNotFound, "MEDIA_NOT_FOUND", "Media not found", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "ADD_ITEM_FAILED", "Failed to add item to playlist", err.Error())
		return
	}

	JSONSuccess(c, http.StatusCreated, item)
}

// UpdateItem updates a playlist item's position or assigned media
// PUT /api/windows/:id/playlist/:itemId
func (h *PlaylistHandler) UpdateItem(c *gin.Context) {
	windowID, err := strconv.Atoi(c.Param("id"))
	if err != nil || windowID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_WINDOW_ID", "Window ID must be a positive integer", nil)
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil || itemID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ITEM_ID", "Playlist Item ID must be a positive integer", nil)
		return
	}

	var req models.UpdatePlaylistItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid update payload", err.Error())
		return
	}

	item, err := h.playlistService.UpdateItem(c.Request.Context(), windowID, itemID, req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "UPDATE_ITEM_FAILED", err.Error(), nil)
		return
	}

	JSONSuccess(c, http.StatusOK, item)
}

// RemoveItem deletes an item from a window's playlist
// DELETE /api/windows/:id/playlist/:itemId
func (h *PlaylistHandler) RemoveItem(c *gin.Context) {
	windowID, err := strconv.Atoi(c.Param("id"))
	if err != nil || windowID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_WINDOW_ID", "Window ID must be a positive integer", nil)
		return
	}

	itemID, err := strconv.Atoi(c.Param("itemId"))
	if err != nil || itemID <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ITEM_ID", "Playlist Item ID must be a positive integer", nil)
		return
	}

	if err := h.playlistService.RemoveItem(c.Request.Context(), windowID, itemID); err != nil {
		JSONError(c, http.StatusInternalServerError, "REMOVE_ITEM_FAILED", "Failed to remove playlist item", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"deleted": true, "itemId": itemID})
}
