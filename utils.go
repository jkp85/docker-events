package main

import (
	"flag"
)

type ContainerArgs struct {
	Key, Namespace, Version, ProjectID, ServerID, Root, Secret, Script, Function, Type string
}

func getContainerArgs(cargs []string) (*ContainerArgs, error) {
	flagSet := flag.NewFlagSet("server", flag.PanicOnError)
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
	flagSet.Parse(cargs[1:])
	return args, nil
}

func sliceConv(sl []*string) []string {
	out := make([]string, len(sl))
	for i, x := range sl {
		out[i] = *x
	}
	return out
}
