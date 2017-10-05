// Package oci provides the importer and the exporter for OCI Image Spec.
package oci

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/imageformats"
	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// V1Importer implements OCI Image Spec v1.
type V1Importer struct {
	Prefix string // mandatory atm, may change in the future
}

var _ imageformats.Importer = &V1Importer{}

func (oi *V1Importer) Import(ctx context.Context, store content.Store, reader io.Reader) ([]images.Image, error) {
	if oi.Prefix == "" {
		return nil, errors.New("empty prefix")
	}
	tr := tar.NewReader(reader)
	var imgrecs []images.Image
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
			if len(imgrecs) != 0 {
				return nil, errors.New("duplicated index.json?")
			}
			imgrecs, err = onUntarIndexJSON(tr, oi.Prefix)
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
	return imgrecs, nil
}

func onUntarIndexJSON(r io.Reader, prefix string) ([]images.Image, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var idx ocispec.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	var imgrecs []images.Image
	for _, m := range idx.Manifests {
		ociRef := m.Annotations[ocispec.AnnotationRefName]
		if ociRef == "" {
			// TODO(AkihiroSuda): print warning?
			continue
		}
		imgrecs = append(imgrecs, images.Image{
			Name:   prefix + ociRef,
			Target: m,
		})
	}
	return imgrecs, nil

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
