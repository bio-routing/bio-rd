# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![CircleCI](https://circleci.com/gh/bio-routing/bio-rd/tree/master.svg?style=shield)](https://circleci.com/gh/bio-routing/bio-rd/tree/master)
[![Coverage Status](https://coveralls.io/repos/bio-routing/bio-rd/badge.svg?branch=master&service=github)](https://coveralls.io/github/bio-routing/bio-rd?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

### Build the examples

#### BGP

    cd examples/bgp/ && go build

#### BMP

    cd examples/bmp/ && go build

#### Netlink

    cd examples/netlink && go build

### Run Tests

    go test -v -cover ./...

### Update vendor/dependencies

#### Install `dep`

    curl https://raw.githubusercontent.com/golang/dep/master/install.sh | sh

Or on MacOS

    brew install dep

#### Update vendor/add dependencies

    dep ensure
