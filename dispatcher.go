package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/docker/engine-api/client"
	"github.com/docker/engine-api/types"
)

// Event response
type Event map[string]interface{}

// Event handler interface for handling received events
type EventHandler interface {
	Handle(Event)
}

// Event handler function for convenience
type EventHandlerFunc func(Event)

func (f EventHandlerFunc) Handle(e Event) {
	f(e)
}

type Dispatcher interface {
	Run() error
	Handle(eventType, name string, eh EventHandler)
	HandleFunc(eventType, name string, eh func(Event))
}

// Dispatches events to proper handlers
type Dispatch struct {
	handlers map[string]EventHandler
}

// Creates new dispatcher
func NewDispatcher() Dispatcher {
	return &Dispatch{
		handlers: make(map[string]EventHandler),
	}
}

// Runs dispatcher
func (d *Dispatch) Run() error {
	log.Println("Starting dispatcher...")
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	events, err := cli.Events(context.Background(), types.EventsOptions{})
	if err != nil {
		return err
	}
	defer events.Close()
	dec := json.NewDecoder(events)
	log.Printf("Dispatcher: %s\n", d)
	for {
		event := make(Event)
		err := dec.Decode(&event)
		if err != nil {
			return err
		}
		eventName := fmt.Sprintf("%s.%s", event["Type"], event["Action"])
		handler, ok := d.handlers[eventName]
		if ok {
			go handler.Handle(event)
		}
	}
}

// Registers event with proper handler
// eventType should be either container or image or volume or network or daemon
func (d *Dispatch) Handle(eventType, name string, handler EventHandler) {
	if validateEvent(eventType, name) {
		eventName := fmt.Sprintf("%s.%s", eventType, name)
		d.handlers[eventName] = handler
	}
}

// Registers event with proper handle function
func (d *Dispatch) HandleFunc(eventType, name string, handler func(Event)) {
	if validateEvent(eventType, name) {
		eventName := fmt.Sprintf("%s.%s", eventType, name)
		d.handlers[eventName] = EventHandlerFunc(handler)
	}
}
