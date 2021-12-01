# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![GitHub Actions](https://github.com/bio-routing/bio-rd/workflows/tests/badge.svg)](https://github.com/bio-routing/bio-rd/actions)
[![Codecov](https://codecov.io/gh/bio-routing/bio-rd/branch/master/graph/badge.svg)](https://codecov.io/gh/bio-routing/bio-rd)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

### Build the examples

#### BGP

```bash
cd examples/bgp/ && go build
```

#### BMP

```bash
cd examples/bmp/ && go build
```

#### Device

```bash
cd examples/device && go build
```

### Run Tests

```bash
go test -v -cover ./...
```

### Update modules

```bash
go mod tidy
```

## Benchmarks

The benchmarks can be found in the [bio-rd-benchmarks](/bio-routing/bio-rd-benchmarks)
repository.
