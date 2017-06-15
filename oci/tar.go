package oci

import (
	"archive/tar"
	"bytes"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/opencontainers/go-digest"
)

// TarWriter is an interface that is implemented by archive/tar.Writer.
// (Using an interface allows hooking)
type TarWriter interface {
	io.WriteCloser
	Flush() error
	WriteHeader(hdr *tar.Header) error
}

// Tar is ImageDriver for TAR representation of an OCI image.
func Tar(w TarWriter) ImageDriver {
	return &tarDriver{
		w: w,
	}
}

type tarDriver struct {
	w TarWriter
}

func (d *tarDriver) Init() error {
	headers := []tar.Header{
		{
			Name:     "blobs/",
			Mode:     0755,
			Typeflag: tar.TypeDir,
		},
		{
			Name:     "blobs/" + string(digest.Canonical) + "/",
			Mode:     0755,
			Typeflag: tar.TypeDir,
		},
	}
	for _, h := range headers {
		if err := d.w.WriteHeader(&h); err != nil {
			return err
		}
	}
	return nil
}

func (d *tarDriver) Remove(path string) error {
	return errors.New("Tar does not support Remove")
}

func (d *tarDriver) Reader(path string) (io.ReadCloser, error) {
	// because tar does not support random access
	return nil, errors.New("Tar does not support Reader")
}

func (d *tarDriver) Writer(path string, perm os.FileMode) (io.WriteCloser, error) {
	name := strings.Join(strings.Split(path, string(os.PathSeparator)), "/")
	return &tarDriverWriter{
		w:    d.w,
		name: name,
		mode: int64(perm),
	}, nil
}

type tarDriverWriter struct {
	bytes.Buffer
	w    TarWriter
	name string
	mode int64
}

func (w *tarDriverWriter) Close() error {
	if err := w.w.WriteHeader(&tar.Header{
		Name:     w.name,
		Mode:     w.mode,
		Size:     int64(w.Len()),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return err
	}
	n, err := io.Copy(w.w, w)
	if err != nil {
		return err
	}
	if n < int64(w.Len()) {
		return io.ErrShortWrite
	}
	return w.w.Flush()
}

func (d *tarDriver) BlobWriter(algo digest.Algorithm) (BlobWriter, error) {
	return &tarBlobWriter{
		w:        d.w,
		digester: algo.Digester(),
	}, nil
}

// tarBlobWriter implements BlobWriter.
type tarBlobWriter struct {
	w        TarWriter
	digester digest.Digester
	buf      bytes.Buffer // TODO: use tmp file for large buffer?
	closed   bool
}

// Write implements io.Writer.
func (bw *tarBlobWriter) Write(b []byte) (int, error) {
	n, err := bw.buf.Write(b)
	if err != nil {
		return n, err
	}
	return bw.digester.Hash().Write(b)
}

// Close implements io.Closer.
func (bw *tarBlobWriter) Close() error {
	path := "blobs/" + bw.digester.Digest().Algorithm().String() + "/" + bw.digester.Digest().Hex()
	if err := bw.w.WriteHeader(&tar.Header{
		Name:     path,
		Mode:     0444,
		Size:     int64(bw.buf.Len()),
		Typeflag: tar.TypeReg,
	}); err != nil {
		return err
	}
	n, err := io.Copy(bw.w, &bw.buf)
	if err != nil {
		return err
	}
	if n < int64(bw.buf.Len()) {
		return io.ErrShortWrite
	}
	if err := bw.w.Flush(); err != nil {
		return err
	}
	bw.closed = true
	return nil
}

// Digest returns the digest when closed.
func (bw *tarBlobWriter) Digest() digest.Digest {
	if !bw.closed {
		panic("blobWriter is unclosed")
	}
	return bw.digester.Digest()
}
