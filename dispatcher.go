package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// Event handler interface for handling received events
type EventHandler interface {
	Handle(events.Message)
}

// Event handler function for convenience
type EventHandlerFunc func(events.Message)

func (f EventHandlerFunc) Handle(e events.Message) {
	f(e)
}

type Dispatcher interface {
	Run() error
	Handle(eventType, name string, eh EventHandler)
	HandleFunc(eventType, name string, eh func(events.Message))
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
	eventsCh, errCh := cli.Events(context.Background(), types.EventsOptions{})
	if err != nil {
		return err
	}
	log.Printf("Dispatcher: %s\n", d)
	select {
	case msg := <-eventsCh:
		eventName := fmt.Sprintf("%s.%s", msg.Type, msg.Action)
		handler, ok := d.handlers[eventName]
		if ok {
			go handler.Handle(msg)
		}
	case err = <-errCh:
		return err
	}
	return nil
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
func (d *Dispatch) HandleFunc(eventType, name string, handler func(events.Message)) {
	if validateEvent(eventType, name) {
		eventName := fmt.Sprintf("%s.%s", eventType, name)
		d.handlers[eventName] = EventHandlerFunc(handler)
	}
}
