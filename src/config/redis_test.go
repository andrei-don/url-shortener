package config

import (
	"fmt"
	"testing"

	"github.com/go-redis/redismock/v9"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestConnectRedis(t *testing.T) {
	dbRedis, mockRedis := redismock.NewClientMock()

	t.Run("Successful Connection", func(t *testing.T) {
		mockRedis.ExpectPing().SetVal("PONG")
		redisClient = func(opt *redis.Options) *redis.Client {
			return dbRedis
		}
		client, err := ConnectRedis("mockdsn")
		assert.NotNil(t, client)
		assert.NoError(t, err)
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})
	t.Run("Failed Connection", func(t *testing.T) {
		mockRedis.ExpectPing().SetErr(fmt.Errorf("redis connection failed"))
		redisClient = func(opt *redis.Options) *redis.Client {
			return dbRedis
		}
		client, err := ConnectRedis("mockdsn")
		assert.Nil(t, client)
		assert.Equal(t, "redis connection failed: redis connection failed", err.Error())
		assert.NoError(t, mockRedis.ExpectationsWereMet())
	})

}
