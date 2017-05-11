// +build linux freebsd

package zfs

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/fs/fsutils"
	"github.com/containerd/containerd/log"
	"github.com/containerd/containerd/plugin"
	"github.com/containerd/containerd/snapshot"
	"github.com/containerd/containerd/snapshot/storage"
	zfs "github.com/mistifyio/go-zfs"
	"github.com/pkg/errors"
)

const (
	// snapshotSuffix is used as follows:
	//	active := filepath.Join(dataset.Name, id)
	//      committed := active + "@" + snapshotSuffix
	snapshotSuffix = "snapshot"
)

func init() {
	plugin.Register("snapshot-zfs", &plugin.Registration{
		Type: plugin.SnapshotPlugin,
		Init: func(ic *plugin.InitContext) (interface{}, error) {
			root := filepath.Join(ic.Root, "snapshot", "zfs")
			return NewSnapshotter(root)
		},
	})
}

type snapshotter struct {
	dataset *zfs.Dataset
	ms      *storage.MetaStore
}

// NewSnapshotter returns a Snapshotter using zfs. Uses the provided
// root directory for snapshots and stores the metadata in
// a file in the provided root.
// root needs to be a mount point of zfs.
func NewSnapshotter(root string) (snapshot.Snapshotter, error) {
	mount, err := fsutils.LookupMount(root)
	if err != nil {
		return nil, err
	}
	if mount.FSType != "zfs" {
		return nil, fmt.Errorf("expected zfs, got %s", mount.FSType)
	}
	dataset, err := zfs.GetDataset(mount.Source)
	if err != nil {
		return nil, err
	}

	ms, err := storage.NewMetaStore(filepath.Join(root, "metadata.db"))
	if err != nil {
		return nil, err
	}

	b := &snapshotter{
		dataset: dataset,
		ms:      ms,
	}
	return b, nil
}

var (
	zfsCreateProperties = map[string]string{
		"mountpoint": "legacy",
	}
)

// createFilesystem creates but not mount.
func createFilesystem(datasetName string) (*zfs.Dataset, error) {
	return zfs.CreateFilesystem(datasetName, zfsCreateProperties)
}

// cloneFilesystem clones but not mount.
func cloneFilesystem(datasetName string, snapshot *zfs.Dataset) (*zfs.Dataset, error) {
	return snapshot.Clone(datasetName, zfsCreateProperties)
}

func destroy(dataset *zfs.Dataset) error {
	// TODO(AkihiroSuda): figure out proper DestroyFlag
	return dataset.Destroy(zfs.DestroyDefault)
}

// Stat returns the info for an active or committed snapshot by name or
// key.
//
// Should be used for parent resolution, existence checks and to discern
// the kind of snapshot.
func (b *snapshotter) Stat(ctx context.Context, key string) (snapshot.Info, error) {
	ctx, t, err := b.ms.TransactionContext(ctx, false)
	if err != nil {
		return snapshot.Info{}, err
	}
	defer t.Rollback()
	_, info, _, err := storage.GetInfo(ctx, key)
	if err != nil {
		return snapshot.Info{}, err
	}

	return info, nil
}

// Usage retrieves the disk usage of the top-level snapshot.
func (b *snapshotter) Usage(ctx context.Context, key string) (snapshot.Usage, error) {
	panic("not implemented")
}

// Walk the committed snapshots.
func (b *snapshotter) Walk(ctx context.Context, fn func(context.Context, snapshot.Info) error) error {
	ctx, t, err := b.ms.TransactionContext(ctx, false)
	if err != nil {
		return err
	}
	defer t.Rollback()
	return storage.WalkInfo(ctx, fn)
}

func (b *snapshotter) Prepare(ctx context.Context, key, parent string) ([]containerd.Mount, error) {
	return b.makeActive(ctx, key, parent, false)
}

func (b *snapshotter) View(ctx context.Context, key, parent string) ([]containerd.Mount, error) {
	return b.makeActive(ctx, key, parent, true)
}

func (b *snapshotter) makeActive(ctx context.Context, key, parent string, readonly bool) ([]containerd.Mount, error) {
	ctx, t, err := b.ms.TransactionContext(ctx, true)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil && t != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("Failure rolling back transaction")
			}
		}
	}()

	a, err := storage.CreateActive(ctx, key, parent, readonly)
	if err != nil {
		return nil, err
	}

	targetName := filepath.Join(b.dataset.Name, a.ID)
	var (
		target  *zfs.Dataset
		parent0 *zfs.Dataset
	)
	if len(a.ParentIDs) == 0 {
		target, err = createFilesystem(targetName)
		if err != nil {
			return nil, err
		}
	} else {
		parent0Name := filepath.Join(b.dataset.Name, a.ParentIDs[0]) + "@" + snapshotSuffix
		parent0, err = zfs.GetDataset(parent0Name)
		if err != nil {
			return nil, err
		}
		target, err = cloneFilesystem(targetName, parent0)
		if err != nil {
			return nil, err
		}
	}

	err = t.Commit()
	t = nil
	if err != nil {
		if derr := destroy(target); derr != nil {
			log.G(ctx).WithError(derr).WithField("targetName", targetName).Error("failed to delete dataset")
		}
		return nil, err
	}
	return b.mounts(target)
}

func (b *snapshotter) mounts(dataset *zfs.Dataset) ([]containerd.Mount, error) {
	return []containerd.Mount{
		{
			Type:    "zfs",
			Source:  dataset.Name,
			Options: nil, // TODO
		},
	}, nil
}

func (b *snapshotter) Commit(ctx context.Context, name, key string) (err error) {
	ctx, t, err := b.ms.TransactionContext(ctx, true)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil && t != nil {
			if rerr := t.Rollback(); rerr != nil {
				log.G(ctx).WithError(rerr).Warn("Failure rolling back transaction")
			}
		}
	}()

	id, err := storage.CommitActive(ctx, key, name, snapshot.Usage{})
	if err != nil {
		return errors.Wrap(err, "failed to commit")
	}

	activeName := filepath.Join(b.dataset.Name, id)
	active, err := zfs.GetDataset(activeName)
	if err != nil {
		return err
	}
	snapshot, err := active.Snapshot(snapshotSuffix, false)
	if err != nil {
		return err
	}
	err = t.Commit()
	t = nil
	if err != nil {
		snapshotName := activeName + "@" + snapshotSuffix
		if derr := destroy(snapshot); derr != nil {
			log.G(ctx).WithError(derr).WithField("snapshotName", snapshotName).Error("failed to delete dataset")
		}
		return err
	}
	return nil
}

// Mounts returns the mounts for the transaction identified by key. Can be
// called on an read-write or readonly transaction.
//
// This can be used to recover mounts after calling View or Prepare.
func (b *snapshotter) Mounts(ctx context.Context, key string) ([]containerd.Mount, error) {
	ctx, t, err := b.ms.TransactionContext(ctx, false)
	if err != nil {
		return nil, err
	}
	a, err := storage.GetActive(ctx, key)
	t.Rollback()
	if err != nil {
		return nil, errors.Wrap(err, "failed to get active snapshot")
	}
	activeName := filepath.Join(b.dataset.Name, a.ID)
	active, err := zfs.GetDataset(activeName)
	if err != nil {
		return nil, err
	}
	return b.mounts(active)
}

// Remove abandons the transaction identified by key. All resources
// associated with the key will be removed.
func (b *snapshotter) Remove(ctx context.Context, key string) (err error) {
	return fmt.Errorf("unimplemented")
}
