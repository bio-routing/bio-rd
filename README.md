# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![CircleCI](https://circleci.com/gh/bio-routing/bio-rd/tree/master.svg?style=shield)](https://circleci.com/gh/bio-routing/bio-rd/tree/master)
[![Codecov](https://codecov.io/gh/bio-routing/bio-rd/branch/master/graph/badge.svg)](https://codecov.io/gh/bio-routing/bio-rd)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

### Build the examples

#### BGP

    cd examples/bgp/ && go build

#### BMP

    cd examples/bmp/ && go build

#### Device

    cd examples/device && go build

#### FIB
See the [README](examples/fib/README.md) in the [examples/fib](examples/fib) folder for more information

    cd examples/fib && go build

### Run Tests

    go test -v -cover ./...

### Update modules

    go mod tidy
