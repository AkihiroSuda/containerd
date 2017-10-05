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
	ArgsUsage: "[flags] <in>",
	Description: `Import an image from a tar stream.
Implemented formats:
- oci.v1     (default)
- docker.v1
`,
	Flags: append([]cli.Flag{
		cli.StringFlag{
			Name:  "format",
			Value: "oci.v1",
			Usage: "image format. See DESCRIPTION.",
		},
		cli.StringFlag{
			Name:  "docker.v1-prefix",
			Value: "imported/docker.v1/",
			Usage: "prefix added to docker.v1 RepoTags",
		},
		cli.StringFlag{
			Name:  "oci.v1-prefix",
			Value: "imported/oci.v1/imported:",
			Usage: "prefix added to oci.v1 ref annotation",
		},
		labelFlag,
	}, snapshotterFlags...),

	Action: func(clicontext *cli.Context) error {
		var (
			in            = clicontext.Args().First()
			labels        = labelArgs(clicontext.StringSlice("label"))
			imageImporter imageformats.Importer
		)

		switch format := clicontext.String("format"); format {
		case "oci.v1":
			// TODO(AkihiroSuda): how to import OCI images without ref annotation?
			imageImporter = &oci.V1Importer{
				Prefix: clicontext.String("oci.v1-prefix"),
			}
		case "docker.v1":
			imageImporter = &docker1.Importer{
				Prefix: clicontext.String("docker.v1-prefix"),
			}
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
		imgs, err := client.Import(ctx,
			r,
			containerd.WithImporter(imageImporter),
			containerd.WithImportLabels(labels),
		)
		if err != nil {
			return err
		}
		if err = r.Close(); err != nil {
			return err
		}

		log.G(ctx).Debugf("unpacking %d images", len(imgs))

		for _, img := range imgs {
			// TODO: Show unpack status
			fmt.Printf("unpacking %s...", img.Target().Digest)
			err = img.Unpack(ctx, clicontext.String("snapshotter"))
			if err != nil {
				return err
			}
			fmt.Println("done")
		}
		return nil
	},
}
