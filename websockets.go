package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
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
	r.Handle("/status/{server}", &ws.Server{Handler: statusHandler, Handshake: handshake})
	r.Handle("/logs/{server}", &ws.Server{Handler: logsHandler, Handshake: handshake})
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

func statusHandler(conn *ws.Conn) {
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
			apiMsg := &Message{Status: "Running"}
			if eventName == "remove" {
				apiMsg.Status = "Stopped"
			}
			err = enc.Encode(apiMsg)
			if err != nil {
				log.Println(err)
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
