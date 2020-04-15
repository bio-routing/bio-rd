#!/bin/bash

result="$(gofmt -e -s -l . 2>&1 | grep -v '^vendor/' )"
if [ -n "$result" ]; then
  echo "Go code is not formatted, run 'gofmt -e -s -w .'" >&2
  echo "$result"
  exit 1
else
  echo "Go code is formatted well"
fi
