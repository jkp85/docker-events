package main

import (
	"context"
	"flag"

	"github.com/docker/docker/client"
)

type ContainerArgs struct {
	Key, Namespace, Version, ProjectID, ServerID, Root, Secret, Script, Function, Type string
}

func getContainerArgs(container string) (*ContainerArgs, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	contJSON, _, err := cli.ContainerInspectWithRaw(context.Background(), container, false)
	if err != nil {
		return nil, err
	}
	flagSet := flag.NewFlagSet("server", flag.ContinueOnError)
	args := new(ContainerArgs)
	flagSet.StringVar(&args.Key, "key", "", "")
	flagSet.StringVar(&args.Namespace, "ns", "", "")
	flagSet.StringVar(&args.Version, "version", "", "")
	flagSet.StringVar(&args.ProjectID, "projectID", "", "")
	flagSet.StringVar(&args.ServerID, "serverID", "", "")
	flagSet.StringVar(&args.Root, "root", "", "")
	flagSet.StringVar(&args.Secret, "secret", "", "")
	flagSet.StringVar(&args.Script, "script", "", "")
	flagSet.StringVar(&args.Function, "function", "", "")
	flagSet.StringVar(&args.Type, "type", "", "")
	flagSet.Parse(contJSON.Args[2:])
	return args, nil
}
