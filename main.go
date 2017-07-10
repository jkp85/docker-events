package main

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/docker/docker/api/types/events"
)

func main() {
	// go websockets()
	d := NewDispatcher()
	d.HandleFunc("service", "remove", Die)
	log.Fatal(d.Run())
}

func Die(e events.Message) {
	name := e.Actor.Attributes["name"]
	if !strings.HasPrefix(name, "server") {
		return
	}
	log.Printf("Handling remove event for service: %s\n", name)
	serverID, err := idFromServerName(name)
	if err != nil {
		log.Println("Error parsing server id: %s", err)
		return
	}
	token, err := getUserToken(name)
	if err != nil {
		log.Println(err)
		return
	}
	action := NewAction("POST", serverID.String())
	resp, err := APIClient.Post("/actions/create/", token, action)
	if err != nil {
		log.Printf("Action create error: %s", err)
	}
	if resp == nil {
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		log.Printf("Error during action create")
		io.Copy(os.Stdout, resp.Body)
	}
}
