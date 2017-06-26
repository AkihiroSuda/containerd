package main

import (
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/log"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var importCommand = cli.Command{
	Name:      "import",
	Usage:     "import an image",
	ArgsUsage: "[flags] <ref> <in>",
	Description: `Import an image
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "representation",
			Value: "oci+directory",
			Usage: "representation (current supported formats: oci+directory, oci+tar)",
		},
	},
	Action: func(clicontext *cli.Context) error {
		var (
			ref = clicontext.Args().First()
			in  = clicontext.Args().Get(1)
		)

		ctx, cancel := appContext(clicontext)
		defer cancel()

		client, err := getClient(clicontext)
		if err != nil {
			return err
		}

		var iopt containerd.ImportOpt
		switch r := clicontext.String("representation"); r {
		case "oci+directory":
			iopt = containerd.WithOCIDirectoryImportation(in)
		case "oci+tar":
			var r io.Reader
			if in == "-" {
				r = os.Stdin
			} else {
				r, err = os.Open(in)
				if err != nil {
					return nil
				}
			}
			iopt = containerd.WithOCITarImportation(r)
		default:
			return errors.Errorf("unsupported representation: %s", r)

		}
		img, err := client.Import(ctx,
			ref,
			iopt,
		)
		if err != nil {
			return err
		}

		log.G(ctx).WithField("image", ref).Debug("unpacking")

		// TODO: Show unpack status
		fmt.Printf("unpacking %s...", img.Target().Digest)
		err = img.Unpack(ctx)
		fmt.Println("done")
		return err
	},
}
