package fstest

import (
	"time"

	"github.com/containerd/containerd/errdefs"
)

// ChtimeNoFollow changes access and mod time of file without following symlink
func ChtimeNoFollow(name string, t time.Time) Applier {
	return applyFn(func(root string) error {
		return errdefs.ErrNotImplemented
	})
}
