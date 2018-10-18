# bio-rd

A re-implementation of BGP, IS-IS and OSPF in go. We value respect and robustness!

[![Build Status](https://travis-ci.org/bio-routing/bio-rd.svg?branch=master)](https://travis-ci.org/bio-routing/bio-rd)
[![Coverage Status](https://coveralls.io/repos/bio-routing/bio-rd/badge.svg?branch=master&service=github)](https://coveralls.io/github/bio-routing/bio-rd?branch=master)
[![Go ReportCard](http://goreportcard.com/badge/bio-routing/bio-rd)](http://goreportcard.com/report/bio-routing/bio-rd)

Building
========

We use [Bazel](https://bazel.io) to test bio-rd.

Build
-----
 
```
bazel build //...
```
Now you can find binaries under bazel-bin/examples/


Run Tests
---------

    bazel test //...


Update Bazel BUILD files
------------------------

To regenerate BUILD files (for both the project and vendored libraries), you will need to run the following:

    bazel run //:gazelle -- update

Be sure to commit the changes.

Update vendor/dependencies
---------------------------

After updating Gopkg.toml, run

    bazel build //vendor/github.com/golang/dep/cmd/dep
    bazel-bin/vendor/github.com/golang/dep/cmd/dep/linux_amd64_stripped/dep use
    # hack: dep of dep gives us these, and it breaks gazelle
    rm -rf vendor/github.com/golang/dep/cmd/dep/testdata
    rm -rf vendor/github.com/golang/dep/internal/fs/testdata/symlinks/dir-symlink
