package handlers

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/andrei-don/url-shortener/config"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
)

func TestRedirect_CacheHit(t *testing.T) {
	db, mock := redismock.NewClientMock()
	config.RedisClient = db

	shortUrl := "test"
	originalUrl := "http://this-is-my-url.com"
	mock.ExpectGet(shortUrl).SetVal(originalUrl)

	router := gin.Default()
	router.GET("/:shortUrl", Redirect)

	req, _ := http.NewRequest("GET", "/"+shortUrl, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, originalUrl, w.Header().Get("Location"))

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestRedirect_CacheMissAndDBSuccess(t *testing.T) {
	dbRedis, mockRedis := redismock.NewClientMock()
	config.RedisClient = dbRedis

	dbPsql, mockPsql, err := sqlmock.New()
	assert.NoError(t, err)
	config.DB = dbPsql

	shortUrl := "test"
	originalUrl := "http://this-is-my-url.com"

	mockRedis.ExpectGet(shortUrl).RedisNil()

	rows := sqlmock.NewRows([]string{"original_url"}).AddRow(originalUrl)
	mockPsql.ExpectQuery("SELECT original_url FROM urls WHERE short_url = \\$1").WithArgs(shortUrl).WillReturnRows(rows)

	mockRedis.ExpectSet(shortUrl, originalUrl, 0)

	router := gin.Default()
	router.GET("/:shortUrl", Redirect)

	req, _ := http.NewRequest("GET", "/"+shortUrl, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, originalUrl, w.Header().Get("Location"))

	assert.NoError(t, mockRedis.ExpectationsWereMet())
	assert.NoError(t, mockPsql.ExpectationsWereMet())
}

func TestRedirect_URLNotFound(t *testing.T) {
	dbRedis, mockRedis := redismock.NewClientMock()
	config.RedisClient = dbRedis

	dbPsql, mockPsql, err := sqlmock.New()
	assert.NoError(t, err)
	config.DB = dbPsql

	shortUrl := "test-non-existent"

	mockRedis.ExpectGet(shortUrl).RedisNil()

	mockPsql.ExpectQuery("SELECT original_url FROM urls WHERE short_url = \\$1").WithArgs(shortUrl).WillReturnError(sql.ErrNoRows)

	router := gin.Default()
	router.GET("/:shortUrl", Redirect)

	req, _ := http.NewRequest("GET", "/"+shortUrl, nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.JSONEq(t, `{"error": "URL not found"}`, w.Body.String())

	assert.NoError(t, mockRedis.ExpectationsWereMet())
	assert.NoError(t, mockPsql.ExpectationsWereMet())
}
