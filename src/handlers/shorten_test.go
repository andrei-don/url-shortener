package handlers

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrei-don/url-shortener/config"
	"github.com/andrei-don/url-shortener/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestShorten_BadRequest(t *testing.T) {

	router := gin.Default()
	router.POST("/shorten", ShortenURL)

	reqBody := []byte(`{"url": "http://example.com"`)

	req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.JSONEq(t, `{"error": "Invalid request"}`, w.Body.String())
}

func TestShorten_ExistingUrl(t *testing.T) {

	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	config.DB = db

	router := gin.Default()
	router.POST("/shorten", ShortenURL)

	originalUrl := "http://this-is-my-url.com"
	existingShortUrl := "test"

	rows := sqlmock.NewRows([]string{"short_url"}).AddRow(existingShortUrl)

	mock.ExpectQuery("SELECT short_url FROM urls WHERE original_url = \\$1").WithArgs(originalUrl).WillReturnRows(rows)

	reqBody := []byte(`{"url": "http://this-is-my-url.com"}`)

	req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(reqBody))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, fmt.Sprintf(`{
		"message": "URL %s is already shortened.",
		"short_url": "http://localhost:8080/%s"
	}`, originalUrl, existingShortUrl), w.Body.String())

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShorten_DatabaseInsert(t *testing.T) {

	dbRedis, mockRedis := redismock.NewClientMock()
	config.RedisClient = dbRedis

	dbPsql, mockPsql, err := sqlmock.New()
	assert.NoError(t, err)
	config.DB = dbPsql

	//shortUrl := "test"
	originalUrl := "http://this-is-my-url.com"
	shortUrl := utils.GenerateShortCode(originalUrl)

	router := gin.Default()
	router.POST("/shorten", ShortenURL)

	t.Run("Insert Success", func(t *testing.T) {
		mockPsql.ExpectExec(`INSERT INTO urls \(short_url, original_url\) VALUES \(\$1, \$2\)`).WithArgs(shortUrl, originalUrl).WillReturnResult(sqlmock.NewResult(1, 1))
		mockRedis.ExpectSet(shortUrl, originalUrl, 0)

		reqBody := []byte(`{"url": "http://this-is-my-url.com"}`)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.JSONEq(t, fmt.Sprintf(`{
			"short_url": "http://localhost:8080/%s"
		}`, shortUrl), w.Body.String())
		assert.NoError(t, mockPsql.ExpectationsWereMet())
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})

	t.Run("Insert Failure", func(t *testing.T) {
		mockPsql.ExpectExec(`INSERT INTO urls \(short_url, original_url\) VALUES \(\$1, \$2\)`).WithArgs(shortUrl, originalUrl).WillReturnError(fmt.Errorf("database error"))

		reqBody := []byte(`{"url": "http://this-is-my-url.com"}`)

		req, _ := http.NewRequest("POST", "/shorten", bytes.NewBuffer(reqBody))
		w := httptest.NewRecorder()

		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
		assert.JSONEq(t, `{"error": "Database error"}`, w.Body.String())
		assert.NoError(t, mockPsql.ExpectationsWereMet())
	})

}
