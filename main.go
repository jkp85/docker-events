package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"
)

func main() {
	ws := make(chan *Message)
	defer close(ws)
	go websockets(ws)
	for e := range ECSEvents() {
		args, err := getContainerArgs(e.Command)
		if err != nil {
			log.Println(err)
			continue
		}
		if filterRoot(e, args) {
			ws <- msgFromEvent(e, args)
			switch e.Status {
			case STOPPED:
				go CreateContainerDeleteAction(e, args)
				go EndStats(e, args)
			case RUNNING:
				go AddStats(e, args)
			}
		}
	}
}

func CreateContainerDeleteAction(e *ECSEvent, args *ContainerArgs) {
	log.Printf("Handle server remove action for server: %s\n", args.ServerID)
	action := NewAction("POST", args.ServerID, e.Time)
	uri := fmt.Sprintf("/%s/actions/create/", os.Getenv("TBS_DEFAULT_VERSION"))
	APIClient.HandlePostEvent(args.Key, uri, action)
}

func AddStats(e *ECSEvent, args *ContainerArgs) {
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

func EndStats(e *ECSEvent, args *ContainerArgs) {
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

func filterRoot(e *ECSEvent, args *ContainerArgs) bool {
	serverRoot, _ := url.Parse(args.Root)
	apiRoot, _ := url.Parse(os.Getenv("TBS_HOST"))
	if serverRoot != nil && apiRoot != nil {
		return serverRoot.Hostname() == apiRoot.Hostname()
	}
	// it's better to send it anyway then discard it
	return true
}
