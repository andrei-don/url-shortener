package handlers

import (
	"database/sql"
	"net/http"

	"github.com/andrei-don/url-shortener/config"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

func redirect(c *gin.Context, dbPsql *sql.DB, dbRedis *redis.Client) {
	shortCode := c.Param("shortUrl")

	url, err := dbRedis.Get(config.Ctx, shortCode).Result()
	if err == nil {
		c.Redirect(http.StatusFound, url)
		return
	}

	row := dbPsql.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortCode)
	err = row.Scan(&url)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		return
	}

	// When there is a cache miss, we write the shortCode:url key-value pair in redis.
	dbRedis.Set(config.Ctx, shortCode, url, 0)
	c.Redirect(http.StatusFound, url)
}

func RedirectHandler(dbPsql *sql.DB, dbRedis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		redirect(c, dbPsql, dbRedis)
	}
}
