/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package testsuite

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

type blob struct {
	blob []byte
	desc ocispec.Descriptor
}

type dockerManifestList struct {
	SchemaVersion int `json:"schemaVersion"`
	// MediaType is not available in ocispec.Index structure.
	// However, Docker registry requires this field to be set: https://github.com/docker/distribution/blob/6664ec703991875e14419ff4960921cce7678bab/registry/storage/manifeststore.go#L105
	MediaType string               `json:"mediaType,omitempty"`
	Manifests []ocispec.Descriptor `json:"manifests"`
}

// exampleBlobs generates two blobs:
//  - application/vnd.example.foobar
//  - application/vnd.docker.distribution.manifest.list.v2+json
// Some implementations may treat the latter one in a specific way.
func exampleBlobs(t *testing.T) []blob {
	var (
		blobs []blob
		blob  blob
		err   error
	)
	blob.blob = []byte(`foobar`)
	blob.desc = ocispec.Descriptor{
		MediaType: "application/vnd.example.foobar",
		Digest:    digest.FromBytes(blob.blob),
		Size:      int64(len(blob.blob)),
	}
	blobs = append(blobs, blob)
	idx := dockerManifestList{
		SchemaVersion: 2,
		MediaType:     images.MediaTypeDockerSchema2ManifestList,
		Manifests:     []ocispec.Descriptor{blob.desc},
	}
	blob.blob, err = json.Marshal(&idx)
	if err != nil {
		t.Fatal(err)
	}
	blob.desc = ocispec.Descriptor{
		MediaType: images.MediaTypeDockerSchema2ManifestList,
		Digest:    digest.FromBytes(blob.blob),
		Size:      int64(len(blob.blob)),
	}
	blobs = append(blobs, blob)
	return blobs
}

// TestMediaType tests the store with several mediatype strings.
func TestMediaType(ctx context.Context, t *testing.T, store interface {
	content.Provider
	content.Ingester
}) {
	blobs := exampleBlobs(t)
	for i, blob := range blobs {
		t.Logf("testing with blob %d, desc=%+v, blob=%q", i, blob.desc, string(blob.blob))
		// test write
		w, err := store.Writer(ctx, fmt.Sprintf("ref-%d", i), blob.desc)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write(blob.blob); err != nil {
			t.Fatal(err)
		}
		if err := w.Commit(ctx, blob.desc.Size, blob.desc.Digest); err != nil {
			t.Fatal(err)
		}
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		// test read
		ra, err := store.ReaderAt(ctx, blob.desc)
		if err != nil {
			t.Fatal(err)
		}
		r := &readerAtReader{ReaderAt: ra, off: 0}
		b, err := ioutil.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if err := ra.Close(); err != nil {
			t.Fatal(err)
		}
		d := digest.FromBytes(b)
		if blob.desc.Digest != d {
			t.Fatalf("expected %s (%q), got %s (%q) for blob %d.", blob.desc.Digest, string(blob.blob),
				d, string(b), i)
		}
	}
}

type readerAtReader struct {
	io.ReaderAt
	off int64
}

func (r *readerAtReader) Read(p []byte) (int, error) {
	l, err := r.ReadAt(p, r.off)
	if err != nil {
		return l, err
	}
	r.off = int64(l)
	return l, err
}
