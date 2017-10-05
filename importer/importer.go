package importer

import (
	"context"
	"io"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
)

// Importer is the interface for image importer
type Importer interface {
	// Import imports an image from a tar stream.
	// Selector is implementation-specific, but generally ref-compatible string.
	Import(ctx context.Context, store content.Store, ref string, reader io.Reader, selector string) (images.Image, error)
}
