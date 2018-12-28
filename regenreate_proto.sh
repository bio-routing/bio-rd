#!/bin/sh
pushd $GOPATH/src
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/net/api/*.proto
protoc --go_out=plugins=grpc:. github.com/bio-routing/bio-rd/protocols/bgp/api/*.proto
popd
