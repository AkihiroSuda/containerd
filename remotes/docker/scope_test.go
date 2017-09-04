package docker

import (
	"testing"

	"github.com/containerd/containerd/reference"
	"github.com/stretchr/testify/assert"
)

func TestRepositoryScope(t *testing.T) {
	testCases := []struct {
		refspec  reference.Spec
		push     bool
		expected string
	}{
		{
			refspec: reference.Spec{
				Locator: "host/foo/bar",
				Object:  "ignored",
			},
			push:     false,
			expected: "repository:foo/bar:pull",
		},
		{
			refspec: reference.Spec{
				Locator: "host:4242/foo/bar",
				Object:  "ignored",
			},
			push:     true,
			expected: "repository:foo/bar:pull,push",
		},
	}
	for _, x := range testCases {
		actual, err := repositoryScope(x.refspec, x.push)
		assert.NoError(t, err)
		assert.Equal(t, x.expected, actual)
	}
}

func TestScopeEquals(t *testing.T) {
	testCases := []struct {
		a        string
		b        string
		expected bool
	}{
		{
			a:        "repository:foo/bar:pull",
			b:        "repository:foo/baz:pull",
			expected: false,
		},
		{
			a:        "repository:foo/bar:pull",
			b:        "repository:foo/bar:pull,push",
			expected: false,
		},
		{
			a:        "repository:foo/bar:push,pull",
			b:        "repository:foo/bar:pull,push",
			expected: true,
		},
		{
			a:        "repository:foo/bar:pull repository:qux/quux:pull",
			b:        "repository:qux/quux:pull repository:foo/bar:pull",
			expected: true,
		},
	}
	for _, x := range testCases {
		assert.Equal(t, x.expected, scopeEquals(x.a, x.b), "a=%q, b=%q", x.a, x.b)
	}
}
