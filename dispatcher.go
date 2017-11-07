package main

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

// EventHandler interface for handling received events
type EventHandler interface {
	Handle(events.Message)
}

// EventHandlerFunc is a  function for convenience
type EventHandlerFunc func(events.Message)

// Handle implements EventHandler interface
func (f EventHandlerFunc) Handle(e events.Message) {
	f(e)
}

// Dispatcher is an interface for dispatching actions
type Dispatcher interface {
	Run() error
	Handle(eventType, name string, eh EventHandler)
	HandleFunc(eventType, name string, eh func(events.Message))
}

// Dispatch fire events to proper handlers
type Dispatch struct {
	handlers map[string][]EventHandler
}

// NewDispatcher creates new dispatcher
func NewDispatcher() Dispatcher {
	return &Dispatch{
		handlers: make(map[string][]EventHandler),
	}
}

// Run is starting a dispatcher
func (d *Dispatch) Run() error {
	log.Println("Starting dispatcher...")
	cli, err := client.NewEnvClient()
	if err != nil {
		return err
	}
	eventsCh, errCh := cli.Events(context.Background(), types.EventsOptions{})
	log.Printf("Dispatcher: %s\n", d)
	for {
		select {
		case msg := <-eventsCh:
			eventName := fmt.Sprintf("%s.%s", msg.Type, msg.Action)
			handlers, ok := d.handlers[eventName]
			if ok {
				for _, handler := range handlers {
					go handler.Handle(msg)
				}
			}
		case err = <-errCh:
			if err != nil {
				return err
			}
		}
	}
}

// Handle registers event with proper handler
// eventType should be either container or image or volume or network or daemon
func (d *Dispatch) Handle(eventType, name string, handler EventHandler) {
	if validateEvent(eventType, name) {
		d.appendHandler(eventType, name, handler)
	}
}

// HandleFunc registers event with proper handle function
func (d *Dispatch) HandleFunc(eventType, name string, handler func(events.Message)) {
	if validateEvent(eventType, name) {
		d.appendHandler(eventType, name, EventHandlerFunc(handler))
	}
}

func (d *Dispatch) appendHandler(eventType, name string, handler EventHandler) {
	eventName := fmt.Sprintf("%s.%s", eventType, name)
	if _, ok := d.handlers[eventName]; !ok {
		d.handlers[eventName] = []EventHandler{}
	}
	d.handlers[eventName] = append(d.handlers[eventName], handler)
}
