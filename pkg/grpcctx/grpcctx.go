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

package grpcctx

import (
	"context"

	"google.golang.org/grpc/metadata"
)

func WithGRPCNamespaceHeader(ctx context.Context, k, v string) context.Context {
	// also store on the grpc headers so it gets picked up by any clients that
	// are using this.
	nsheader := metadata.Pairs(k, v)
	md, ok := metadata.FromOutgoingContext(ctx) // merge with outgoing context.
	if !ok {
		md = nsheader
	} else {
		// order ensures the latest is first in this list.
		md = metadata.Join(nsheader, md)
	}

	return metadata.NewOutgoingContext(ctx, md)
}

func FromGRPCHeader(ctx context.Context, k string) (string, bool) {
	// try to extract for use in grpc servers.
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		// TODO(stevvooe): Check outgoing context?
		return "", false
	}

	values := md[k]
	if len(values) == 0 {
		return "", false
	}

	return values[0], true
}
