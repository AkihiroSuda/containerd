#!/bin/bash

# A work-around for generating gen.go for grpc packages.
# This should definitely go away when we came up with a better solution.

set -e

validate_dir() {
    dir=$1
    parent=$(dirname $dir)
    ( [ $parent = api/types ] || [ $parent = api/services ] ) || { echo "Unexpected dir ${dir}"; exit 1; }
    [ $(find $dir -name '*.proto' | wc -l) -eq 1 ] || { echo "${dir} has unexpected number of proto files"; exit 1; }
}

protos=$(find api -name '*.proto')
for proto in $protos; do
    dir=$(dirname $proto)
    validate_dir $dir
    gengo="${dir}/gen.go"
    proto_base=$(basename $proto)
    pkg="github.com/docker/containerd/${dir}"
    pkg_base=$(basename $dir)
    echo "Generating ${gengo}"
    cat <<EOF > $gengo
package ${pkg_base}

//go:generate protoc -I.:../../../vendor:../../../vendor/github.com/gogo/protobuf:../../../../../..:/usr/local/include --gogoctrd_out=plugins=grpc,import_path=${pkg},Mgogoproto/gogo.proto=github.com/gogo/protobuf/gogoproto,Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:. ${proto_base}
EOF
done
