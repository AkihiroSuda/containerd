// +build !windows

package fstest

import (
	"path/filepath"
	"time"

	"github.com/containerd/continuity/sysx"
	"golang.org/x/sys/unix"
)

// SetXAttr sets the xatter for the file
func SetXAttr(name, key, value string) Applier {
	return applyFn(func(root string) error {
		return sysx.LSetxattr(name, key, []byte(value), 0)
	})
}

// ChtimeNoFollow changes access and mod time of file without following symlink
func ChtimeNoFollow(name string, t time.Time) Applier {
	return applyFn(func(root string) error {
		path := filepath.Join(root, name)
		atime := unix.NsecToTimespec(t.UnixNano())
		mtime := unix.NsecToTimespec(t.UnixNano())
		utimes := [2]unix.Timespec{atime, mtime}
		return unix.UtimesNanoAt(unix.AT_FDCWD, path, utimes[0:], unix.AT_SYMLINK_NOFOLLOW)
	})

}
