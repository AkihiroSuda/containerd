package main

import (
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/reference"
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
			Name:  "format",
			Value: "oci",
			Usage: "Format (current supported formats: oci)",
		},
		cli.BoolFlag{
			Name:  "translate-media-types",
			Usage: "Translate media types (e.g. docker -> oci)",
		},
		cli.StringFlag{
			Name:  "oci-ref-name",
			Value: "",
			Usage: "Override org.opencontainers.image.ref.name annotation",
		},
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

		if desc.Annotations == nil {
			desc.Annotations = make(map[string]string)
		}
		if ociRefName := determineOCIRefName(local); ociRefName != "" {
			desc.Annotations[ocispec.AnnotationRefName] = ociRefName
		}
		if ociRefName := clicontext.String("oci-ref-name"); ociRefName != "" {
			desc.Annotations[ocispec.AnnotationRefName] = ociRefName
		}
		if format := clicontext.String("format"); format != "oci" {
			return errors.Errorf("unsupported format: %s", format)
		}
		return client.Export(ctx,
			desc,
			containerd.WithOCIDirectoryExportation(dir),
			containerd.WithMediaTypeTranslation(clicontext.Bool("translate-media-types")),
		)
	},
}

func determineOCIRefName(local string) string {
	refspec, err := reference.Parse(local)
	if err != nil {
		return ""
	}
	tag, _ := reference.SplitObject(refspec.Object)
	return tag
}
