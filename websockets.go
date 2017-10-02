package main

import (
	"bufio"
	"context"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func websockets() {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Fatal(err)
	}
	r := mux.NewRouter()
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/status/", statusHandler(cli))
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/logs/", logsHandler(cli))
	server := &http.Server{
		Addr:           ":8000",
		Handler:        r,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	log.Fatal(server.ListenAndServe())
}

type Message struct {
	Status string `json:"status"`
}

func logsHandler(cli *client.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		server := vars["serverID"]
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
		var logs io.ReadCloser
		var scanner *bufio.Scanner
		logs, err = getLogs(cli, server, true)
		scanner = bufio.NewScanner(logs)
		for {
			for scanner.Scan() {
				conn.WriteMessage(websocket.BinaryMessage, scanner.Bytes())
			}
			if scanner.Err() != nil {
				log.Printf("Scanner err: %s\n", scanner.Err())
			}
			logs.Close()
			waitForServer(cli, server)
			logs, err = getLogs(cli, server, false)
			scanner = bufio.NewScanner(logs)
		}
	})
}

func waitForServer(cli *client.Client, server string) {
	filter := filters.NewArgs()
	filter.Add("container", server)
	filter.Add("event", "start")
	eventsCh, _ := cli.Events(context.Background(), types.EventsOptions{
		Filters: filter,
	})
	log.Printf("Waiting for server: %s\n", server)
	<-eventsCh
}

func getLogs(cli *client.Client, server string, tail bool) (io.ReadCloser, error) {
	log.Printf("Getting logs for server: %s\n", server)
	opts := types.ContainerLogsOptions{
		Follow:     true,
		ShowStderr: true,
		ShowStdout: true,
	}
	if tail {
		opts.Tail = "all"
	}
	return cli.ContainerLogs(context.Background(), server, opts)
}

func statusHandler(cli *client.Client) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		server := vars["serverID"]
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
		filter := filters.NewArgs()
		filter.Add("container", server)
		filter.Add("event", "start")
		filter.Add("event", "die")
		eventsCh, errCh := cli.Events(context.Background(), types.EventsOptions{
			Filters: filter,
		})
		for {
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
	})
}
