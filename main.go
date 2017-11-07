package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types/events"
	uuid "github.com/satori/go.uuid"
)

func main() {
	go websockets()
	d := NewDispatcher()
	d.HandleFunc("container", "die", CreateContainerDeleteAction)
	d.HandleFunc("container", "start", AddStats)
	d.HandleFunc("container", "die", EndStats)
	log.Fatal(d.Run())
}

func CreateContainerDeleteAction(e events.Message) {
	name := e.Actor.Attributes["name"]
	serverID, err := uuid.FromString(name)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Handle server remove action for server: %s\n", serverID)
	action := NewAction("POST", serverID.String())
	uri := fmt.Sprintf("/%s/actions/create/", os.Getenv("TBS_DEFAULT_VERSION"))
	APIClient.HandlePostEvent(e, uri, action)
}

func AddStats(e events.Message) {
	name := e.Actor.Attributes["name"]
	serverID, err := uuid.FromString(name)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Creates server stats: %s\n", name)
	args, err := getContainerArgs(name)
	if err != nil {
		log.Println(err)
		return
	}
	uri := fmt.Sprintf(
		"/%s/%s/projects/%s/servers/%s/run-stats/",
		args.Version,
		args.Namespace,
		args.ProjectID,
		serverID,
	)
	stats := NewStats()
	APIClient.HandlePostEvent(e, uri, stats)
}

func EndStats(e events.Message) {
	name := e.Actor.Attributes["name"]
	serverID, err := uuid.FromString(name)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Creates server stats: %s\n", name)
	args, err := getContainerArgs(name)
	if err != nil {
		log.Println(err)
		return
	}
	uri := fmt.Sprintf(
		"/%s/%s/projects/%s/servers/%s/run-stats/update_latest/",
		args.Version,
		args.Namespace,
		args.ProjectID,
		serverID,
	)
	stats := &struct {
		Stop time.Time `json:"stop"`
	}{
		time.Now().UTC(),
	}
	APIClient.HandlePostEvent(e, uri, stats)
}
