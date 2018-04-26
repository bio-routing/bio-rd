# tflow2


[![Build Status](https://travis-ci.org/taktv6/tflow2.svg?branch=master)](https://travis-ci.org/taktv6/tflow2)
[![Coverage Status](https://coveralls.io/repos/taktv6/tflow2/badge.svg?branch=master&service=github)](https://coveralls.io/github/taktv6/tflow2?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/taktv6/tflow2)](http://goreportcard.com/report/taktv6/tflow2)

tflow2 is an in memory netflow version 9, IPFIX and Sflow analyzer.
It is designed for fast arbitrary queries and exports data to [Prometheus](https://prometheus.io/).

## Usage

Quick install with `go get -u github.com/taktv6/tflow2`
and `go build github.com/taktv6/tflow2`
or download a pre-built binary from the
[releases page](https://github.com/taktv6/tflow2/releases).

The release binaries have an additional command, `tflow2 -version`,
which reports the release version.

Once you start the main binary it will start reading netflow version 9 packets
on port 2055 UDP and IPFIX packets on port 4739 on all interfaces.
For user interaction it starts a webserver on port 4444 TCP on all interfaces. 

The webinterface allows you to run queries against the collected data.
Start time and router are mandatory criteria. If you don't provide any of
these you will always receive an empty result.

### Config file

There is YAML file as config. Defaults can be found in config.yml.example.
You'll at least need to add your Netflow/IPFIX/Sflow agents and adjust (if you don't 
want to work with interface IDs) your SNMP RO community.

### Command line arguments

`-alsologtostderr`

  Will send logs to stderr on top.

`-channelBuffer=int`

  This is the amount of elements that any channel within the program can buffer.

`-dbaddworkers=int`

  This is the amount of workers that are used to add flows into the in memory
  database.

`-log_backtrace_at`

  when logging hits line file:N, emit a stack trace (default :0).

`-log_dir`

  If non-empty, write log files in this directory.

`-logtostderr`

  log to standard error instead of files.

`-samplerate=int`

  Samplerate of your routers. This is used to deviate real packet and volume rates
  in case you use sampling.

`-sockreaders=int`

  Num of go routines reading and parsing netflow packets (default 24).

`-stderrthreshold`

  logs at or above this threshold go to stderr.

`-v value`

  log level for V logs.

`-vmodule value`

  comma-separated list of pattern=N settings for file-filtered logging.

## Limitations

Please be aware this software is not platform indipendent. It will only work
on little endian machines (such as x86)

## License

(c) Google, EXARING, Oliver Herms, 2017. Licensed under [Apache-2](LICENSE) license.

This is not an official Google product.
