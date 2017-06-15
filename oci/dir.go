package oci

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/opencontainers/go-digest"
)

// Directory is ImageDriver for directory representation of an OCI image.
// The directory must not exist before calling Init function of this structure.
func Directory(path string) ImageDriver {
	return &directory{
		path: path,
	}
}

type directory struct {
	path string
}

func (d *directory) Init() error {
	if _, err := os.Stat(d.path); err == nil {
		return os.ErrExist
	}
	// Create the directory
	if err := os.MkdirAll(d.path, 0755); err != nil {
		return err
	}
	// Create blobs/sha256
	return d.initAlgo(digest.Canonical)
}

func (d *directory) initAlgo(algo digest.Algorithm) error {
	return os.MkdirAll(filepath.Join(d.path, "blobs", string(algo)), 0755)
}

func (d *directory) joinPath(path string) (string, error) {
	j := filepath.Clean(filepath.Join(d.path, path))
	if !strings.HasPrefix(j, d.path) {
		return "", fmt.Errorf("unexpected path: %q", path)
	}
	return j, nil
}

func (d *directory) Remove(path string) error {
	j, err := d.joinPath(path)
	if err != nil {
		return err
	}
	return os.Remove(j)
}

func (d *directory) Reader(path string) (io.ReadCloser, error) {
	j, err := d.joinPath(path)
	if err != nil {
		return nil, err
	}
	return os.Open(j)
}

func (d *directory) Writer(path string, perm os.FileMode) (io.WriteCloser, error) {
	j, err := d.joinPath(path)
	if err != nil {
		return nil, err
	}
	return os.Create(j)
}

type writer struct {
	*os.File
	perm os.FileMode
}

func (w *writer) Close() error {
	if err := w.Close(); err != nil {
		return err
	}
	return w.Chmod(w.perm)
}

func (d *directory) BlobWriter(algo digest.Algorithm) (BlobWriter, error) {
	if err := d.initAlgo(algo); err != nil {
		return nil, err
	}
	// use d.path rather than the default tmp, so as to make sure rename(2) can be applied
	f, err := ioutil.TempFile(d.path, "tmp.blobwriter")
	if err != nil {
		return nil, err
	}
	return &blobWriter{
		path:     d.path,
		digester: algo.Digester(),
		f:        f,
	}, nil
}

// blobWriter implements BlobWriter.
type blobWriter struct {
	path     string
	digester digest.Digester
	f        *os.File
	closed   bool
}

// Write implements io.Writer.
func (bw *blobWriter) Write(b []byte) (int, error) {
	n, err := bw.f.Write(b)
	if err != nil {
		return n, err
	}
	return bw.digester.Hash().Write(b)
}

// Close implements io.Closer.
func (bw *blobWriter) Close() error {
	oldPath := bw.f.Name()
	if err := bw.f.Close(); err != nil {
		return err
	}
	newPath := filepath.Join(bw.path, blobPath(bw.digester.Digest()))
	if err := os.Chmod(oldPath, 0444); err != nil {
		return err
	}
	if err := os.Rename(oldPath, newPath); err != nil {
		return err
	}
	bw.closed = true
	return nil
}

// Digest returns the digest when closed.
func (bw *blobWriter) Digest() digest.Digest {
	if !bw.closed {
		panic("blobWriter is unclosed")
	}
	return bw.digester.Digest()
}
