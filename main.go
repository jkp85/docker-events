package main

import (
	"fmt"
	"log"
	"os"
	"time"
)

func main() {
	for e := range ECSEvents() {
		switch e.Status {
		case STOPPED:
			go CreateContainerDeleteAction(e)
			go EndStats(e)
		case RUNNING:
			go AddStats(e)
		}
	}
}

func CreateContainerDeleteAction(e *ECSEvent) {
	args, err := getContainerArgs(e.Command)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Handle server remove action for server: %s\n", args.ServerID)
	action := NewAction("POST", args.ServerID, e.Time)
	uri := fmt.Sprintf("/%s/actions/create/", os.Getenv("TBS_DEFAULT_VERSION"))
	APIClient.HandlePostEvent(args.Key, uri, action)
}

func AddStats(e *ECSEvent) {
	args, err := getContainerArgs(e.Command)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Creates server stats: %s\n", args.ServerID)
	uri := fmt.Sprintf(
		"/%s/%s/projects/%s/servers/%s/run-stats/",
		os.Getenv("TBS_DEFAULT_VERSION"),
		args.Namespace,
		args.ProjectID,
		args.ServerID,
	)
	stats := NewStats()
	APIClient.HandlePostEvent(args.Key, uri, stats)
}

func EndStats(e *ECSEvent) {
	args, err := getContainerArgs(e.Command)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Creates server stats: %s\n", args.ServerID)
	uri := fmt.Sprintf(
		"/%s/%s/projects/%s/servers/%s/run-stats/update_latest/",
		os.Getenv("TBS_DEFAULT_VERSION"),
		args.Namespace,
		args.ProjectID,
		args.ServerID,
	)
	stats := &struct {
		Stop time.Time `json:"stop"`
	}{
		time.Now().UTC(),
	}
	APIClient.HandlePostEvent(args.Key, uri, stats)
}
