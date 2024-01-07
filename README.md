# Bio-Routing

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![GitHub Actions](https://github.com/bio-routing/bio-rd/workflows/tests/badge.svg)](https://github.com/bio-routing/bio-rd/actions)
[![Codecov](https://codecov.io/gh/bio-routing/bio-rd/branch/master/graph/badge.svg)](https://codecov.io/gh/bio-routing/bio-rd)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

To build Bio-Routing binares and/or examples you need Go installed and in your `$PATH`.
Currently the minimum supported Go version is v1.20.

To build all commands and examples, you can leverage the `Makefile`, if you have `make` installed, and run
```bash
make build
```

If you're only interested in one particular command/service or example, found in the `cmd/` or `examples/` sub-directories within this repository,
enter the respective directory on a shell and run `go build`.
You should get a binary named like the current directory, which you can run.

To build the BGP examples, this would look like
```bash
cd exmaples/bgp
go build
```

To build the `bio-rd` service binary, this would look like
```bash
cd cmd/bio-rd
go build
```

### Run Tests

```bash
go test -v -cover ./...
```

## Running bio-rd

`bio-rd` is the main binary which provides a configurable BGP speaker.
It supports the following command-line parameters:

    Usage of ./bio-rd:
      -bgp.listen-addr-ipv4 string
        	BGP listen address for IPv4 AFI (default "0.0.0.0:179")
      -bgp.listen-addr-ipv6 string
        	BGP listen address for IPv6 AFI (default "[::]:179")
      -config.file string
        	bio-rd config file (default "bio-rd.yml")
      -grpc_keepalive_min_time uint
        	Minimum time (seconds) for a client to wait between GRPC keepalive pings (default 1)
      -grpc_port uint
        	GRPC API server port (default 5566)
      -metrics_port uint
        	Metrics HTTP server port (default 55667)

You can find an [example configuration file](cmd/bio-rd/bio-rd.yml) within in `cmd/bio-rd` directory.

As `bio-rd` needs to listen on the priviledged TCP port 179 for BGP connections, you either need to start the service as `root` or using `sudo`, e.g.

```bash
$ sudo ./bio-rd -config.file bio-rd.yml
```

## Benchmarks

The benchmarks can be found in the [bio-routing/bio-rd-benchmarks](https://github.com/bio-routing/bio-rd-benchmarks)
repository.
