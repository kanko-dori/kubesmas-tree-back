package main

import (
	"encoding/json"
	"fmt"
	"log"
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

var errorResponse = []byte("ERROR")

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

func getHandler() ([]byte, error) {
	id := IlluminationData{
		Pattern1: 20,
		Pattern2: 40,
		Pattern3: 30,
		Pattern4: 10,
	}
	r := GETResponse{
		Pods:                10,
		IlluminationPattern: 1,
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
	id := IlluminationData{
		Pattern1: 20,
		Pattern2: 40,
		Pattern3: 30,
		Pattern4: 10,
	}
	gr := GETResponse{
		Pods:                10,
		IlluminationPattern: 1,
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
