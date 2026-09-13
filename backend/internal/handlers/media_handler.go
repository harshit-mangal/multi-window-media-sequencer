package handlers

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/models"
	"media-sequencer-backend/internal/services"
)

type MediaHandler struct {
	mediaService services.MediaService
}

func NewMediaHandler(mediaService services.MediaService) *MediaHandler {
	return &MediaHandler{
		mediaService: mediaService,
	}
}

// GetAll returns all available media items
// GET /api/media
func (h *MediaHandler) GetAll(c *gin.Context) {
	mediaList, err := h.mediaService.GetAll(c.Request.Context())
	if err != nil {
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve media list", err.Error())
		return
	}
	JSONSuccess(c, http.StatusOK, mediaList)
}

// GetByID returns a single media item by ID
// GET /api/media/:id
func (h *MediaHandler) GetByID(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ID", "Media ID must be a positive integer", nil)
		return
	}

	media, err := h.mediaService.GetByID(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "media item not found" {
			JSONError(c, http.StatusNotFound, "MEDIA_NOT_FOUND", "The requested media item does not exist", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to retrieve media item", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, media)
}

// Create registers a new media item
// POST /api/media
func (h *MediaHandler) Create(c *gin.Context) {
	var req models.CreateMediaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		JSONError(c, http.StatusBadRequest, "VALIDATION_ERROR", "Invalid media payload", err.Error())
		return
	}

	media, err := h.mediaService.Create(c.Request.Context(), req)
	if err != nil {
		JSONError(c, http.StatusBadRequest, "CREATE_MEDIA_FAILED", err.Error(), nil)
		return
	}

	JSONSuccess(c, http.StatusCreated, media)
}

// Delete removes a media item
// DELETE /api/media/:id
func (h *MediaHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.Atoi(idParam)
	if err != nil || id <= 0 {
		JSONError(c, http.StatusBadRequest, "INVALID_ID", "Media ID must be a positive integer", nil)
		return
	}

	if err := h.mediaService.Delete(c.Request.Context(), id); err != nil {
		if err.Error() == "media not found" {
			JSONError(c, http.StatusNotFound, "MEDIA_NOT_FOUND", "The requested media item does not exist", nil)
			return
		}
		JSONError(c, http.StatusInternalServerError, "DELETE_MEDIA_FAILED", "Failed to delete media item", err.Error())
		return
	}

	JSONSuccess(c, http.StatusOK, gin.H{"deleted": true, "id": id})
}
