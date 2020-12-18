package main

import (
	"context"
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/rand"
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
	Pattern4 int `json:"pattern4"`
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

func handler(s []byte) []byte {
	var r Request
	if err := json.Unmarshal(s, &r); err != nil {
		log.Println("cannot unmarshal request: %v, err: %v\n", s, err)
		return errorResponse
	}
	fmt.Printf("%v\n", r)

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

func getPods() (*v1.PodList, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	pods, err := clientset.CoreV1().Pods("kubesmas-tree").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	return pods, nil
}

func getHandler() ([]byte, error) {
	n := rand.Intn(5) + 1
	ns := rand.Intn(5) + 1
	id := IlluminationData{
		Pattern1: n * 5,
		Pattern2: (10 - n) * 5,
		Pattern3: ns * 5,
		Pattern4: (10 - ns) * 5,
	}
	pods, err := getPods()
	if err != nil {
		log.Println("failed to getPods(): %v", err)
	}
	r := GETResponse{
		Pods:                len(pods.Items),
		IlluminationPattern: rand.Intn(4),
		IlluminationData:    id,
	}
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("cannot marshal struct: %v", err)
		return nil, err
	}
	return b, nil
}

func voteHandler(uid string, votedPattern int) ([]byte, error) {
	log.Println(uid, votedPattern)
	n := rand.Intn(5) + 1
	ns := rand.Intn(5) + 1
	id := IlluminationData{
		Pattern1: n * 5,
		Pattern2: (10 - n) * 5,
		Pattern3: ns * 5,
		Pattern4: (10 - ns) * 5,
	}
	pods, err := getPods()
	if err != nil {
		log.Println("failed to getPods(): %v", err)
	}
	gr := GETResponse{
		Pods:                len(pods.Items),
		IlluminationPattern: rand.Intn(4),
		IlluminationData:    id,
	}
	r := PostResponse{
		Response:    "OK",
		CurrentData: gr,
	}
	b, err := json.Marshal(r)
	if err != nil {
		log.Println("cannot marshal struct: %v", err)
		return nil, err
	}
	return b, nil
}
