#!/bin/sh
dir=$(pwd)
echo "Switching to GOPATH"
cd "$GOPATH/src"
echo "Generating protobuf code"
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/net/api/*.proto
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/route/api/*.proto
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/protocols/bgp/api/*.proto
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/cmd/ris/api/*.proto
echo "Switching back to working directory"
cd $dir
