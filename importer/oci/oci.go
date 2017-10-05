// Package oci provides the importer for OCI Image Spec.
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
	"github.com/containerd/containerd/importer"
	"github.com/containerd/containerd/reference"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
)

// OCIv1Importer implements OCI Image Spec v1.
//
// Selector spec: containerd reference string (e.g. example.com/foo/bar:tag@digest) or reference object string (e.g. tag@digest).
// Locator part in reference string is ignored.
var OCIv1Importer importer.Importer = &ociImporter{}

type ociImporter struct {
}

func resolveOCIIndex(idx ocispec.Index, selector string) (*ocispec.Descriptor, error) {
	tag, dgst, err := resolveOCISelector(selector)
	if err != nil {
		return nil, err
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
	return nil, errors.Errorf("not found: %q (tag=%q, dgst=%q)", selector, tag, dgst)
}

// resolveOCISelector resolves selector string into (tag, digest)
//   containerd reference string (e.g. example.com/foo/bar:tag@digest) or reference object string (e.g. tag@digest).
//   Locator part in reference string is ignored.
func resolveOCISelector(selector string) (string, digest.Digest, error) {
	// if selector is containerd ref object string (e.g. tag, @digest, tag@digest)
	if !strings.Contains(selector, "/") {
		tag, dgst := reference.SplitObject(selector)
		if tag == "" && dgst == "" {
			return tag, dgst, errors.Errorf("unexpected selector (looks like containerd reference object string): %q", selector)
		}
		return tag, dgst, nil
	}

	// if selector is containerd ref string
	parsed, err := reference.Parse(selector)
	if err != nil {
		return "", "", errors.Wrapf(err, "unexpected selector (looks like containerd reference string): %q", selector)
	}
	// ignore parsed.Locator, which is out of scope of OCI Image Format.
	tag, dgst := reference.SplitObject(parsed.Object)
	if tag == "" && dgst == "" {
		return tag, dgst,
			errors.Errorf("unexpected selector (looks like containerd reference string with object %q): %q", parsed.Object, selector)
	}
	return tag, dgst, nil
}

func onUntarIndexJSON(r io.Reader, selector string) (*ocispec.Descriptor, error) {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}
	var idx ocispec.Index
	if err := json.Unmarshal(b, &idx); err != nil {
		return nil, err
	}
	return resolveOCIIndex(idx, selector)
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

func (oi *ociImporter) Import(ctx context.Context, store content.Store, ref string, reader io.Reader, selector string) (images.Image, error) {
	imgrec := images.Image{Name: ref}
	tr := tar.NewReader(reader)
	var desc *ocispec.Descriptor
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return imgrec, err
		}
		if hdr.Typeflag != tar.TypeReg && hdr.Typeflag != tar.TypeRegA {
			continue
		}
		if hdr.Name == "index.json" {
			desc, err = onUntarIndexJSON(tr, selector)
			if err != nil {
				return imgrec, err
			}
			continue
		}
		if strings.HasPrefix(hdr.Name, "blobs/") {
			if err := onUntarBlob(ctx, tr, store, hdr.Name, hdr.Size); err != nil {
				return imgrec, err
			}
		}
	}
	if desc == nil {
		return imgrec, errors.Errorf("no descriptor found for selector %q", selector)
	}
	imgrec.Target = *desc
	return imgrec, nil
}
