package main

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/client"
	uuid "github.com/satori/go.uuid"
)

func idFromServerName(name string) (uuid.UUID, error) {
	serverID := strings.Split(name, "_")[1]
	return uuid.FromString(serverID)
}

func getUserToken(container string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}
	containerJSON, err := cli.ContainerInspect(context.Background(), container)
	for _, arg := range containerJSON.Args {
		if strings.HasPrefix(arg, "-key") {
			return strings.Split(arg, "=")[1], nil
		}
	}
	return "", errors.New("No user token.")
}
