package redis

import (
	"main/testhelper"
	"testing"
)

func TestSetToken(t *testing.T) {
	client := testhelper.NewMockRedis(t)

	if err := redis.SetToken(client, "test", 1); err != nil {
		t.Fatalf("unexpected error while SetToken '%#v'", err)
	}
}