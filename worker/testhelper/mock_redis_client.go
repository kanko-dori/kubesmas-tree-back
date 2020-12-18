package testhelper

import (
	"github.com/go-redis/redis"
	"testing"
)

func NewMockRedis(t *testing.T) *redis.Client {
	t.Helper()

	s, err := miniredis.Run()
	if err != nil {
		t.Fatalf("unexpected error while creating test redis server '%#v'", err)
	}
	client := redis.NewClient(&redis.Options{
		Addr:     s.Addr(),
		Password: "",
		DB:       0,
	})
	return client
}
