package main

import (
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"time"
)

const (
	tokenKey = "TOKEN_"
	tokenDuration = time.Hour * 24
)

func New(dsn string) (*redis.Client, error){
	client := redis.NewClient(&redis.Options{
		Addr: dsn,
		Password: "",
		DB: 0,
	})
	if err := client.Ping().Err(); err != nil {
		return nil, errors.Wrapf(err, "failed to ping redis server")
	}
	return client, nil
}

func SetToken(cli)


