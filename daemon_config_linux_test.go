/*
   Copyright The containerd Authors.

   Licensed under the Apache License, Version 2.0 (the "License");
   you may not use this file except in compliance with the License.
   You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

   Unless required by applicable law or agreed to in writing, software
   distributed under the License is distributed on an "AS IS" BASIS,
   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
   See the License for the specific language governing permissions and
   limitations under the License.
*/

package containerd

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/containerd/containerd/oci"
)

func testDaemonRuntimeRoot(t *testing.T, noShim bool) {
	runtimeRoot, err := ioutil.TempDir("", "containerd-test-runtime-root")
	if err != nil {
		t.Fatal(err)
	}
	configTOML := fmt.Sprintf(`
[plugins]
 [plugins.linux]
   no_shim = %v
   runtime_root = "%s"
`, noShim, runtimeRoot)

	client, _, cleanup := newDaemonWithConfig(t, configTOML)
	defer cleanup()

	ctx, cancel := testContext()
	defer cancel()
	// FIXME(AkihiroSuda): import locally frozen image?
	image, err := client.Pull(ctx, testImage, WithPullUnpack)
	if err != nil {
		t.Fatal(err)
	}

	id := t.Name()
	container, err := client.NewContainer(ctx, id, WithNewSpec(oci.WithImageConfig(image), withProcessArgs("top")), WithNewSnapshot(id, image))
	if err != nil {
		t.Fatal(err)
	}
	defer container.Delete(ctx, WithSnapshotCleanup)

	task, err := container.NewTask(ctx, empty())
	if err != nil {
		t.Fatal(err)
	}
	defer task.Delete(ctx)

	if err := task.Start(ctx); err != nil {
		t.Fatal(err)
	}

	stateJSONPath := filepath.Join(runtimeRoot, testNamespace, id, "state.json")
	_, err = os.Stat(stateJSONPath)
	if err != nil {
		t.Errorf("error while getting stat for %s: %v", stateJSONPath, err)
	}

	finishedC, err := task.Wait(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := task.Kill(ctx, syscall.SIGKILL); err != nil {
		t.Error(err)
	}
	<-finishedC
}

// TestDaemonRuntimeRoot ensures plugin.linux.runtime_root is not ignored
func TestDaemonRuntimeRoot(t *testing.T) {
	testDaemonRuntimeRoot(t, false)
}

// TestDaemonRuntimeRootNoShim ensures plugin.linux.runtime_root is not ignored when no_shim is true
func TestDaemonRuntimeRootNoShim(t *testing.T) {
	t.Skip("no_shim is not functional now: https://github.com/containerd/containerd/issues/2181")
	testDaemonRuntimeRoot(t, true)
}
