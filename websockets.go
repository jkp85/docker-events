package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type StatusDispatcher struct {
	Consumers map[string]chan *Message
	Messages  <-chan *Message
}

func (sd *StatusDispatcher) Dispatch() {
	for msg := range sd.Messages {
		if con, ok := sd.Consumers[msg.ServerID]; ok {
			con <- msg
		}
	}
}

func (sd *StatusDispatcher) AddConsumer(serverID string) chan *Message {
	out := make(chan *Message)
	sd.Consumers[serverID] = out
	return out
}

func (sd *StatusDispatcher) RemoveConsumer(serverID string) {
	close(sd.Consumers[serverID])
	delete(sd.Consumers, serverID)
}

func websockets(statuses <-chan *Message) {
	sd := &StatusDispatcher{Messages: statuses, Consumers: make(map[string]chan *Message)}
	go sd.Dispatch()
	r := mux.NewRouter()
	r.Handle("/{version}/{namespace}/projects/{projectID}/servers/{serverID}/status/", statusHandler(sd))
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
	ServerID string `json:"-"`
	Status   string `json:"status"`
}

func statusHandler(sd *StatusDispatcher) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		server := vars["serverID"]
		if server == "" {
			http.NotFound(w, r)
			return
		}
		log.Printf("Handling server: %s\n", server)
		statuses := sd.AddConsumer(server)
		defer sd.RemoveConsumer(server)
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			return
		}
		defer conn.Close()
		for {
			select {
			case msg := <-statuses:
				log.Printf("Status update: %+v\n", msg)
				err = conn.WriteJSON(msg)
				if err != nil {
					log.Println(err)
				}
			}
		}
	})
}

func msgFromEvent(e *ECSEvent, args *ContainerArgs) *Message {
	return &Message{Status: e.Status, ServerID: args.ServerID}
}
