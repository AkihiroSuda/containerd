package containerd

import (
	"context"
	"io"

	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes"

	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func (c *Client) importFromOCIDirectory(ctx context.Context, ref string, iopts importOpts) (Image, error) {
	store := c.ContentStore()

	desc, err := resolveFromOCIDirectory(iopts)
	if err != nil {
		return nil, err
	}
	fetcher := &ociDirectoryFetcher{path: iopts.path}

	handler := images.Handlers(
		remotes.FetchHandler(store, fetcher),
		images.ChildrenHandler(store),
	)

	if err := images.Dispatch(ctx, handler, desc); err != nil {
		return nil, err
	}

	is := c.ImageService()
	if err := is.Update(ctx, ref, desc); err != nil {
		return nil, err
	}
	i, err := is.Get(ctx, ref)
	if err != nil {
		return nil, err
	}
	img := &image{
		client: c,
		i:      i,
	}
	return img, nil
}

func resolveFromOCIDirectory(iopts importOpts) (ocispec.Descriptor, error) {
	tag, dgst := reference.SplitObject(iopts.refObject)
	if tag == "" && dgst == "" {
		return ocispec.Descriptor{}, errors.Errorf("unexpected object: %q", iopts.refObject)
	}
	img := oci.Directory(iopts.path)
	idx, err := oci.ReadIndex(img)
	if err != nil {
		return ocispec.Descriptor{}, err
	}
	for _, m := range idx.Manifests {
		if m.Digest == dgst {
			return m, nil
		}
		annot, ok := m.Annotations[ocispec.AnnotationRefName]
		if ok && annot == tag && tag != "" {
			return m, nil
		}
	}
	return ocispec.Descriptor{}, errors.Errorf("not found: %q", iopts.refObject)
}

type ociDirectoryFetcher struct {
	path string
}

func (f *ociDirectoryFetcher) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	img := oci.Directory(f.path)
	return oci.GetBlobReader(img, desc.Digest)
}
