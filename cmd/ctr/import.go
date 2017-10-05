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
	Name:      "import",
	Usage:     "import an image",
	ArgsUsage: "[flags] <ref> <in>",
	Description: `Import an image from a tar stream.
Implemented formats:
- oci.v1     (default)

Planned but unimplemented formats:
- docker.v1

Selector string specification:
- oci.v1:    containerd reference string (e.g. example.com/foo/bar:tag@digest) or reference object string (e.g. tag@digest).
             Locator part in reference string is ignored.
- docker.v1: Docker RepoTag string (e.g. example.com/foo/bar:tag)
             Unlike containerd ref string, Docker RepoTag string MUST not contain digest. So don't put digest in selector.
             Selector needs to contain locator part at the moment. (e.g. "docker.io/library/busybox:latest" is correct but "busybox:latest" not)

The default selector string is <ref>, but it is not guaranteed to be a valid selector string, depending on the importer implementation.
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "selector",
			Value: "",
			Usage: "string for selecting which image object to import from the archive stream. See DESCRIPTION.",
		},
		labelFlag,
	},

	Action: func(clicontext *cli.Context) error {
		var (
			ref      = clicontext.Args().First()
			in       = clicontext.Args().Get(1)
			selector = clicontext.String("selector")
			labels   = labelArgs(clicontext.StringSlice("label"))
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
			ref,
			r,
			containerd.WithImportSelector(selector),
			containerd.WithImportLabels(labels),
		)
		if err != nil {
			return err
		}
		if err = r.Close(); err != nil {
			return err
		}

		log.G(ctx).WithField("image", ref).Debug("unpacking")

		// TODO: Show unpack status
		fmt.Printf("unpacking %s...", img.Target().Digest)
		err = img.Unpack(ctx, clicontext.String("snapshotter"))
		fmt.Println("done")
		return err
	},
}
