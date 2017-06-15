package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/log"
	"github.com/urfave/cli"
)

var importCommand = cli.Command{
	Name:      "import",
	Usage:     "import an image",
	ArgsUsage: "[flags] <ref> <in>",
	Description: `Import an image from a tar stream or a directory
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "ref-object",
			Value: "",
			Usage: "reference object e.g. tag@digest (default: use the object specified in ref)",
		},
	},
	Action: func(clicontext *cli.Context) error {
		var (
			ref       = clicontext.Args().First()
			in        = clicontext.Args().Get(1)
			refObject = clicontext.String("ref-object")
		)

		ctx, cancel := appContext(clicontext)
		defer cancel()

		client, err := getClient(clicontext)
		if err != nil {
			return err
		}

		var img containerd.Image
		if isExistingDirectory(in) {
			img, err = client.ImportDirectory(ctx,
				ref,
				func(relPath string) (io.ReadCloser, error) { return os.Open(filepath.Join(in, relPath)) },
				containerd.WithRefObject(refObject),
			)
			if err != nil {
				return err
			}
		} else {
			var r io.ReadCloser
			if in == "-" {
				r = os.Stdin
			} else {
				r, err = os.Open(in)
				if err != nil {
					return err
				}
			}
			img, err = client.Import(ctx,
				ref,
				r,
				containerd.WithRefObject(refObject),
			)
			if err != nil {
				return err
			}
			if err = r.Close(); err != nil {
				return err
			}
		}

		log.G(ctx).WithField("image", ref).Debug("unpacking")

		// TODO: Show unpack status
		fmt.Printf("unpacking %s...", img.Target().Digest)
		err = img.Unpack(ctx)
		fmt.Println("done")
		return err
	},
}

func isExistingDirectory(path string) bool {
	path, err := filepath.EvalSymlinks(path)
	if err != nil {
		return false
	}
	fi, err := os.Stat(path)
	if err != nil {
		return false
	}
	return fi.IsDir()
}
