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

const (
	// DefaultMaxRecvMsgSize defines the default maximum message size for
	// receiving protobufs passed over the GRPC API.
	DefaultMaxRecvMsgSize = 16 << 20
	// DefaultMaxSendMsgSize defines the default maximum message size for
	// sending protobufs passed over the GRPC API.
	DefaultMaxSendMsgSize = 16 << 20
)

var (
	// UserRootDir is typically set to "/home/$USER/.local/share/containerd" on Linux during init().
	UserRootDir = DefaultRootDir

	// UserStateDir is typically set to "/run/user/$UID/containerd" on Linux.
	// Typically this directory needs to be created with sticky bit.
	// See https://github.com/opencontainers/runc/issues/1694
	UserStateDir = DefaultStateDir

	// UserAddress is typically set to "/run/user/$UID/containerd/containerd.sock" on Linux during init().
	UserAddress = DefaultAddress

	// UserDebugAddress is typically set to "/run/user/$UID/containerd/debug.sock" on Linux during init().
	UserDebugAddress = DefaultDebugAddress

	// UserFIFODir is typically set to "/run/user/$UID/containerd/fifo" on Linux during init().
	UserFIFODir = DefaultFIFODir
)
