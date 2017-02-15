package snapshot

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/docker/containerd"
	"github.com/docker/containerd/fs/fstest"
	"github.com/docker/containerd/testutil"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// Kind identifies the kind of snapshot.
type Kind int

// definitions of snapshot kinds
const (
	KindActive Kind = iota
	KindCommitted
)

// Info provides information about a particular snapshot.
type Info struct {
	Name     string // name or key of snapshot
	Parent   string // name of parent snapshot
	Kind     Kind   // active or committed snapshot
	Readonly bool   // true if readonly, only valid for active
}

// Snapshotter defines the methods required to implement a snapshot snapshotter for
// allocating, snapshotting and mounting filesystem changesets. The model works
// by building up sets of changes with parent-child relationships.
//
// A snapshot represents a filesystem state. Every snapshot has a parent, where
// the empty parent is represented by the empty string. A diff can be taken
// between a parent and its snapshot to generate a classic layer.
//
// An active snapshot is created by calling `Prepare`. After mounting, changes
// can be made to the snapshot. The act of commiting creates a committed
// snapshot. The committed snapshot will get the parent of active snapshot. The
// committed snapshot can then be used as a parent. Active snapshots can never
// act as a parent.
//
// Snapshots are best understood by their lifecycle. Active snapshots are
// always created with Prepare or View. Committed snapshots are always created
// with Commit.  Active snapshots never become committed snapshots and vice
// versa. All snapshots may be removed.
//
// For consistency, we define the following terms to be used throughout this
// interface for snapshotter implementations:
//
// 	`key` - refers to an active snapshot
// 	`name` - refers to a committed snapshot
// 	`parent` - refers to the parent in relation
//
// Most methods take various combinations of these identifiers. Typically,
// `name` and `parent` will be used in cases where a method *only* takes
// committed snapshots. `key` will be used to refer to active snapshots in most
// cases, except where noted. All variables used to access snapshots use the
// same key space. For example, an active snapshot may not share the same key
// with a committed snapshot.
//
// We cover several examples below to demonstrate the utility of a snapshot
// snapshotter.
//
// Importing a Layer
//
// To import a layer, we simply have the Snapshotter provide a list of
// mounts to be applied such that our dst will capture a changeset. We start
// out by getting a path to the layer tar file and creating a temp location to
// unpack it to:
//
//	layerPath, tmpDir := getLayerPath(), mkTmpDir() // just a path to layer tar file.
//
// We start by using a Snapshotter to Prepare a new snapshot transaction, using a
// key and descending from the empty parent "":
//
//	mounts, err := snapshotter.Prepare(key, "")
// 	if err != nil { ... }
//
// We get back a list of mounts from Snapshotter.Prepare, with the key identifying
// the active snapshot. Mount this to the temporary location with the
// following:
//
//	if err := MountAll(mounts, tmpDir); err != nil { ... }
//
// Once the mounts are performed, our temporary location is ready to capture
// a diff. In practice, this works similar to a filesystem transaction. The
// next step is to unpack the layer. We have a special function unpackLayer
// that applies the contents of the layer to target location and calculates the
// DiffID of the unpacked layer (this is a requirement for docker
// implementation):
//
//	layer, err := os.Open(layerPath)
//	if err != nil { ... }
// 	digest, err := unpackLayer(tmpLocation, layer) // unpack into layer location
// 	if err != nil { ... }
//
// When the above completes, we should have a filesystem the represents the
// contents of the layer. Careful implementations should verify that digest
// matches the expected DiffID. When completed, we unmount the mounts:
//
//	unmount(mounts) // optional, for now
//
// Now that we've verified and unpacked our layer, we commit the active
// snapshot to a name. For this example, we are just going to use the layer
// digest, but in practice, this will probably be the ChainID:
//
//	if err := snapshotter.Commit(digest.String(), key); err != nil { ... }
//
// Now, we have a layer in the Snapshotter that can be accessed with the digest
// provided during commit. Once you have committed the snapshot, the active
// snapshot can be removed with the following:
//
// 	snapshotter.Remove(key)
//
// Importing the Next Layer
//
// Making a layer depend on the above is identical to the process described
// above except that the parent is provided as parent when calling
// Manager.Prepare, assuming a clean tmpLocation:
//
// 	mounts, err := snapshotter.Prepare(tmpLocation, parentDigest)
//
// We then mount, apply and commit, as we did above. The new snapshot will be
// based on the content of the previous one.
//
// Running a Container
//
// To run a container, we simply provide Snapshotter.Prepare the committed image
// snapshot as the parent. After mounting, the prepared path can
// be used directly as the container's filesystem:
//
// 	mounts, err := snapshotter.Prepare(containerKey, imageRootFSChainID)
//
// The returned mounts can then be passed directly to the container runtime. If
// one would like to create a new image from the filesystem, Manager.Commit is
// called:
//
// 	if err := snapshotter.Commit(newImageSnapshot, containerKey); err != nil { ... }
//
// Alternatively, for most container runs, Snapshotter.Remove will be called to
// signal the Snapshotter to abandon the changes.
type Snapshotter interface {
	// Stat returns the info for an active or committed snapshot by name or
	// key.
	//
	// Should be used for parent resolution, existence checks and to discern
	// the kind of snapshot.
	Stat(ctx context.Context, key string) (Info, error)

	// Mounts returns the mounts for the active snapshot transaction identified
	// by key. Can be called on an read-write or readonly transaction. This is
	// available only for active snapshots.
	//
	// This can be used to recover mounts after calling View or Prepare.
	Mounts(ctx context.Context, key string) ([]containerd.Mount, error)

	// Prepare creates an active snapshot identified by key descending from the
	// provided parent.  The returned mounts can be used to mount the snapshot
	// to capture changes.
	//
	// If a parent is provided, after performing the mounts, the destination
	// will start with the content of the parent. The parent must be a
	// committed snapshot. Changes to the mounted destination will be captured
	// in relation to the parent. The default parent, "", is an empty
	// directory.
	//
	// The changes may be saved to a committed snapshot by calling Commit. When
	// one is done with the transaction, Remove should be called on the key.
	//
	// Multiple calls to Prepare or View with the same key should fail.
	Prepare(ctx context.Context, key, parent string) ([]containerd.Mount, error)

	// View behaves identically to Prepare except the result may not be
	// committed back to the snapshot snapshotter. View returns a readonly view on
	// the parent, with the active snapshot being tracked by the given key.
	//
	// This method operates identically to Prepare, except that Mounts returned
	// may have the readonly flag set. Any modifications to the underlying
	// filesystem will be ignored. Implementations may perform this in a more
	// efficient manner that differs from what would be attempted with
	// `Prepare`.
	//
	// Commit may not be called on the provided key and will return an error.
	// To collect the resources associated with key, Remove must be called with
	// key as the argument.
	View(ctx context.Context, key, parent string) ([]containerd.Mount, error)

	// Commit captures the changes between key and its parent into a snapshot
	// identified by name.  The name can then be used with the snapshotter's other
	// methods to create subsequent snapshots.
	//
	// A committed snapshot will be created under name with the parent of the
	// active snapshot.
	//
	// Commit may be called multiple times on the same key. Snapshots created
	// in this manner will all reference the parent used to start the
	// transaction.
	Commit(ctx context.Context, name, key string) error

	// Remove the committed or active snapshot by the provided key.
	//
	// All resources associated with the key will be removed.
	//
	// If the snapshot is a parent of another snapshot, its children must be
	// removed before proceeding.
	Remove(ctx context.Context, key string) error

	// Walk the committed snapshots. For each snapshot in the snapshotter, the
	// function will be called.
	Walk(ctx context.Context, fn func(context.Context, Info) error) error
}

