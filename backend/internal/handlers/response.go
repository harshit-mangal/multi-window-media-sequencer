package handlers

import (
	"github.com/gin-gonic/gin"

	"media-sequencer-backend/internal/models"
)

func JSONError(c *gin.Context, status int, code string, message string, details any) {
	c.JSON(status, models.ErrorResponse{
		Error: models.ErrorDetail{
			Code:    code,
			Message: message,
			Details: details,
		},
	})
}

func JSONSuccess(c *gin.Context, status int, data any) {
	c.JSON(status, gin.H{
		"data": data,
	})
}
