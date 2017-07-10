package main

import (
	"context"
	"errors"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	uuid "github.com/satori/go.uuid"
)

func idFromServerName(name string) (uuid.UUID, error) {
	serverID := strings.Split(name, "_")[1]
	return uuid.FromString(serverID)
}

func getUserToken(service string) (string, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return "", err
	}
	serviceJSON, _, err := cli.ServiceInspectWithRaw(context.Background(), service, types.ServiceInspectOptions{})
	args := serviceJSON.Spec.TaskTemplate.ContainerSpec.Args
	for _, arg := range args {
		if strings.HasPrefix(arg, "--key") {
			return strings.Split(arg, "=")[1], nil
		}
	}
	return "", errors.New("No user token.")
}