// SnapshotterSuite runs a test suite on the snapshotter given a factory function.
func SnapshotterSuite(t *testing.T, name string, snapshotterFn func(root string) (Snapshotter, func(), error)) {
	t.Run("Basic", makeTest(t, name, snapshotterFn, checkSnapshotterBasic))
	t.Run("StatActive", makeTest(t, name, snapshotterFn, checkSnapshotterStatActive))
	t.Run("StatComitted", makeTest(t, name, snapshotterFn, checkSnapshotterStatCommitted))
	t.Run("TransitivityTest", makeTest(t, name, snapshotterFn, checkSnapshotterTransitivity))

}

func makeTest(t *testing.T, name string, snapshotterFn func(root string) (Snapshotter, func(), error), fn func(t *testing.T, snapshotter Snapshotter, work string)) func(t *testing.T) {
	return func(t *testing.T) {
		oldumask := syscall.Umask(0)
		defer syscall.Umask(oldumask)
		// Make two directories: a snapshotter root and a play area for the tests:
		//
		// 	/tmp
		// 		work/ -> passed to test functions
		// 		root/ -> passed to snapshotter
		//
		tmpDir, err := ioutil.TempDir("", "snapshot-suite-"+name+"-")
		if err != nil {
			t.Fatal(err)
		}
		defer os.RemoveAll(tmpDir)

		root := filepath.Join(tmpDir, "root")
		if err := os.MkdirAll(root, 0777); err != nil {
			t.Fatal(err)
		}

		snapshotter, cleanup, err := snapshotterFn(root)
		if err != nil {
			t.Fatal(err)
		}
		defer cleanup()

		work := filepath.Join(tmpDir, "work")
		if err := os.MkdirAll(work, 0777); err != nil {
			t.Fatal(err)
		}

		defer testutil.DumpDir(t, tmpDir)
		fn(t, snapshotter, work)
	}
}

