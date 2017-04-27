package main

import (
	gocontext "context"
	"fmt"
	"strings"

	"github.com/containerd/containerd/api/services/introspection"
	"github.com/urfave/cli"
)

var introspectCommand = cli.Command{
	Name:      "introspect",
	Usage:     "introspect the daemon",
	ArgsUsage: "[TARGET]",
	Action: func(context *cli.Context) error {
		if context.NArg() > 2 {
			return fmt.Errorf("number of args expected to be <= 1, got %d", context.NArg())
		}
		intro, err := getIntrospection(context)
		if err != nil {
			return err
		}
		ctx := gocontext.Background()
		var response interface{}
		switch target := strings.ToLower(context.Args().First()); target {
		case "version":
			response, err = intro.IntrospectVersion(ctx, &introspection.IntrospectVersionRequest{})
		case "all", "":
			response, err = intro.IntrospectAll(ctx, &introspection.IntrospectAllRequest{})
		default:
			err = fmt.Errorf("unknown target: %s (must be \"all\" or \"version\")", target)
		}
		if err != nil {
			return err
		}
		getSpewDumper().Dump(response)
		return nil
	},
}
