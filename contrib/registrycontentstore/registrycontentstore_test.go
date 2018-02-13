package registrycontentstore

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"testing"

	"github.com/containerd/containerd/content"
	"github.com/containerd/containerd/images"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/testutil/registry"
	digest "github.com/opencontainers/go-digest"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

func newRegistryContentStore(t *testing.T, verbose bool) (*registry.Registry, *Store, func()) {
	reg := registry.New(t, verbose)
	reg.Start(t)
	regURL, err := url.Parse(reg.URL())
	if err != nil {
		t.Fatal(err)
	}
	ref := regURL.Host + "/foo/bar:latest"
	t.Logf("started registry %q for ref %q", regURL.String(), ref)
	resolver := docker.NewResolver(docker.ResolverOptions{PlainHTTP: true})
	store, err := NewStore(ref, resolver)
	if err != nil {
		t.Fatal(err)
	}
	cleanup := func() {
		reg.Close()
	}
	return reg, store, cleanup
}

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

func TestRegistryContentStore(t *testing.T) {
	_, store, cleanup := newRegistryContentStore(t, false)
	defer cleanup()
	testContentStore(t, store)
}

func testContentStore(t *testing.T, store interface {
	content.Provider
	content.Ingester
}) {
	blobs := exampleBlobs(t)
	ctx := context.TODO()
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
		r := &readerAtReader{ReaderAt: ra}
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
}

func (r *readerAtReader) Read(p []byte) (int, error) {
	return r.ReadAt(p, 0)
}
