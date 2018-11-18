# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![Build Status](https://travis-ci.org/bio-routing/bio-rd.svg?branch=master)](https://travis-ci.org/bio-routing/bio-rd)
[![Coverage Status](https://coveralls.io/repos/bio-routing/bio-rd/badge.svg?branch=master&service=github)](https://coveralls.io/github/bio-routing/bio-rd?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

### Build the examples

#### BGP

    cd examples/bgp/ && go build

#### BMP

    cd examples/bmp/ && go build

### Run Tests

    go test -v -cover ./...

### Update vendor/dependencies

#### Update vendor/add dependencies

    dep ensure
