package mount

import (
	"strings"

	"golang.org/x/sys/unix"
)

const (
	// ptypes is the set propagation types.
	ptypes = unix.MS_SHARED | unix.MS_PRIVATE | unix.MS_SLAVE | unix.MS_UNBINDABLE

	// pflags is the full set valid flags for a change propagation call.
	pflags = ptypes | unix.MS_REC | unix.MS_SILENT

	// broflags is the combination of bind and read only
	broflags = unix.MS_BIND | unix.MS_RDONLY
)

// isremount returns true if either device name or flags identify a remount request, false otherwise.
func isremount(device string, flags int) bool {
	// We treat device "" and "none" as a remount request to provide compatibility with
	// requests that don't explicitly set MS_REMOUNT such as those manipulating bind mounts.
	return flags&unix.MS_REMOUNT != 0 || device == "" || device == "none"
}

func (m *Mount) Mount(target string) error {
	flags, data := parseMountOptions(m.Options)
	oflags := flags &^ ptypes
	if !isremount(m.Source, flags) || data != "" {
		// Initial call applying all non-propagation flags for mount
		// or remount with changed data
		if err := unix.Mount(m.Source, target, m.Type, uintptr(oflags), data); err != nil {
			return err
		}
	}

	if flags&ptypes != 0 {
		// Change the propagation type.
		if err := unix.Mount("", target, "", uintptr(flags&pflags), ""); err != nil {
			return err
		}
	}

	if flags&broflags == broflags {
		// Remount the bind to apply read only.
		return unix.Mount("", target, "", uintptr(flags|unix.MS_REMOUNT), "")
	}
	return nil
}

func Unmount(mount string, flags int) error {
	return unix.Unmount(mount, flags)
}

// UnmountAll repeatedly unmounts the given mount point until there
// are no mounts remaining (EINVAL is returned by mount), which is
// useful for undoing a stack of mounts on the same mount point.
func UnmountAll(mount string, flags int) error {
	for {
		if err := Unmount(mount, flags); err != nil {
			// EINVAL is returned if the target is not a
			// mount point, indicating that we are
			// done. It can also indicate a few other
			// things (such as invalid flags) which we
			// unfortunately end up squelching here too.
			if err == unix.EINVAL {
				return nil
			}
			return err
		}
	}
}

// parseMountOptions takes fstab style mount options and parses them for
// use with a standard mount() syscall
func parseMountOptions(options []string) (int, string) {
	var (
		flag int
		data []string
	)
	flags := map[string]struct {
		clear bool
		flag  int
	}{
		"async":         {true, unix.MS_SYNCHRONOUS},
		"atime":         {true, unix.MS_NOATIME},
		"bind":          {false, unix.MS_BIND},
		"defaults":      {false, 0},
		"dev":           {true, unix.MS_NODEV},
		"diratime":      {true, unix.MS_NODIRATIME},
		"dirsync":       {false, unix.MS_DIRSYNC},
		"exec":          {true, unix.MS_NOEXEC},
		"mand":          {false, unix.MS_MANDLOCK},
		"noatime":       {false, unix.MS_NOATIME},
		"nodev":         {false, unix.MS_NODEV},
		"nodiratime":    {false, unix.MS_NODIRATIME},
		"noexec":        {false, unix.MS_NOEXEC},
		"nomand":        {true, unix.MS_MANDLOCK},
		"norelatime":    {true, unix.MS_RELATIME},
		"nostrictatime": {true, unix.MS_STRICTATIME},
		"nosuid":        {false, unix.MS_NOSUID},
		"rbind":         {false, unix.MS_BIND | unix.MS_REC},
		"relatime":      {false, unix.MS_RELATIME},
		"remount":       {false, unix.MS_REMOUNT},
		"ro":            {false, unix.MS_RDONLY},
		"rw":            {true, unix.MS_RDONLY},
		"strictatime":   {false, unix.MS_STRICTATIME},
		"suid":          {true, unix.MS_NOSUID},
		"sync":          {false, unix.MS_SYNCHRONOUS},
	}
	for _, o := range options {
		// If the option does not exist in the flags table or the flag
		// is not supported on the platform,
		// then it is a data value for a specific fs type
		if f, exists := flags[o]; exists && f.flag != 0 {
			if f.clear {
				flag &^= f.flag
			} else {
				flag |= f.flag
			}
		} else {
			data = append(data, o)
		}
	}
	return flag, strings.Join(data, ",")
}
