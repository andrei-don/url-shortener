package handlers

import (
	"net/http"

	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/utils"
	"github.com/gin-gonic/gin"
)

type ShortenRequest struct {
	URL string
}

func ShortenURL(c *gin.Context) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortCode := utils.GenerateShortCode(req.URL)

	_, err := config.DB.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortCode, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	config.RedisClient.Set(config.Ctx, shortCode, req.URL, 0)

	c.JSON(http.StatusOK, gin.H{"short_url": "http://localhost:8080/" + shortCode})
}
