package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
	"github.com/docker/engine-api/types/filters"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func websockets() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serverHandler)
	server := &http.Server{
		Addr:           ":8000",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}

type Message struct {
	Status string `json:"status"`
}

func serverHandler(w http.ResponseWriter, r *http.Request) {
	server := r.URL.Path[1:]
	if server == "" {
		http.NotFound(w, r)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println(err)
		return
	}
	filter := filters.NewArgs()
	filter.Add("container", server)
	filter.Add("event", "start")
	filter.Add("event", "die")
	events, err := cli.Events(context.Background(), types.EventsOptions{
		Filters: filter,
	})
	if err != nil {
		log.Println(err)
		return
	}
	defer events.Close()
	dec := json.NewDecoder(events)
	for {
		event := make(Event)
		err := dec.Decode(&event)
		if err != nil {
			log.Println(err)
		}
		eventName := event["status"].(string)
		exitCode, ok := event["Actor"].(map[string]interface{})["Attributes"].(map[string]interface{})["exitCode"].(string)
		if !ok {
			exitCode = ""
		}
		msg := new(Message)
		if eventName == "start" {
			msg.Status = "Running"
		} else {
			if exitCode != "0" {
				msg.Status = "Error"
			} else {
				msg.Status = "Stopped"
			}
		}
		err = conn.WriteJSON(msg)
		if err != nil {
			log.Println(err)
		}
	}
}
