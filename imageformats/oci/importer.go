// Package oci provides the importer and the exporter for OCI Image Spec.
package oci

import (
	"archive/tar"
	"context"
	"encoding/json"
	"io"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/errdefs"
	"github.com/containerd/containerd/filters"
	"github.com/containerd/containerd/imageformats"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

type v1Importer struct {
}

// V1Importer implements OCI Image Spec v1.
//
// Supported filters:
// - index:        manifest list index (e.g. "0")
// - tag:          org.opencontainers.image.ref.name tag (e.g. "latest")
// - digest:       manifest digest (e.g. sha256:deadbeefdeadbeef)
// - mediaType:    manifest media type ("e.g. application/vnd.oci.image.manifest.v1+json")
// - architecture: e.g. "amd64"
// - os:           e.g. "linux"
// - variant:      e.g. "v7" (when architecture is "arm")
var V1Importer imageformats.Importer = &v1Importer{}

func (oi *v1Importer) Import(ctx context.Context, store content.Store, reader io.Reader, filter string) (*ocispec.Descriptor, error) {
	tr := tar.NewReader(reader)
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
			desc, err = onUntarIndexJSON(tr, filter)
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
		return nil, errors.Errorf("no descriptor found for filter %q", filter)
	}
	return desc, nil
}

func onUntarIndexJSON(r io.Reader, filter string) (*ocispec.Descriptor, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var idx ocispec.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return filterManifests(idx.Manifests, filter)
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

func filterManifests(manifests []ocispec.Descriptor, filter string) (*ocispec.Descriptor, error) {
	f, err := filters.Parse(filter)
	if err != nil {
		return nil, err
	}
	var matched []ocispec.Descriptor
	for i, m := range manifests {
		adapter := imageAdapter{
			index:    i,
			manifest: &m,
		}
		ok := f.Match(adaptImage(adapter))
		if ok {
			matched = append(matched, m)
		}
	}
	if l := len(matched); l > 1 {
		return nil, errors.Wrapf(errdefs.ErrAmbiguous, "filter %q matched %d items", filter, l)
	}
	if len(matched) == 0 {
		return nil, errors.Wrapf(errdefs.ErrNotFound, "filter %q did not match any item", filter)
	}
	return &matched[0], nil
}

type imageAdapter struct {
	index    int
	manifest *ocispec.Descriptor
}

func adaptImage(obj imageAdapter) filters.Adaptor {
	return filters.AdapterFunc(func(fieldpath []string) (string, bool) {
		if len(fieldpath) == 0 {
			return "", false
		}
		platform := obj.manifest.Platform
		switch fieldpath[0] {
		case "index":
			return strconv.Itoa(obj.index), true
		case "tag":
			tag, ok := obj.manifest.Annotations[ocispec.AnnotationRefName]
			return tag, ok
		case "digest":
			dgst := obj.manifest.Digest.String()
			return dgst, len(dgst) > 0
		case "mediaType":
			return obj.manifest.MediaType, len(obj.manifest.MediaType) > 0
		case "architecture":
			if platform != nil && platform.Architecture != "" {
				return platform.Architecture, true
			}
		case "os":
			if platform != nil && platform.OS != "" {
				return platform.OS, true
			}
		case "variant":
			if platform != nil && platform.Variant != "" {
				return platform.Variant, true
			}
		}
		return "", false
	})
}
