# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![Build Status](https://travis-ci.org/bio-routing/bio-rd.svg?branch=master)](https://travis-ci.org/bio-routing/bio-rd)
[![Coverage Status](https://coveralls.io/repos/bio-routing/bio-rd/badge.svg?branch=master&service=github)](https://coveralls.io/github/bio-routing/bio-rd?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)
[![Go Doc](https://godoc.org/github.com/bio-routing/bio-rd?status.svg)](https://godoc.org/github.com/bio-routing/bio-rd)

## Building

We use [Bazel](https://bazel.io) to build bio-rd.

### Build

    bazel build //:bio-rd
    bazel-bin/linux_amd64_stripped/bio-rd -arguments go -here

or

    bazel run //:bio-rd -- -arguments go -here

### Run Tests

    bazel test //...

### Update Bazel BUILD files

To regenerate BUILD files (for both the project and vendored libraries), you will need to run the following:

    bazel run //:gazelle -- update

Be sure to commit the changes.

### Update vendor/dependencies

#### build `dep`

    bazel build //vendor/github.com/golang/dep/cmd/dep

#### Update vendor/add dependencies

    bazel-bin/vendor/github.com/golang/dep/cmd/dep/linux_amd64_stripped/dep ensure

dep of dep breaks gazelle. Therefore execute the following commands after updating Gopkg.toml

    rm -rf vendor/github.com/golang/dep/cmd/dep/testdata
    rm -rf vendor/github.com/golang/dep/internal/fs/testdata/symlinks/dir-symlink
