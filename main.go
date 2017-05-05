package main

import (
	"io"
	"log"
	"os"
	"strings"
)

func main() {
	d := NewDispatcher()
	d.HandleFunc("container", "die", Die)
	log.Fatal(d.Run())
}

func Die(e Event) {
	attrs := e["Actor"].(map[string]interface{})["Attributes"].(map[string]interface{})
	name := attrs["name"].(string)
	if !strings.HasPrefix(name, "server") {
		return
	}
	log.Printf("Handling die event for container: %s\n", name)
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
