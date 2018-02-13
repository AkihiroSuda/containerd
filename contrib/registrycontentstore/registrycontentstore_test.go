package registrycontentstore

import (
	"context"
	"net/url"
	"testing"

	"github.com/containerd/containerd/content/testsuite"
	"github.com/containerd/containerd/remotes/docker"
	"github.com/containerd/containerd/testutil/registry"
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

func TestRegistryContentStore(t *testing.T) {
	_, store, cleanup := newRegistryContentStore(t, false)
	defer cleanup()
	testsuite.TestMediaType(context.TODO(), t, store)
}
