package oci

import (
	"testing"

	"github.com/containerd/containerd/errdefs"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var exampleManifests = []ocispec.Descriptor{
	{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    "sha256:772a956f8de2e8883b5b74dac1599eff29ee0beb09d02689a6d1805bf8487422",
		Annotations: map[string]string{
			ocispec.AnnotationRefName: "foo",
			"test": "manifest0",
		},
		Platform: &ocispec.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	},
	{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    "sha256:e5c50d48ae80fafb15b390277fc76b6f1cb6b942097cd4c5a55077af351bc28f",
		Annotations: map[string]string{
			ocispec.AnnotationRefName: "bar",
			"test": "manifest1",
		},
	},
	{
		MediaType: ocispec.MediaTypeImageManifest,
		// same digest with different annotations
		Digest: "sha256:e5c50d48ae80fafb15b390277fc76b6f1cb6b942097cd4c5a55077af351bc28f",
		Annotations: map[string]string{
			ocispec.AnnotationRefName: "baz",
			"test": "manifest2",
		},
		Platform: &ocispec.Platform{
			Architecture: "amd64",
			OS:           "linux",
		},
	},
	{
		MediaType: ocispec.MediaTypeImageManifest,
		Digest:    "sha256:2f96eb505eeb87da4cdcedd23c62865b3a5ab0fd8e55dd1427e95101e0702217",
		Annotations: map[string]string{
			// same annotation ref name but different platform
			ocispec.AnnotationRefName: "baz",
			"test": "manifest3",
		},
		Platform: &ocispec.Platform{
			Architecture: "arm",
			OS:           "linux",
			Variant:      "v7",
		},
	},
}

func TestImporterFilter(t *testing.T) {
	testCases := []struct {
		filter      string
		expected    string
		expectedErr error
	}{
		{
			filter:   "tag==foo",
			expected: "manifest0",
		},
		{
			filter:   "tag==foo,digest==sha256:772a956f8de2e8883b5b74dac1599eff29ee0beb09d02689a6d1805bf8487422",
			expected: "manifest0",
		},
		{
			filter:   "tag==baz,digest==sha256:e5c50d48ae80fafb15b390277fc76b6f1cb6b942097cd4c5a55077af351bc28f",
			expected: "manifest2",
		},
		{
			filter:   "index==2",
			expected: "manifest2",
		},
		{
			filter:      "digest==sha256:e5c50d48ae80fafb15b390277fc76b6f1cb6b942097cd4c5a55077af351bc28f",
			expectedErr: errdefs.ErrAmbiguous,
		},
		{
			filter:   "tag==baz,architecture==amd64",
			expected: "manifest2",
		},
		{
			filter:   "tag==baz,architecture==arm",
			expected: "manifest3",
		},
		{
			filter:      "tag==baz,os==linux",
			expectedErr: errdefs.ErrAmbiguous,
		},
		{
			filter:      "tag==f00,digest==sha256:e5c50d48ae80fafb15b390277fc76b6f1cb6b942097cd4c5a55077af351bc28f",
			expectedErr: errdefs.ErrNotFound,
		},
		{
			filter:      "tag==a",
			expectedErr: errdefs.ErrNotFound,
		},
		{
			filter:      "",
			expectedErr: errdefs.ErrAmbiguous,
		},
	}
	for _, tc := range testCases {
		filtered, err := filterManifests(exampleManifests, tc.filter)
		if tc.expectedErr != nil {
			assert.Equal(t, tc.expectedErr, errors.Cause(err), tc.filter)
		} else {
			assert.Nil(t, err, tc.filter)
		}
		if tc.expected != "" {
			assert.NotNil(t, filtered, tc.filter)
			assert.Equal(t, tc.expected, filtered.Annotations["test"], tc.filter)
		}
	}
}

func TestImporterFilterWithSingleManifest(t *testing.T) {
	x := []ocispec.Descriptor{
		{
			Digest: "sha256:772a956f8de2e8883b5b74dac1599eff29ee0beb09d02689a6d1805bf8487422",
			Annotations: map[string]string{
				ocispec.AnnotationRefName: "foo",
				"test": "manifest0",
			},
		},
	}
	testCases := []struct {
		filter      string
		expected    string
		expectedErr error
	}{
		{
			filter:   "tag==foo",
			expected: "manifest0",
		},
		{
			filter:   "", // empty filter is not ambiguous if we have only 1 item
			expected: "manifest0",
		},
		{
			filter:      "tag==a",
			expectedErr: errdefs.ErrNotFound,
		},
	}
	for _, tc := range testCases {
		filtered, err := filterManifests(x, tc.filter)

		if tc.expectedErr != nil {
			assert.Equal(t, tc.expectedErr, errors.Cause(err), tc.filter)
		} else {
			assert.Nil(t, err, tc.filter)
		}
		if tc.expected != "" {
			assert.NotNil(t, filtered, tc.filter)
			assert.Equal(t, tc.expected, filtered.Annotations["test"], tc.filter)
		}
	}
}
