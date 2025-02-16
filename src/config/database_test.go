package config

import (
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestConnectDatabase(t *testing.T) {
	dbPsql, mockPsql, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	assert.NoError(t, err)

	t.Run("Successful Connection", func(t *testing.T) {

		// Expect db.Ping to be called and return no error
		mockPsql.ExpectPing().WillReturnError(nil)

		// Replace sql.Open with a function that returns the mockDB
		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			return dbPsql, nil
		}

		db, err := ConnectDatabase("mock_dsn")

		assert.NoError(t, err)
		assert.NotNil(t, db)
		assert.NoError(t, mockPsql.ExpectationsWereMet())
	})

	t.Run("Failure in sql.Open", func(t *testing.T) {
		expectedErr := errors.New("failed to open database")

		// Mock sql.Open to return an error
		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			return nil, expectedErr
		}

		db, err := ConnectDatabase("invalid_dsn")
		assert.Nil(t, db)
		assert.ErrorIs(t, err, expectedErr)
	})

	t.Run("Failure in db.Ping", func(t *testing.T) {

		// Expect db.Ping to return an error
		expectedErr := errors.New("database is not reachable")
		mockPsql.ExpectPing().WillReturnError(expectedErr)

		// Mock sql.Open to return the mockDB
		sqlOpen = func(driverName string, dataSourceName string) (*sql.DB, error) {
			return dbPsql, nil
		}

		db, err := ConnectDatabase("mock_dsn")
		assert.Nil(t, db)
		assert.EqualError(t, err, fmt.Sprintf("database is not reachable: %v", expectedErr))
		assert.NoError(t, mockPsql.ExpectationsWereMet())
	})
}
