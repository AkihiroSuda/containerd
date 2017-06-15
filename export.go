package containerd

import (
	"archive/tar"
	"context"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/oci"
	ocispecs "github.com/opencontainers/image-spec/specs-go"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func (c *Client) exportToOCITar(ctx context.Context, desc ocispec.Descriptor, writer io.Writer, eopts exportOpts) error {
	tw := tar.NewWriter(writer)
	img := oci.Tar(tw)

	// For tar, we defer creating index until end of the function.
	if err := oci.Init(img, oci.InitOpts{SkipCreateIndex: true}); err != nil {
		return err
	}
	cs := c.ContentStore()
	handlers := images.Handlers(
		images.ChildrenHandler(cs),
		exportHandler(cs, img),
	)
	if err := images.Dispatch(ctx, handlers, desc); err != nil {
		return err
	}
	// For tar, we don't use oci.PutManifestDescriptorToIndex() which allows appending desc to existing index.json
	// but requires img to support random read access.
	return oci.WriteIndex(img,
		ocispec.Index{
			Versioned: ocispecs.Versioned{
				SchemaVersion: 2,
			},
			Manifests: []ocispec.Descriptor{desc},
		},
	)
}

func exportHandler(cs content.Store, img oci.ImageDriver) images.HandlerFunc {
	return func(ctx context.Context, desc ocispec.Descriptor) ([]ocispec.Descriptor, error) {
		r, err := cs.Reader(ctx, desc.Digest)
		if err != nil {
			return nil, err
		}
		w, err := oci.NewBlobWriter(img, desc.Digest.Algorithm())
		if err != nil {
			return nil, err
		}
		if _, err = io.Copy(w, r); err != nil {
			return nil, err
		}
		if err = w.Close(); err != nil {
			return nil, err
		}

		if d := w.Digest(); d != desc.Digest {
			return nil, errors.Errorf("descriptor has digest %s, written %s", desc.Digest, d)
		}
		return nil, nil
	}
}
