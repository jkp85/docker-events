package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
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
	eventsCh, errCh := cli.Events(context.Background(), types.EventsOptions{
		Filters: filter,
	})
	select {
	case msg := <-eventsCh:
		eventName := msg.Status
		exitCode := msg.Actor.Attributes["exitCode"]
		apiMsg := new(Message)
		if eventName == "start" {
			apiMsg.Status = "Running"
		} else {
			switch exitCode {
			case "0", "143":
				apiMsg.Status = "Stopped"
			default:
				apiMsg.Status = "Error"
			}
		}
		err = conn.WriteJSON(apiMsg)
		if err != nil {
			log.Println(err)
		}
	case err = <-errCh:
		log.Println(err)
	}
}
