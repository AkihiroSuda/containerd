// +build windows

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

package defaults

import (
	"os"
	"path/filepath"
)

var (
	// DefaultRootDir is the default location used by containerd to store
	// persistent data
	DefaultRootDir = filepath.Join(os.Getenv("programfiles"), "containerd", "root")
	// DefaultStateDir is the default location used by containerd to store
	// transient data
	DefaultStateDir = filepath.Join(os.Getenv("programfiles"), "containerd", "state")
)

const (
	// DefaultAddress is the default winpipe address
	DefaultAddress = `\\.\pipe\containerd-containerd`
	// DefaultDebugAddress is the default winpipe address for pprof data
	DefaultDebugAddress = `\\.\pipe\containerd-debug`
	// DefaultFIFODir is the default location used by client-side cio library
	// to store FIFOs. Unused on Windows.
	DefaultFIFODir = ""
)

// UserRootDir returns DefaultRootDir
func UserRootDir() string {
	return DefaultRootDir
}

// UserStateDir returns DefaultStateDir
func UserStateDir() string {
	return DefaultStateDir
}

// UserAddress returns DefaultAddress
func UserAddress() string {
	return DefaultAddress
}

// UserDebugAddress returns DefaultDebugAddress
func UserDebugAddress() string {
	return DefaultDebugAddress
}

// UserFIFODir returns DefaultFIFODir
func UserFIFODir() string {
	return DefaultFIFODir
}
