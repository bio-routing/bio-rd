
$(GOPATH)/bin/go-fuzz-build:
	go get -u github.com/dvyukov/go-fuzz/...

$(GOPATH)/bin/go-fuzz:
	go get -u github.com/dvyukov/go-fuzz/...

packet-fuzz.zip: $(GOPATH)/bin/go-fuzz-build
	$(GOPATH)/bin/go-fuzz-build github.com/bio-routing/bio-rd/protocols/bgp/packet

corpus:
	mkdir -p corpus
	go run gen_initial_corpus.go

fuzz: packet-fuzz.zip corpus $(GOPATH)/bin/go-fuzz
	$(GOPATH)/bin/go-fuzz -bin=./packet-fuzz.zip -workdir=.
