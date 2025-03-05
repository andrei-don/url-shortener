package handlers

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/andrei-don/url-shortener/config"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/redis/go-redis/v9"
)

var (
	redirectCounterRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "redirect_requests_total",
			Help: "Total number of requests to the /redirect endpoint.",
		},
		[]string{"status"},
	)

	redirectHistogramLatency = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "redirect_request_latency_seconds",
			Help:    "Latency of requests to the /redirect endpoint.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	redisCacheHits = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_hits_total",
			Help: "Total number of Redis cache hits.",
		},
	)

	redisCacheMisses = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "redis_cache_misses_total",
			Help: "Total number of Redis cache misses.",
		},
	)
)

func init() {
	prometheus.MustRegister(redirectCounterRequests)
	prometheus.MustRegister(redirectHistogramLatency)
	prometheus.MustRegister(redisCacheHits)
	prometheus.MustRegister(redisCacheMisses)
}

func redirect(c *gin.Context, dbPsql *sql.DB, dbRedis *redis.Client) {
	shortCode := c.Param("shortUrl")

	start := time.Now()
	url, err := dbRedis.Get(config.Ctx, shortCode).Result()
	if err == nil {
		c.Redirect(http.StatusFound, url)
		log.Printf("Cache hit for %s", url)
		redirectCounterRequests.WithLabelValues("302").Inc()
		redirectHistogramLatency.WithLabelValues("302").Observe(time.Since(start).Seconds())
		redisCacheHits.Inc()
		return
	}

	row := dbPsql.QueryRow("SELECT original_url FROM urls WHERE short_url = $1", shortCode)
	err = row.Scan(&url)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "URL not found"})
		redirectCounterRequests.WithLabelValues("404").Inc()
		redirectHistogramLatency.WithLabelValues("404").Observe(time.Since(start).Seconds())
		return
	}

	// When there is a cache miss, we write the shortCode:url key-value pair in redis.
	dbRedis.Set(config.Ctx, shortCode, url, 0)
	log.Printf("Cache miss for %s", url)
	c.Redirect(http.StatusFound, url)
	redirectCounterRequests.WithLabelValues("302").Inc()
	redirectHistogramLatency.WithLabelValues("302").Observe(time.Since(start).Seconds())
	redisCacheMisses.Inc()
}

func RedirectHandler(dbPsql *sql.DB, dbRedis *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		redirect(c, dbPsql, dbRedis)
	}
}
