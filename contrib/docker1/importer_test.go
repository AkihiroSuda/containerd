package docker1

import (
	"testing"

	"github.com/containerd/containerd/errdefs"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var exampleManifests = []manifest{
	{
		Config: "0000000000000000000000000000000000000000000000000000000000000000.json",
		RepoTags: []string{
			"busybox:latest",
		},
	},
	{
		Config: "0000000000000000000000000000000000000000000000000000000000000001.json",
		RepoTags: []string{
			"foo:bar",
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
			filter:   "index==1",
			expected: "0000000000000000000000000000000000000000000000000000000000000001.json",
		},
		// repoTag filter is not supported yet
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
			assert.Equal(t, tc.expected, filtered.Config, tc.filter)
		}
	}
}

func TestImporterFilterWithSingleManifest(t *testing.T) {
	x := []manifest{
		{
			Config: "0000000000000000000000000000000000000000000000000000000000000000.json",
			RepoTags: []string{
				"busybox:latest",
			},
		},
	}
	testCases := []struct {
		filter      string
		expected    string
		expectedErr error
	}{
		{
			filter:   "index==0",
			expected: "0000000000000000000000000000000000000000000000000000000000000000.json",
		},
		{
			filter:   "", // empty filter is not ambiguous if we have only 1 item
			expected: "0000000000000000000000000000000000000000000000000000000000000000.json",
		},
		{
			filter:      "repoTag==foobar",
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
			assert.Equal(t, tc.expected, filtered.Config, tc.filter)
		}
	}
}
