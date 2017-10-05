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
	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// V1Importer implements OCI Image Spec v1.
type V1Importer struct {
	// ImageName is preprended to either `:` + OCI ref name or `@` + digest (for anonymous refs).
	// This field is mandatory atm, but may change in the future. maybe ref map[string]string as in moby/moby#33355
	ImageName string
}

var _ images.Importer = &V1Importer{}

// Import implements Importer.
func (oi *V1Importer) Import(ctx context.Context, store content.Store, reader io.Reader) ([]images.Image, error) {
	if oi.ImageName == "" {
		return nil, errors.New("ImageName not set")
	}
	tr := tar.NewReader(reader)
	var imgrecs []images.Image
	foundIndexJSON := false
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
			if foundIndexJSON {
				return nil, errors.New("duplicated index.json")
			}
			foundIndexJSON = true
			imgrecs, err = onUntarIndexJSON(tr, oi.ImageName)
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
	if !foundIndexJSON {
		return nil, errors.New("no index.json found")
	}
	return imgrecs, nil
}

func onUntarIndexJSON(r io.Reader, imageName string) ([]images.Image, error) {
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
		ref, err := normalizeImageRef(imageName, m)
		if err != nil {
			return nil, err
		}
		imgrecs = append(imgrecs, images.Image{
			Name:   ref,
			Target: m,
		})
	}
	return imgrecs, nil

}

func normalizeImageRef(imageName string, manifest ocispec.Descriptor) (string, error) {
	digest := manifest.Digest
	if digest == "" {
		return "", errors.Errorf("manifest with empty digest: %v", manifest)
	}
	ociRef := manifest.Annotations[ocispec.AnnotationRefName]
	if ociRef == "" {
		return imageName + "@" + digest.String(), nil
	}
	return imageName + ":" + ociRef, nil
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
