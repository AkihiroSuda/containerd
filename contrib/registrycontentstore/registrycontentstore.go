// Package registrycontentstore provides a registry-based implementation of the content package.
package registrycontentstore

import (
	"context"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/remotes"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// NewStore returns a store of which backend is a registry referenced by ref.
func NewStore(ref string, resolver remotes.Resolver) (*Store, error) {
	return &Store{
		ref:      ref,
		resolver: resolver,
	}, nil
}

// Store implements content.Provider and content.Ingester
type Store struct {
	ref      string
	resolver remotes.Resolver
}

// ReaderAt implements content.Provider. desc.MediaType must be set for manifest blobs.
func (s *Store) ReaderAt(ctx context.Context, desc ocispec.Descriptor) (content.ReaderAt, error) {
	fetcher, err := s.resolver.Fetcher(ctx, s.ref)
	if err != nil {
		return nil, err
	}
	// fetcher requires desc.MediaType to determine the GET URL, especially for manifest blobs.
	r, err := fetcher.Fetch(ctx, desc)
	if err != nil {
		return nil, err
	}
	return &readerAt{
		r:    r,
		desc: desc,
	}, nil
}

type readerAt struct {
	r    io.ReadCloser
	desc ocispec.Descriptor
}

func (r *readerAt) ReadAt(p []byte, off int64) (n int, err error) {
	if ra, ok := r.r.(io.ReaderAt); ok {
		return ra.ReadAt(p, off)
	}
	if off != 0 {
		return 0, errors.Wrap(errdefs.ErrInvalidArgument, "fetcher does not support non-zero offset")
	}
	return r.r.Read(p)
}

func (r *readerAt) Close() error {
	return r.r.Close()
}

func (r *readerAt) Size() int64 {
	return r.desc.Size
}

// ReaderAt implements content.Ingester. desc.MediaType must be set for manifest blobs.
func (s *Store) Writer(ctx context.Context, ref string, desc ocispec.Descriptor) (content.Writer, error) {
	pusher, err := s.resolver.Pusher(ctx, s.ref)
	if err != nil {
		return nil, err
	}
	// pusher requires desc.MediaType to determine the PUT URL, especially for manifest blobs.
	contentWriter, err := pusher.Push(ctx, desc)
	if err != nil {
		return nil, err
	}
	return &writer{
		Writer:           contentWriter,
		contentWriterRef: ref,
	}, nil
}

type writer struct {
	content.Writer          // returned from pusher.Push
	contentWriterRef string // ref passed for Writer()
}

func (w *writer) Status() (content.Status, error) {
	st, err := w.Writer.Status()
	if err != nil {
		return st, err
	}
	if w.contentWriterRef != "" {
		st.Ref = w.contentWriterRef
	}
	return st, nil
}

var _ content.Provider = &Store{}
var _ content.Ingester = &Store{}
