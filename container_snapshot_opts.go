package containerd

import (
	"context"
	"fmt"

	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/mount"
	"github.com/containerd/containerd/platforms"
	"github.com/opencontainers/go-digest"
	"github.com/opencontainers/image-spec/identity"
	"github.com/pkg/errors"
)

// SnapshotModifier modifies the snapshot for initialization purpose.
type SnapshotModifier interface {
	// ID is unique string per snapshot modifier.
	// The client library try to find the cached modified snapshot using the hash of IDs of []SnapshotModifier.
	ID() string
	// Modify modifies the snapshot
	Modify(mounts []mount.Mount) error
}

func chainSnapshotModifierID(parent string, modifiers []SnapshotModifier) string {
	s := ""
	for _, m := range modifiers {
		s += m.ID()
	}
	return fmt.Sprintf("%s-mod-%s", parent, digest.SHA256.FromString(s).Hex())
}

// WithSnapshotModifiers configures the container to use the modified snapshot with modifiers.
// The container will have c.SnapshotKey as key and c.Image as i.Name().
func WithSnapshotModifiers(key string, i Image, makeReadonly bool, modifiers ...SnapshotModifier) NewContainerOpts {
	return func(ctx context.Context, client *Client, c *containers.Container) error {
		diffIDs, err := i.(*image).i.RootFS(ctx, client.ContentStore(), platforms.Default())
		if err != nil {
			return errors.Wrapf(err, "could not call RootFS for image %s", i.Name())
		}

		setSnapshotterIfEmpty(c)

		var (
			snapshotter = client.SnapshotService(c.Snapshotter)
			parent      = identity.ChainID(diffIDs).String()
			committed   = chainSnapshotModifierID(parent, modifiers)
		)
		// FIXME: should return err if err != nil && err != ErrNotFound
		if _, err := snapshotter.Stat(ctx, committed); err != nil {
			workspace := committed + "-workspace"
			mounts, err := snapshotter.Prepare(ctx, workspace, parent)
			if err != nil {
				return errors.Wrapf(err, "error while preparing workspace from %s", parent)
			}
			defer snapshotter.Remove(ctx, workspace)
			for _, modifier := range modifiers {
				if err := modifier.Modify(mounts); err != nil {
					return errors.Wrapf(err, "error while calling %+v (%s)", modifier, modifier.ID())
				}
			}
			if err := snapshotter.Commit(ctx, committed, workspace); err != nil {
				return errors.Wrap(err, "error while committing workspace")
			}
		}

		if makeReadonly {
			_, err = snapshotter.View(ctx, key, committed)
		} else {
			_, err = snapshotter.Prepare(ctx, key, committed)
		}
		if err != nil {
			return errors.Wrapf(err, "error while preparing %s from %s", key, committed)
		}
		c.SnapshotKey = key
		c.Image = i.Name()
		return nil
	}
}