func tFatal(t *testing.T, err error) {
	t.Fatal(errors.WithStack(err))
}

// checkSnapshotterBasic tests the basic workflow of a snapshot snapshotter.
func checkSnapshotterBasic(t *testing.T, snapshotter Snapshotter, work string) {
	ctx := context.TODO()

	initialApplier := fstest.Apply(
		fstest.CreateFile("/foo", []byte("foo\n"), 0777),
		fstest.CreateDir("/a", 0755),
		fstest.CreateDir("/a/b", 0755),
		fstest.CreateDir("/a/b/c", 0755),
	)

	diffApplier := fstest.Apply(
		fstest.CreateFile("/bar", []byte("bar\n"), 0777),
		// also, change content of foo to bar
		fstest.CreateFile("/foo", []byte("bar\n"), 0777),
		fstest.RemoveAll("/a/b"),
	)

	preparing := filepath.Join(work, "preparing")
	if err := os.MkdirAll(preparing, 0777); err != nil {
		tFatal(t, err)
	}

	mounts, err := snapshotter.Prepare(ctx, preparing, "")
	if err != nil {
		tFatal(t, err)
	}

	if len(mounts) < 1 {
		tFatal(t, errors.New("expected mounts to have entries"))
	}

	if err := containerd.MountAll(mounts, preparing); err != nil {
		tFatal(t, err)
	}
	defer testutil.Unmount(t, preparing)

	if err := initialApplier.Apply(preparing); err != nil {
		tFatal(t, err)
	}

	committed := filepath.Join(work, "committed")
	if err := snapshotter.Commit(ctx, committed, preparing); err != nil {
		tFatal(t, err)
	}

	si, err := snapshotter.Stat(ctx, committed)
	if err != nil {
		tFatal(t, err)
	}

	assert.Equal(t, "", si.Parent)
	assert.Equal(t, KindCommitted, si.Kind)

	next := filepath.Join(work, "nextlayer")
	if err := os.MkdirAll(next, 0777); err != nil {
		tFatal(t, err)
	}

	mounts, err = snapshotter.Prepare(ctx, next, committed)
	if err != nil {
		tFatal(t, err)
	}
	if err := containerd.MountAll(mounts, next); err != nil {
		tFatal(t, err)
	}
	defer testutil.Unmount(t, next)

	if err := fstest.CheckDirectoryEqualWithApplier(next, initialApplier); err != nil {
		tFatal(t, err)
	}

	if err := diffApplier.Apply(next); err != nil {
		tFatal(t, err)
	}

	ni, err := snapshotter.Stat(ctx, next)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, committed, ni.Parent)
	assert.Equal(t, KindActive, ni.Kind)

	nextCommitted := filepath.Join(work, "committed-next")
	if err := snapshotter.Commit(ctx, nextCommitted, next); err != nil {
		tFatal(t, err)
	}

	si2, err := snapshotter.Stat(ctx, nextCommitted)
	if err != nil {
		tFatal(t, err)
	}

	assert.Equal(t, committed, si2.Parent)
	assert.Equal(t, KindCommitted, si2.Kind)

	expected := map[string]Info{
		si.Name:  si,
		si2.Name: si2,
	}
	walked := map[string]Info{} // walk is not ordered
	assert.NoError(t, snapshotter.Walk(ctx, func(ctx context.Context, si Info) error {
		walked[si.Name] = si
		return nil
	}))

	assert.Equal(t, expected, walked)

	nextnext := filepath.Join(work, "nextnextlayer")
	if err := os.MkdirAll(nextnext, 0777); err != nil {
		tFatal(t, err)
	}

	mounts, err = snapshotter.View(ctx, nextnext, nextCommitted)
	if err != nil {
		tFatal(t, err)
	}
	if err := containerd.MountAll(mounts, nextnext); err != nil {
		tFatal(t, err)
	}
	defer testutil.Unmount(t, nextnext)

	if err := fstest.CheckDirectoryEqualWithApplier(nextnext,
		fstest.Apply(initialApplier, diffApplier)); err != nil {
		tFatal(t, err)
	}
}
