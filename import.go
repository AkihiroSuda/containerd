package containerd

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/oci"
	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

func resolveOCIIndex(idx ocispec.Index, refObject string) (*ocispec.Descriptor, error) {
	tag, dgst := reference.SplitObject(refObject)
	if tag == "" && dgst == "" {
		return nil, errors.Errorf("unexpected object: %q", refObject)
	}
	for _, m := range idx.Manifests {
		if m.Digest == dgst {
			return &m, nil
		}
		annot, ok := m.Annotations[ocispec.AnnotationRefName]
		if ok && annot == tag && tag != "" {
			return &m, nil
		}
	}
	return nil, errors.Errorf("not found: %q", refObject)
}

func (c *Client) importFromOCITar(ctx context.Context, ref string, reader io.Reader, iopts importOpts) (Image, error) {
	tr := tar.NewReader(reader)
	store := c.ContentStore()
	var desc *ocispec.Descriptor
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		if hdr.Name == "index.json" {
			desc, err = onUntarIndexJSON(tr, iopts.refObject)
			if err != nil {
				return nil, err
			}
			continue
		}
		if strings.HasPrefix(hdr.Name, "blobs/") {
			if err := onUntarBlob(ctx, tr, store, hdr.Name, hdr.Size); err != nil {
				return nil, err
			}
		}
	}
	if desc == nil {
		return nil, errors.Errorf("no descriptor found for reference object %q", iopts.refObject)
	}
	is := c.ImageService()
	if err := is.Update(ctx, ref, *desc); err != nil {
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

func onUntarIndexJSON(r io.Reader, refObject string) (*ocispec.Descriptor, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var idx ocispec.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return resolveOCIIndex(idx, refObject)
}

func onUntarBlob(ctx context.Context, r io.Reader, store content.Store, name string, size int64) error {
	// name is like "blobs/sha256/deadbeef"
	split := strings.Split(name, "/")
	if len(split) != 3 {
		return errors.Errorf("unexpected name: %q", name)
	}
	algo := digest.Algorithm(split[1])
	if !algo.Available() {
		return errors.Errorf("unsupported algorithm: %s", algo)
	}
	dgst := digest.NewDigestFromHex(algo.String(), split[2])
	return content.WriteBlob(ctx, store, "unknown-"+dgst.String(), r, size, dgst)
}

func (c *Client) importFromOCIDirectory(ctx context.Context, ref string, opener func(string) (io.ReadCloser, error), iopts importOpts) (Image, error) {
	store := c.ContentStore()

	imgdrv := oci.ReadOnly(opener)
	idx, err := oci.ReadIndex(imgdrv)
	if err != nil {
		return nil, err
	}
	desc, err := resolveOCIIndex(idx, iopts.refObject)
	if err != nil {
		return nil, err
	}
	fetcher := &ociImageFetcher{img: imgdrv}

	handler := images.Handlers(
		remotes.FetchHandler(store, fetcher),
		images.ChildrenHandler(store),
	)

	if err := images.Dispatch(ctx, handler, *desc); err != nil {
		return nil, err
	}

	is := c.ImageService()
	if err := is.Update(ctx, ref, *desc); err != nil {
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

type ociImageFetcher struct {
	img oci.ImageDriver
}

func (f *ociImageFetcher) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return oci.GetBlobReader(f.img, desc.Digest)
}
