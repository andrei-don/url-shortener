package handlers

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

type ShortenRequest struct {
	URL string
}

func shortenUrl(c *gin.Context, dbPsql *sql.DB, dbRedis *redis.Client) {
	var req ShortenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	shortCode := utils.GenerateShortCode(req.URL)

	var existingShortURL string
	err := dbPsql.QueryRow("SELECT short_url FROM urls WHERE original_url = $1", req.URL).Scan(&existingShortURL)
	if err == nil {
		c.JSON(http.StatusOK, gin.H{
			"message":   fmt.Sprintf("URL %s is already shortened.", req.URL),
			"short_url": "http://localhost:8080/" + existingShortURL,
		})
		return
	}

	_, err = dbPsql.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortCode, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	dbRedis.Set(config.Ctx, shortCode, req.URL, 0)

	c.JSON(http.StatusOK, gin.H{"short_url": "http://localhost:8080/" + shortCode})
}

func ShortenUrlHandler(dbPsql *sql.DB, dbRedis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortenUrl(c, dbPsql, dbRedis)
	}
}
