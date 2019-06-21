// Package redistest implements support code for testing with Redis.
package redistest

import (
	"os"
	"testing"
	"time"

	"github.com/go-redis/redis"
)

// RedisCredentials holds the credentials for connecting to Redis.
type RedisCredentials struct {
	Username string
	Password string
	IP       string
}

// GetCredentials gets the Redis credentials from environment variables.
func GetCredentials() (rc RedisCredentials, ok bool) {
	u := os.Getenv("REDIS_USER")
	p := os.Getenv("REDIS_PASS")
	i := os.Getenv("REDIS_IP")
	if len(i) > 0 {
		return RedisCredentials{
			Username: u,
			Password: p,
			IP:       i,
		}, true
	}
	return RedisCredentials{}, false
}

// Connect connects to Redis and returns the Client object.
func Connect(t *testing.T) *redis.Client {
	creds, ok := GetCredentials()
	if !ok {
		t.Skip("Missing Redis credentials")
	}

	client := redis.NewClient(&redis.Options{
		Addr:         creds.IP,
		Password:     creds.Password,
		DB:           0,
		MaxRetries:   3,
		DialTimeout:  10 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	return client
}
