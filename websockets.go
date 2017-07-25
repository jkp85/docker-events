package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/gorilla/mux"
	ws "golang.org/x/net/websocket"
)

func handshake(conf *ws.Config, r *http.Request) error {
	r.Header.Set("Access-Control-Allow-Origin", "*")
	return nil
}

func websockets() {
	r := mux.NewRouter()
	r.HandleFunc("/status/{server}/", statusHandler)
	r.Handle("/logs/{server}/", ws.Handler(logsHandler))
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

func statusHandler(w http.ResponseWriter, r *http.Request) {
	upgrade := w.Header().Get("Upgrade")
	if upgrade == "websocket" {
		ws.Handler(wsStatusHandler).ServeHTTP(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	client := &http.Client{}
	vars := mux.Vars(r)
	uri := fmt.Sprintf("http://%s/api", os.Getenv("DOCKER_DOMAIN"))
	msg := &Message{}
	enc := json.NewEncoder(w)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		log.Println(err)
		msg.Status = "Error"
		enc.Encode(msg)
		return
	}
	req.Host = fmt.Sprintf("%s.traefik", vars["server"])
	_, err = client.Do(req)
	if err != nil {
		log.Println(err)
		msg.Status = "Error"
	} else {
		msg.Status = "Running"
	}
	enc.Encode(msg)
}

func wsStatusHandler(conn *ws.Conn) {
	vars := mux.Vars(conn.Request())
	server := vars["server"]
	defer conn.Close()
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println(err)
		return
	}
	filter := filters.NewArgs()
	filter.Add("service", server)
	filter.Add("event", "create")
	filter.Add("event", "remove")
	eventsCh, errCh := cli.Events(context.Background(), types.EventsOptions{
		Filters: filter,
	})
	enc := json.NewEncoder(conn)
	for {
		select {
		case msg := <-eventsCh:
			eventName := msg.Action
			apiMsg := &Message{}
			if eventName == "remove" {
				apiMsg.Status = "Stopped"
				err = enc.Encode(apiMsg)
				if err != nil {
					log.Println(err)
				}
			}
		case err = <-errCh:
			if err == io.EOF {
				break
			}
			log.Println(err)
		}
	}
}

func logsHandler(conn *ws.Conn) {
	conn.PayloadType = ws.BinaryFrame
	defer conn.Close()
	vars := mux.Vars(conn.Request())
	server := vars["server"]
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Println(err)
		return
	}
	read, err := cli.ServiceLogs(context.Background(), server, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	if err != nil {
		log.Println(err)
		return
	}
	_, err = io.Copy(conn, read)
	if err != nil {
		log.Println(err)
	}
}
