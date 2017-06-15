package main

import (
	"github.com/containerd/containerd"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
)

var exportCommand = cli.Command{
	Name:      "export",
	Usage:     "export an image",
	ArgsUsage: "[flags] <directory> <local>",
	Description: `Export an image
`,
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "manifest",
			Usage: "Digest of manifest",
		},
		cli.StringFlag{
			Name:  "manifest-type",
			Usage: "Media type of manifest digest",
			Value: ocispec.MediaTypeImageManifest,
		},
	},
	Action: func(clicontext *cli.Context) error {
		var (
			dir   = clicontext.Args().First()
			local = clicontext.Args().Get(1)
			desc  ocispec.Descriptor
		)

		ctx, cancel := appContext(clicontext)
		defer cancel()

		client, err := getClient(clicontext)
		if err != nil {
			return err
		}

		if manifest := clicontext.String("manifest"); manifest != "" {
			desc.Digest, err = digest.Parse(manifest)
			if err != nil {
				return errors.Wrap(err, "invalid manifest digest")
			}
			desc.MediaType = clicontext.String("manifest-type")
		} else {
			img, err := client.ImageService().Get(ctx, local)
			if err != nil {
				return errors.Wrap(err, "unable to resolve image to manifest")
			}
			desc = img.Target
		}

		return client.Export(ctx, desc,
			containerd.ExportSpec{
				Path: dir,
			},
		)
	},
}
