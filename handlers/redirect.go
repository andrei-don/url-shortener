package handlers

import (
	"net/http"

	"github.com/andrei-don/url-shortener/config"
	"github.com/gin-gonic/gin"
)

func Redirect(c *gin.Context) {
	shortCode := c.Param("shortUrl")

	url, err := config.RedisClient.Get(config.Ctx, shortCode).Result()
	if err == nil {
		c.Redirect(http.StatusFound, url)
		return
	}

	row := config.DB.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortCode)
	err = row.Scan(&url)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// When there is a cache miss, we write the shortCode:url key-value pair in redis.
	config.RedisClient.Set(config.Ctx, shortCode, url, 0)
	c.Redirect(http.StatusFound, url)
}
