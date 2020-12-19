package main

import (
	"context"
	"encoding/json"
	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"main/persistence/redis"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/rand"
	"net/http"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Request struct {
	Action  string `json:"action"`
	Pattern int    `json:"pattern"`
	Uid     string `json:"uid"`
}

type IlluminationData struct {
	Pattern1 int `json:"pattern1"` //TODO: パターンが確定したら増やすか減らすかする
	Pattern2 int `json:"pattern2"`
	Pattern3 int `json:"pattern3"`
}

type GETResponse struct {
	Pods                int              `json:"pods"`
	IlluminationPattern int              `json:"illuminationPattern"`
	IlluminationData    IlluminationData `json:"illuminationData"`
}
type PostResponse struct {
	Response    string      `json:"response"`
	CurrentData GETResponse `json:"currentData"`
}

var errorResponse = []byte(`{"response":"NG"}`)
var voteLabel = []string{"PATTERN_1", "PATTERN_2", "PATTERN_3"}

func handler(s []byte) []byte {
	var r Request
	if err := json.Unmarshal(s, &r); err != nil {
		log.Println("cannot unmarshal request: %v, err: %v\n", s, err)
		return errorResponse
	}
	//fmt.Printf("%v\n", r)

	switch {
	case r.Action == "GET":
		r, err := getHandler()
		if err != nil {
			log.Println("failed to call getHandler: %v", err)
			return errorResponse
		}
		return r
	case r.Action == "VOTE":
		r, err := voteHandler(r.Uid, r.Pattern)
		if err != nil {
			log.Println("failed to call voteHandler: %v", err)
			return errorResponse
		}
		return r
	}
	return errorResponse
}

func iotEndpoint(w http.ResponseWriter, r *http.Request) {
	p, _ := getPods()
	i := rand.Intn(4)
	m := fmt.Sprintf("%d,%d", len(p.Items), i)
	fmt.Fprintf(w, m)
}

func getPods() (*v1.PodList, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods("kubesmas-tree").List(context.TODO(), metav1.ListOptions{
		LabelSelector: "app=nginx",
	})
	if err != nil {
		return nil, err
	}

	return pods, nil
}
func getVotedValue() (*IlluminationData, error) {
	a, err := getValue("PATTERN_1")
	if err != nil {
		return nil, err
	}
	b, err := getValue("PATTERN_2")
	if err != nil {
		return nil, err
	}
	c, err := getValue("PATTERN_3")
	if err != nil {
		return nil, err
	}

	var illuminationData = IlluminationData{
		Pattern1: a,
		Pattern2: b,
		Pattern3: c,
	}

	return &illuminationData, nil
}

func getCurrentValues() (*GETResponse, error) {
	id, err := getVotedValue()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get VotedValue")
	}
	pods, err := getPods()
	if err != nil {
		return nil, errors.Wrap(err, "failed to getPods()")
	}

	cip, err := getValue("CURRENT_ILLUMINATION_PATTERN")
	if err != nil {
		return nil, err
	}

	r := GETResponse{
		Pods:                len(pods.Items),
		IlluminationPattern: cip,
		IlluminationData:    *id,
	}
	return &r, nil
}

func getHandler() ([]byte, error) {
	r, err := getCurrentValues()
	if err != nil {
		return nil, errors.Wrapf(err, "failed to call getCurrentValues")
	}
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("cannot marshal struct: %v", err)
		return nil, err
	}
	return b, nil
}

func voteHandler(uid string, votePattern int) ([]byte, error) {
	log.Println(uid, votePattern)
	if 3 < votePattern {
		return nil, errors.New("VotedPattern is invalid")
	}
	err := addValue(voteLabel[votePattern])
	if err != nil {
		return nil, errors.Wrap(err, "failed to addValue")
	}

	currentValue, err := getCurrentValues()
	if err != nil {
		return nil, errors.Wrap(err, "failed to call getCurrentValues")
	}
	r := PostResponse{
		Response:    "OK",
		CurrentData: *currentValue,
	}
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("cannot marshal struct: %v", err)
		return nil, err
	}
	return b, nil
}

func getValue(target string) (int, error) {
	redisPath := os.Getenv("REDIS_PATH")
	client, err := redis.New(redisPath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get redis client")
	}
	defer client.Close()
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

func addValue(target string) error {
	redisPath := os.Getenv("REDIS_PATH")
	client, err := redis.New(redisPath)
	if err != nil {
		return errors.Wrap(err, "failed to get redis client")
	}
	defer client.Close()
	err = client.Get(target).Err()
	if err == redis.Nil {
		log.Printf("%s does not exist. creating now...\n", target)

		err = client.Set(target, 1, time.Hour*24).Err()
		if err != nil {
			return errors.Wrap(err, "failed to set client")
		}
	} else if err != nil {
		return errors.Wrapf(err, "failed to get %s", target)

	} else {
		currentNum, err := client.Incr(target).Result()
		if err != nil {
			return errors.Wrapf(err, "failed to incr %s", target)
		}
		log.Printf("currentNum: %d\n", currentNum)
	}
	return nil
}

func declValue(target string) (int, error){
	redisPath := os.Getenv("REDIS_PATH")
	client, err := redis.New(redisPath)
	if err != nil {
		return 0, errors.Wrap(err, "failed to get redis client")
	}
	defer client.Close()
	currentNum, err := client.Decr(target).Result()
	if err != nil {
		return 0, errors.Wrap(err, "failed to decr CLIENT_NUM")
	}
	return int(currentNum), nil
}