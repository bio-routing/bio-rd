package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime/pprof"
	"time"

	"github.com/bio-routing/bio-rd/protocols/bgp/packet"

	log "github.com/sirupsen/logrus"
)

var (
	nRuns = flag.Int("runs", 1, "# runs")
)

type task struct {
	num int
	raw *bytes.Buffer
	msg *packet.BGPMessage
}

func main() {
	flag.Parse()

	updates := make([][]*bytes.Buffer, *nRuns)
	for i := 0; i < *nRuns; i++ {
		updates[i] = make([]*bytes.Buffer, 0)
	}

	raw, err := ioutil.ReadFile("AS8881.raw")
	if err != nil {
		log.Errorf("Unable to open PCAP file: %v", err)
		os.Exit(1)
	}

	msgs := extractBGPMessages(raw)
	for _, msg := range msgs {
		for i := 0; i < *nRuns; i++ {
			updates[i] = append(updates[i], bytes.NewBuffer(msg))
		}
	}

	c := len(updates[0])

	fmt.Printf("Decoding %d BGP messages\n", c)

	buf := bytes.NewBuffer(nil)
	err = pprof.StartCPUProfile(buf)
	if err != nil {
		panic(err)
	}

	dco := &packet.DecodeOptions{
		Use32BitASN: true,
	}

	start := time.Now().UnixNano()

	nlriCount := 0
	for j := 0; j < *nRuns; j++ {
		for i := 0; i < c; i++ {
			msg, err := packet.Decode(updates[j][i], dco)
			if err != nil {
				fmt.Printf("Unable to decode msg %d: %v\n", i, err)
				continue
			}

			if msg.Header.Type == 2 {
				n := msg.Body.(*packet.BGPUpdate).NLRI
				for {
					if n == nil {
						break
					}

					nlriCount++

					n = n.Next
				}
			}
		}
	}
	fmt.Printf("NLRIs: %d\n", nlriCount)

	end := time.Now().UnixNano()

	d := end - start
	pprof.StopCPUProfile()
	fmt.Printf("decoding updates took %d ms\n", d/1000000)

	ioutil.WriteFile("profile.pprof", buf.Bytes(), 0644)

	x := bytes.NewBuffer(nil)
	pprof.WriteHeapProfile(x)

	ioutil.WriteFile("heap.pprof", x.Bytes(), 0644)
}

func hexDump(input []byte) string {
	s := ""
	for _, x := range input {
		s += fmt.Sprintf("%x ", x)
	}

	return s
}

func extractBGPMessages(input []byte) [][]byte {
	fmt.Printf("Extracting BGP messages from %d bytes\n", len(input))
	ret := make([][]byte, 0)

	//fmt.Printf("Data: %v\n", input[0:24])
	l := len(input)
	i := 0
	for {
		if i+17 > l {
			break
		}

		for j := 0; j < 16; j++ {
			if input[i+j] != 255 {
				panic(fmt.Sprintf("Invalid BGP marker: (%d+%d=%d): %s", i, j, i+j, hexDump(input[i:i+16])))
			}
		}

		msgLen := uint16(input[i+16])*256 + uint16(input[i+17])

		ret = append(ret, input[i:i+int(msgLen)])

		if msgLen == 0 {
			panic(msgLen)
		}

		i += int(msgLen)
	}

	fmt.Printf("Done\n")

	return ret
}
