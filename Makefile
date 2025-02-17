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

yamldoc-go:
	GOBIN=$(shell pwd)/bin/ go install github.com/projectdiscovery/yamldoc-go/cmd/docgen@main

yaml-docs: yamldoc-go
	bin/docgen cmd/bio-rd/config/config.go cmd/bio-rd/config/config_docs.go config
	bin/docgen cmd/bio-rd/config/policy.go cmd/bio-rd/config/policy_docs.go policy
	bin/docgen cmd/bio-rd/config/routing_options.go cmd/bio-rd/config/routing_options_docs.go routing_options
	bin/docgen cmd/bio-rd/config/routing_instance.go cmd/bio-rd/config/routing_instance_docs.go routing_instance
	bin/docgen cmd/bio-rd/config/protocols.go cmd/bio-rd/config/protocols_docs.go protocols
	bin/docgen cmd/bio-rd/config/static_route.go cmd/bio-rd/config/static_route_docs.go static_route
	bin/docgen cmd/bio-rd/config/bgp.go cmd/bio-rd/config/bgp_docs.go bgp
	bin/docgen cmd/bio-rd/config/isis.go cmd/bio-rd/config/isis_docs.go isis
	go run cmd/doc-gen/main.go
