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
	GetHandler(string) EventHandler
}

// Dispatches events to proper handlers
type Dispatch struct {
	handlers map[string]EventHandler
}

// Gets event handler by name
func (d *Dispatch) GetHandler(name string) EventHandler {
	handler, ok := d.handlers[name]
	if ok {
		return handler
	}
	return nil
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
	log.Printf("Dispatcher: %s\n", d)
	return dispatch(cli, d)
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

func listen(cli *client.Client) (<-chan events.Message, <-chan error) {
	return cli.Events(context.Background(), types.EventsOptions{})
}

func dispatch(cli *client.Client, d Dispatcher) error {
	for {
		eventsCh, errCh := listen(cli)
		select {
		case msg := <-eventsCh:
			eventName := fmt.Sprintf("%s.%s", msg.Type, msg.Action)
			handler := d.GetHandler(eventName)
			if handler != nil {
				go handler.Handle(msg)
			}
		case err := <-errCh:
			return err
		}
	}
	return nil
}

func eventName(msg events.Message) string {
	return fmt.Sprintf("%s.%s", msg.Type, msg.Action)
}
