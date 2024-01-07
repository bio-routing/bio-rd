#!/usr/bin/make
#
# Bio-Routing Makefile
#
# Maximilian Wilhelm <max@sdn.clinic>
#  --  Tue 02 Jan 2024 08:21:17 PM CET
#

expand_binary = $(dir)/$(notdir $(dir))

CMD_DIRS := $(wildcard cmd/*)
CMDS := $(foreach dir, $(CMD_DIRS), $(expand_binary))
EXAMPLE_DIRS = $(wildcard examples/*)
EXAMPLES := $(foreach dir, $(EXAMPLE_DIRS), $(expand_binary))

%:
	cd $(dir $(@)) && go build


build: $(CMDS) $(EXAMPLES)

all: clean build test

clean:
	rm -f -- $(CMDS) $(EXMAPLES)

test:
	@echo "Running tests..."
	go test ./...

test-coverage:
	go test -v -cover -coverprofile=coverage.txt ./...

.PHONY: all build clean test test-coverage