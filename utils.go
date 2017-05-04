package main

import (
	"strings"

	uuid "github.com/satori/go.uuid"
)

func idFromServerName(name string) (uuid.UUID, error) {
	serverID := strings.Split(name, "_")[1]
	return uuid.FromString(serverID)
}
