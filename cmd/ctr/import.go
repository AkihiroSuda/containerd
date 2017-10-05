package main

import (
	"fmt"
	"io"
	"os"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/contrib/docker1"
	"github.com/containerd/containerd/imageformats"
	oci "github.com/containerd/containerd/imageformats/oci"
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
- docker.v1

Supported filters:
- oci.v1:
-- index:        manifest list index (e.g. "0")
-- tag:          org.opencontainers.image.ref.name tag (e.g. "latest")
-- digest:       manifest digest (e.g. sha256:deadbeefdeadbeef)
-- mediaType:    manifest media type ("e.g. application/vnd.oci.image.manifest.v1+json")
-- architecture: e.g. "amd64"
-- os:           e.g. "linux"
-- variant:      e.g. "v7" (when architecture is "arm")

- docker.v1:
-- index:        manifest list index (e.g. "0")

Specifying filter is mandatory if an image contains multiple objects.
If filtered result contains multiple items, it results in ErrAmbiguous.
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "format",
			Value: "oci.v1",
			Usage: "image format. See DESCRIPTION.",
		},
		cli.StringFlag{
			Name:  "filter",
			Value: "",
			Usage: "string for selecting which image object to import from the archive stream. See DESCRIPTION.",
		},
		labelFlag,
	},

	Action: func(clicontext *cli.Context) error {
		var (
			ref           = clicontext.Args().First()
			in            = clicontext.Args().Get(1)
			filter        = clicontext.String("filter")
			labels        = labelArgs(clicontext.StringSlice("label"))
			imageImporter imageformats.Importer
		)

		switch format := clicontext.String("format"); format {
		case "oci.v1":
			imageImporter = oci.V1Importer
		case "docker.v1":
			imageImporter = docker1.Importer
		default:
			return fmt.Errorf("unknown format %s", format)
		}

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
			filter,
			containerd.WithImporter(imageImporter),
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
