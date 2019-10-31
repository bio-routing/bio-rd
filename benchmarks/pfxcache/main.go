package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"runtime/pprof"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
)

func main() {
	pfxs := make([]*bnet.Prefix, 0)
	for i := 0; i < 255; i++ {
		for j := 0; j < 255; j++ {
			for k := 0; k < 11; k++ {
				addr := bnet.IPv4FromOctets(uint8(k)+1, uint8(i), uint8(j), 0)
				addr.Dedup()

				pfxs = append(pfxs, bnet.NewPfx(addr, 24).Dedup())
			}
		}
	}

	buf := bytes.NewBuffer(nil)
	err := pprof.StartCPUProfile(buf)
	if err != nil {
		panic(err)
	}

	start := time.Now().UnixNano()

	for i := range pfxs {
		pfxs[i].Dedup()
	}

	end := time.Now().UnixNano()

	d := end - start
	pprof.StopCPUProfile()
	fmt.Printf("Looking up Prefixes took %d ms\n", d/1000000)

	ioutil.WriteFile("profile.pprof", buf.Bytes(), 0644)

	x := bytes.NewBuffer(nil)
	pprof.WriteHeapProfile(x)

	ioutil.WriteFile("heap.pprof", x.Bytes(), 0644)
}
