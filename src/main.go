package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/handlers"
	"github.com/gin-gonic/gin"
)

const (
	psqlPort  = 5432
	redisPort = 6379
)

func main() {
	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	addr := fmt.Sprintf("%s:%d", os.Getenv("REDIS_HOST"), redisPort)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, psqlPort, user, password, dbname)

	dbPsql, err := config.ConnectDatabase(psqlInfo, 10, 1*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to Postgres: %v", err)
	}

	dbRedis, err := config.ConnectRedis(addr)
	if err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	r := gin.Default()

	r.GET("/healthz", func(c *gin.Context) {

		if err := dbPsql.Ping(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"details": "PostgreSQL is not reachable",
			})
			return
		}

		if err := dbRedis.Ping(config.Ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":  "unhealthy",
				"details": "Redis is not reachable",
			})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"status": "healthy",
		})
	})

	r.POST("/shorten", handlers.ShortenUrlHandler(dbPsql, dbRedis))
	r.GET("/:shortUrl", handlers.RedirectHandler(dbPsql, dbRedis))
	r.Run(":8080")
}
