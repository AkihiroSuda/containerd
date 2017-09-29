package main

import (
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/log"
	"github.com/urfave/cli"
)

var imagesImportCommand = cli.Command{
	Name:        "import",
	Usage:       "import an image",
	ArgsUsage:   "[flags] <name> <in>",
	Description: `Import an image from a tar stream.`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "ref-object",
			Value: "",
			Usage: "reference object e.g. tag@digest (default: use the object specified in the image name)",
		},
		labelFlag,
	},
	Action: func(clicontext *cli.Context) error {
		var (
			name      = clicontext.Args().First()
			in        = clicontext.Args().Get(1)
			refObject = clicontext.String("ref-object")
			labels    = labelArgs(clicontext.StringSlice("label"))
		)

		ctx, cancel := appContext(clicontext)
		defer cancel()

		client, err := newClient(clicontext)
		if err != nil {
			return err
		}

		var r io.ReadCloser
		if in == "-" {
			r = os.Stdin
		} else {
			r, err = os.Open(in)
			if err != nil {
				return err
			}
		}
		img, err := client.Import(ctx,
			name,
			r,
			containerd.WithRefObject(refObject),
			containerd.WithImportLabels(labels),
		)
		if err != nil {
			return err
		}
		if err = r.Close(); err != nil {
			return err
		}

		log.G(ctx).WithField("image", name).Debug("unpacking")

		// TODO: Show unpack status
		fmt.Printf("unpacking %s...", img.Target().Digest)
		err = img.Unpack(ctx, clicontext.String("snapshotter"))
		fmt.Println("done")
		return err
	},
}
