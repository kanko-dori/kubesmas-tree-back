package main

import (
	"fmt"
	"github.com/go-redis/redis"
	"github.com/pkg/errors"
	"log"
	"os"
	"strconv"
	"time"
)

type IlluminationData struct {
	Pattern1 int `json:"pattern1"` //TODO: パターンが確定したら増やすか減らすかする
	Pattern2 int `json:"pattern2"`
	Pattern3 int `json:"pattern3"`
}

func main(){
	fmt.Println("Check for first vote")
	checkVote()
	fmt.Println("sleep for 25 sec")
	time.Sleep(time.Second * 25)

	fmt.Println("Check for second vote")
	checkVote()
}

func checkVote(){
	redisPath := os.Getenv("REDIS_PATH")
	client, err := getNewRedis(redisPath)
	if err != nil {
		fmt.Printf("failed to get redis client: %v", err)
		os.Exit(-1)
		return
	}
	defer client.Close()

	p1, err := getValue(client, "PATTERN_1")
	if err != nil && err != redis.Nil {
		fmt.Printf("failed to get value from client: %v", err)
		os.Exit(-1)
		return
	}
	p2, err := getValue(client, "PATTERN_2")
	if err != nil && err != redis.Nil {
		fmt.Printf("failed to get value from client: %v", err)
		os.Exit(-1)
		return
	}
	p3, err := getValue(client, "PATTERN_3")
	if err != nil && err != redis.Nil {
		fmt.Printf("failed to get value from client: %v", err)
		os.Exit(-1)
		return
	}

	fmt.Printf("p1: %d, p2: %d, p3: %d\n", p1, p2, p3)
	if p1 > p2 || p1 > p3 {
		// p1 is max
		err := setValue(client, "CURRENT_ILLUMINATION_PATTERN", 1)
		if err != nil {
			fmt.Printf("failed to setValue: %v\n", err)
			os.Exit(-1)
			return
		}
		fmt.Println("current pattern is 1")
	} else if p2 > p1 || p1 > p3 {
		// p2 is max
		err := setValue(client, "CURRENT_ILLUMINATION_PATTERN", 2)
		if err != nil {
			fmt.Printf("failed to setValue: %v\n", err)
			os.Exit(-1)
			return
		}
		fmt.Println("current pattern is 2")
	} else if p3 > p1 || p3 > p1 {
		// p3 is max
		err := setValue(client, "CURRENT_ILLUMINATION_PATTERN", 3)
		if err != nil {
			fmt.Printf("failed to setValue: %v\n", err)
			os.Exit(-1)
			return
		}
		fmt.Println("current pattern is 3")
	}
	err = setValue(client, "PATTERN_1", 0)
	if  err == redis.Nil && err != nil {
		fmt.Printf("failed to setValue: %v\n", err)
		os.Exit(-1)
		return
	}
	err = setValue(client, "PATTERN_2", 0)
	if err == redis.Nil && err != nil {
		fmt.Printf("failed to setValue: %v\n", err)
		os.Exit(-1)
		return
	}
	err = setValue(client, "PATTERN_3", 0)
	if err == redis.Nil && err != nil {
		fmt.Printf("failed to setValue: %v\n", err)
		os.Exit(-1)
		return
	}

	fmt.Println("Successfully initialized patterns.")
}

func getNewRedis(dsn string) (*redis.Client, error){
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

func getValue(client *redis.Client,target string) (int, error) {
	v, err := client.Get(target).Result()
	if err == redis.Nil {
		log.Printf("%s does not exist. creating now...\n", target)

		err = client.Set(target, 1, time.Hour*24).Err()
		if err != nil {
			return 0, errors.Wrap(err, "failed to set client")
		}
	} else if err != nil {
		return 0, errors.Wrapf(err, "failed to get %s", target)
	}
	i, err := strconv.Atoi(v)
	return i, nil
}

func setValue(client *redis.Client, target string, number int) error {
	err := client.Get(target).Err()
	if err == redis.Nil {
		log.Printf("does not exists: %s\n", target)
	} else if err != nil {
		return errors.Wrapf(err, "failed to get %s", target)
	} else {
		log.Printf("%s does not exist. creating now...\n", target)
		err = client.Set(target, number, time.Hour*24).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set client")
		}
	}
	return nil
}