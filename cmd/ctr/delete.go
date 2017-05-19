package main

import (
	gocontext "context"
	"fmt"

	"github.com/containerd/containerd/api/services/execution"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var deleteCommand = cli.Command{
	Name:      "delete",
	Usage:     "delete an existing container",
	ArgsUsage: "CONTAINER",
	Action: func(context *cli.Context) error {
		containers, err := getExecutionService(context)
		if err != nil {
			return err
		}
		id := context.Args().First()
		if id == "" {
			return errors.New("container id must be provided")
		}
		ctx := gocontext.TODO()
		_, err = containers.Delete(ctx, &execution.DeleteRequest{
			ID: id,
		})
		if err != nil {
			return errors.Wrap(err, "failed to delete container")
		}

		fmt.Printf("Please remove snapshot %q manually (TODO: automatically introspect the current snapshotter and remove the snapshot accordingly)", id)

		return nil
	},
}
