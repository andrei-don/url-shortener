package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

type ShortenRequest struct {
	URL string
}

var (
	shortenCounterRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "shorten_requests_total",
			Help: "Total number of requests to the /shorten endpoint.",
		},
		[]string{"status"},
	)

	shortenHistogramLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "shorten_request_latency_seconds",
			Help:    "Latency of requests to the /shorten endpoint.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)
)

func init() {
	prometheus.MustRegister(shortenCounterRequests)
	prometheus.MustRegister(shortenHistogramLatency)
}

func shortenUrl(c *gin.Context, dbPsql *sql.DB, dbRedis *redis.Client) {
	var req ShortenRequest

	start := time.Now()
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		shortenCounterRequests.WithLabelValues("400").Inc()
		shortenHistogramLatency.WithLabelValues("400").Observe(time.Since(start).Seconds())
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
		shortenCounterRequests.WithLabelValues("200").Inc()
		shortenHistogramLatency.WithLabelValues("200").Observe(time.Since(start).Seconds())
		return
	}

	_, err = dbPsql.Exec("INSERT INTO urls (short_url, original_url) VALUES ($1, $2)", shortCode, req.URL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		shortenCounterRequests.WithLabelValues("500").Inc()
		shortenHistogramLatency.WithLabelValues("500").Observe(time.Since(start).Seconds())
		return
	}

	dbRedis.Set(config.Ctx, shortCode, req.URL, 0)

	c.JSON(http.StatusOK, gin.H{"short_url": "http://localhost:8080/" + shortCode})
	shortenCounterRequests.WithLabelValues("200").Inc()
	shortenHistogramLatency.WithLabelValues("200").Observe(time.Since(start).Seconds())
}

func ShortenUrlHandler(dbPsql *sql.DB, dbRedis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		shortenUrl(c, dbPsql, dbRedis)
	}
}
