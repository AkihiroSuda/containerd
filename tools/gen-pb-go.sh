#!/bin/bash

# This script generates "*.pb.go" files.
# This script should be called only from Makefile

set -e

validate_dir() {
    dir=$1
    parent=$(dirname $dir)
    ( [ $parent = api/types ] || [ $parent = api/services ] ) || { echo "Unexpected dir ${dir}"; exit 1; }
    [ $(find $dir -name '*.proto' | wc -l) -eq 1 ] || { echo "${dir} has unexpected number of proto files"; exit 1; }
}

repo=$(pwd)
[ $(basename $repo) = containerd ] || { echo "Unexpected cwd ${repo}"; exit 1; }
protos=$(find api -name '*.proto')
for proto in $protos; do
    dir=$(dirname $proto)
    validate_dir $dir
    proto_base=$(basename $proto)
    pkg="github.com/docker/containerd/${dir}"
    pkg_base=$(basename $dir)
    (
	cd $dir
	protoc -I.:${repo}/vendor:${repo}/vendor/github.com/gogo/protobuf:${repo}/../../..:/usr/local/include --gogoctrd_out=plugins=grpc,import_path=${pkg},Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. ${proto_base}
    )
done
