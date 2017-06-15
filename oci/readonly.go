package oci

import (
	"errors"
	"io"
	"os"

	"github.com/opencontainers/go-digest"
)

// ReadOnly is ImageDriver for readonly representation of an OCI image.
// opener is typically `func(relPath string)(io.ReadCloser, error){ return os.Open(filepath.Join(someDir, relPath)) }`.
func ReadOnly(opener func(relPath string) (io.ReadCloser, error)) ImageDriver {
	return &readonly{opener: opener}
}

type readonly struct {
	opener func(relPath string) (io.ReadCloser, error)
}

func (d *readonly) Init() error {
	return errors.New("ReadOnly does not support Init")
}

func (d *readonly) Remove(path string) error {
	return errors.New("ReadOnly does not support Remove")
}

func (d *readonly) Reader(path string) (io.ReadCloser, error) {
	return d.opener(path)
}

func (d *readonly) Writer(path string, perm os.FileMode) (io.WriteCloser, error) {
	return nil, errors.New("ReadOnly does not support Writer")
}

func (d *readonly) BlobWriter(algo digest.Algorithm) (BlobWriter, error) {
	return nil, errors.New("ReadOnly does not support BlobWriter")
}
