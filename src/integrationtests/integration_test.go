//go:build integration

package integrationtests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/utils"
	"github.com/stretchr/testify/assert"
)

const (
	baseUrl   = "http://localhost:8080"
	psqlPort  = 5432
	redisPort = 6379
)

var (
	addr      = fmt.Sprintf("%s:%d", os.Getenv("REDIS_HOST"), redisPort)
	testUrl   = "http://example.com"
	shortCode = utils.GenerateShortCode(testUrl)
	client    = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // Prevents following redirects
		},
	}
)

func Test_ShortenUrl(t *testing.T) {

	t.Run("Successful URL Shortening", func(t *testing.T) {
		requestBody, _ := json.Marshal(map[string]string{
			"url": testUrl,
		})
		resp, err := client.Post(baseUrl+"/shorten", "application/json", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseData map[string]string
		json.NewDecoder(resp.Body).Decode(&responseData)
		resp.Body.Close()

		shortUrl, existsShortUrl := responseData["short_url"]

		assert.True(t, existsShortUrl)
		assert.Equal(t, shortUrl, fmt.Sprintf("%s/%s", baseUrl, shortCode))
	})

	t.Run("URL already shortened", func(t *testing.T) {
		requestBody, _ := json.Marshal(map[string]string{
			"url": testUrl,
		})
		resp, err := client.Post(baseUrl+"/shorten", "application/json", bytes.NewBuffer(requestBody))
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var responseData map[string]string
		json.NewDecoder(resp.Body).Decode(&responseData)
		resp.Body.Close()

		shortUrl, existsShortUrl := responseData["short_url"]
		message, existsMessage := responseData["message"]

		assert.True(t, existsShortUrl)
		assert.True(t, existsMessage)
		assert.Equal(t, shortUrl, fmt.Sprintf("%s/%s", baseUrl, shortCode))
		assert.Equal(t, message, fmt.Sprintf("URL %s is already shortened.", testUrl))
	})

	t.Run("Invalid Request", func(t *testing.T) {
		reqBody := []byte(`{"url": "http://example.com"`)
		resp, _ := client.Post(baseUrl+"/shorten", "application/json", bytes.NewBuffer(reqBody))
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func Test_Redirect(t *testing.T) {

	t.Run("Redirects When URL Exists in Redis (Cache Hit)", func(t *testing.T) {

		resp, err := client.Get(fmt.Sprintf("%s/%s", baseUrl, shortCode))

		assert.NoError(t, err)

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, testUrl, resp.Header.Get("Location"))
	})

	t.Run("Redirects When URL does not exist in Redis (Cache Miss)", func(t *testing.T) {
		//removing shortCode from cache
		dbRedis, err := config.ConnectRedis(addr)
		assert.NoError(t, err)
		err = dbRedis.Del(config.Ctx, shortCode).Err()
		assert.NoError(t, err)

		resp, err := client.Get(fmt.Sprintf("%s/%s", baseUrl, shortCode))

		assert.NoError(t, err)

		assert.Equal(t, http.StatusFound, resp.StatusCode)
		assert.Equal(t, testUrl, resp.Header.Get("Location"))
	})

	t.Run("URL not found", func(t *testing.T) {
		nonExistentShortCode := "test"

		resp, err := client.Get(fmt.Sprintf("%s/%s", baseUrl, nonExistentShortCode))
		assert.NoError(t, err)

		var responseData map[string]string
		json.NewDecoder(resp.Body).Decode(&responseData)
		resp.Body.Close()
		t.Log(responseData)

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		assert.Equal(t, "URL not found", responseData["error"])
	})
}
