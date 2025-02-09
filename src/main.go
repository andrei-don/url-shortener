package main

import (
	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	config.ConnectDatabase()
	config.ConnectRedis()
	r := gin.Default()
	r.POST("/shorten", handlers.ShortenURL)
	r.GET("/:shortUrl", handlers.Redirect)
	r.Run(":8080")
}
