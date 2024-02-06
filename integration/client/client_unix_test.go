//go:build !windows

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

package client

import (
	"github.com/containerd/containerd/v2/integration/images"
)

var (
	testImage             = images.Get(images.BusyBox)
	testMultiLayeredImage = images.Get(images.VolumeCopyUp)
	shortCommand          = withProcessArgs("true")
	// NOTE: The TestContainerPids needs two running processes in one
	// container. But busybox:1.36 sh shell, the `sleep` is a builtin.
	//
	// 	/bin/sh -c "type sleep"
	//      sleep is a shell builtin
	//
	// We should use `/bin/sleep` instead of `sleep`. And busybox sh shell
	// will execve directly instead of clone-execve if there is only one
	// command. There will be only one process in container if we use
	// '/bin/sh -c "/bin/sleep inf"'.
	//
	// So we append `&& exit 0` to force sh shell uses clone-execve.
	longCommand = withProcessArgs("/bin/sh", "-c", "/bin/sleep inf && exit 0")
)
