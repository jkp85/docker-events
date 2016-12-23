package main

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/garyburd/redigo/redis"
	"github.com/speps/go-hashids"
)

type ServerId int

func (s ServerId) Hash() string {
	hashId, err := encodeHashID(int(s))
	if err != nil {
		return ""
	}
	return hashId
}


func sendNotification(id ServerId) {
	var body bytes.Buffer
	url := fmt.Sprintf("http://%s/notifications/servers/%s/update",
		os.Getenv("REALTIME_NOTIFICATIONS_SERVER"), id.Hash())
	resp, err := http.Post(url, "application/json", &body)
	if err != nil {
		log.Printf("Notifications request error: %s\n", err)
	}
	if resp.StatusCode != 200 {
		var buf bytes.Buffer
		buf.ReadFrom(resp.Body)
		log.Printf("Request status error: %s\n", buf.String())
	}
}

func setCache(id ServerId) {
	conn, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Printf("Redis connection error: %s", err)
		return
	}
	defer conn.Close()
	key := fmt.Sprintf("server_state_%s", id.Hash())
	_, err = conn.Do("HSET", key, "update", "We launched new version of this server. Click to update.")
	if err != nil {
		log.Printf("Set status error: %s", err)
	}
}

func checkIfSent(ServerHashId string) bool {
	conn, err := redis.DialURL(os.Getenv("REDIS_URL"))
	if err != nil {
		log.Printf("Redis connection error: %s", err)
		return false
	}
	defer conn.Close()
	key := fmt.Sprintf("server_state_%s", ServerHashId)
	resp, err := redis.Bool(conn.Do("HEXISTS", key, "update"))
	if err != nil {
		return false
	}
	return resp
}

func decodeHashID(hashid string) (int, error) {
	hd := hashids.NewData()
	hd.MinLength = 8
	hd.Salt = "70fc9c91-a6dc-4c1b-a494-b16d7f3b3ce7"
	h := hashids.NewWithData(hd)
	ids, err := h.DecodeWithError(hashid)
	if err != nil {
		return 0, err
	}
	if len(ids) == 0 {
		return 0, fmt.Errorf("No id for hash: %s\n", hashid)
	}
	return ids[0], nil
}

func encodeHashID(id int) (string, error) {
	hd := hashids.NewData()
	hd.MinLength = 8
	hd.Salt = "70fc9c91-a6dc-4c1b-a494-b16d7f3b3ce7"
	h := hashids.NewWithData(hd)
	e, err := h.Encode([]int{id})
	if err != nil {
		return "", err
	}
	return e, nil
}
