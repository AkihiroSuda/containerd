package filesystem

import (
	"context"
	"io"
	"path/filepath"
	"strings"

	ociimage "github.com/AkihiroSuda/filegrain/image"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/reference"
	"github.com/containerd/containerd/remotes"
	"github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type imageType int

const (
	imageTypeUnknown imageType = iota
	imageTypeOCI
)

const (
	localhost = "localhost"
)

type filesystemResolver struct {
}

// ResolverOptions are used to configured a new Filesystem register resolver
type ResolverOptions struct {
}

// NewResolver returns a new resolver to a Filesystem registry
func NewResolver(options ResolverOptions) remotes.Resolver {
	return &filesystemResolver{}
}

var _ remotes.Resolver = &filesystemResolver{}

func (r *filesystemResolver) imageType(ctx context.Context, dir string) (imageType, error) {
	layout, err := ociimage.ReadImageLayout(dir)
	if err != nil {
		return imageTypeUnknown, err
	}
	if layout.Version != ocispec.ImageLayoutVersion {
		log.G(ctx).Warnf("supported layout version %s, image version %s (%s)",
			ocispec.ImageLayoutVersion, layout.Version, dir)
	}
	return imageTypeOCI, nil
}

func (r *filesystemResolver) resolveOCI(ctx context.Context, ref, dir, tag string, dgst digest.Digest) (string, ocispec.Descriptor, remotes.Fetcher, error) {
	idx, err := ociimage.ReadIndex(dir)
	if err != nil {
		return "", ocispec.Descriptor{}, nil, err
	}
	fetcher := &ociFetcher{dir: dir}
	for _, m := range idx.Manifests {
		if m.Digest == dgst && dir != "" {
			return ref, m, fetcher, nil
		}
		annot, ok := m.Annotations[ociimage.RefNameAnnotation]
		if ok && annot == tag && tag != "" {
			return ref, m, fetcher, nil
		}
	}
	return "", ocispec.Descriptor{}, nil, errors.Errorf("%v not found", ref)
}

func (r *filesystemResolver) Resolve(ctx context.Context, ref string) (string, ocispec.Descriptor, remotes.Fetcher, error) {
	refspec, err := reference.Parse(ref)
	if err != nil {
		return "", ocispec.Descriptor{}, nil, err
	}
	locator := strings.Split(refspec.Locator, "/")
	if len(locator) < 2 {
		return "", ocispec.Descriptor{}, nil, errors.Errorf("invalid locator: %q", refspec.Locator)
	}
	host, dir := locator[0], filepath.Join("/", strings.Join(locator[1:], "/"))
	if host != localhost {
		return "", ocispec.Descriptor{}, nil, errors.Errorf("host must be %q, got %q", localhost, host)
	}
	tag, dgst := reference.SplitObject(refspec.Object)
	typ, err := r.imageType(ctx, dir)
	if err != nil {
		return "", ocispec.Descriptor{}, nil, err
	}
	switch typ {
	case imageTypeOCI:
		return r.resolveOCI(ctx, ref, dir, tag, dgst)
	}
	return "", ocispec.Descriptor{}, nil, errors.Errorf("image type unknown for %s", dir)
}

type ociFetcher struct {
	dir string
}

func (f *ociFetcher) Fetch(ctx context.Context, desc ocispec.Descriptor) (io.ReadCloser, error) {
	return ociimage.GetBlobReader(f.dir, desc.Digest)
}
