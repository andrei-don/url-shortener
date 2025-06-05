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
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	psqlPort  = 5432
	redisPort = 6379
)

func startMetricsServer() {
	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting metrics server on :9090")
	if err := http.ListenAndServe(":9090", nil); err != nil {
		log.Fatalf("Failed to start metrics server: %v", err)
	}
}

func main() {
	host := os.Getenv("DATABASE_HOST")
	user := os.Getenv("DATABASE_USER")
	password := os.Getenv("DATABASE_PASSWORD")
	dbname := os.Getenv("DATABASE_NAME")
	baseUrl := os.Getenv("BASE_URL")
	addr := fmt.Sprintf("%s:%d", os.Getenv("REDIS_HOST"), redisPort)

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=disable",
		host, psqlPort, user, password, dbname)

	dbPsql, err := config.ConnectDatabase(psqlInfo, 10, 5*time.Second)
	if err != nil {
		log.Fatalf("Cannot connect to Postgres: %v", err)
	}

	dbRedis, err := config.ConnectRedis(addr)
	if err != nil {
		log.Fatalf("Cannot connect to Redis: %v", err)
	}

	go startMetricsServer()

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	r.Static("/static", "./static")
	r.StaticFile("/", "./static/index.html")

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

	r.POST("/shorten", handlers.ShortenUrlHandler(dbPsql, dbRedis, baseUrl))
	r.GET("/:shortUrl", handlers.RedirectHandler(dbPsql, dbRedis))
	r.Run(":8080")
}
