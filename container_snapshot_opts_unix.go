// +build !windows

package containerd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"github.com/containerd/containerd/mount"
)

// RemapSnapshotModifier creates a new snapshot and remaps the uid/gid for the
// filesystem to be used by a container with user namespaces
type RemapSnapshotModifier struct {
	UID uint32
	GID uint32
}

// ID implements SnapshotModifier
func (m *RemapSnapshotModifier) ID() string {
	return fmt.Sprintf("remap-%d-%d", m.UID, m.GID)
}

// Modify implements SnapshotModifier
func (m *RemapSnapshotModifier) Modify(mounts []mount.Mount) error {
	root, err := ioutil.TempDir("", "ctrd-snapshot-modifier-remap")
	if err != nil {
		return err
	}
	defer os.RemoveAll(root)
	if err := mount.All(mounts, root); err != nil {
		return err
	}
	defer mount.Unmount(root, 0)
	return filepath.Walk(root, m.incrementFS(root, m.UID, m.GID))
}

func (m *RemapSnapshotModifier) incrementFS(root string, uidInc, gidInc uint32) filepath.WalkFunc {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var (
			stat = info.Sys().(*syscall.Stat_t)
			u, g = int(stat.Uid + uidInc), int(stat.Gid + gidInc)
		)
		// be sure the lchown the path as to not de-reference the symlink to a host file
		return os.Lchown(path, u, g)
	}
}
