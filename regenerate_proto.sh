#!/bin/sh
(
  cd third_party/tools
  ./build.sh
)

third_party/tools/bin/buf generate
