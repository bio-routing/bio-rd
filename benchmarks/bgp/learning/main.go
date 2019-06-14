package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"runtime/pprof"
	"time"

	"net/http"
	_ "net/http/pprof"

	"github.com/bio-routing/bio-rd/config"
	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/server"
	"github.com/bio-routing/bio-rd/routingtable"
	"github.com/bio-routing/bio-rd/routingtable/filter"
	"github.com/bio-routing/bio-rd/routingtable/vrf"
	btesting "github.com/bio-routing/bio-rd/testing"
	"github.com/sirupsen/logrus"
)

/*
*
* This benchmark measure the time to learn 750k BGP prefixes
*
 */

func main() {
	go http.ListenAndServe("localhost:1337", nil)

	b := server.NewBgpServer()
	v, err := vrf.New("master", 0)
	if err != nil {
		log.Fatal(err)
	}

	iEnd := 100
	jEnd := 100
	kEnd := 75

	ch := make(chan struct{})
	fmt.Printf("Learning %d routes\n", kEnd*iEnd*jEnd)
	v.IPv4UnicastRIB().SetCountTarget(uint64(kEnd*iEnd*jEnd), ch)

	err = b.Start(&config.Global{
		Listen:   false,
		RouterID: 1000,
	})
	if err != nil {
		logrus.Fatalf("Unable to start BGP server: %v", err)
	}

	con := btesting.NewMockConnBidi(&btesting.MockAddr{
		Addr:  "169.254.200.0:1234",
		Proto: "TCP",
	}, &btesting.MockAddr{
		Addr:  "172.17.0.3:179",
		Proto: "TCP",
	})

	openMSG := []byte{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 29, // Length
		1,      // Type = Open
		4,      // Version
		0, 200, //ASN,
		0, 15, // Holdtime
		10, 20, 30, 40, // BGP Identifier
		0, // Opt Parm Len
	}
	con.WriteB(openMSG)

	keepAlive := []byte{
		255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
		0, 19, // Length
		4, // Type = Keepalive
	}
	con.WriteB(keepAlive)

	c := 0
	for i := 0; i < iEnd; i++ {
		for j := 0; j < jEnd; j++ {
			for k := 1; k <= kEnd; k++ {
				update := []byte{
					255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, 255, // Marker
					0, 80, // Length
					2, // Type = Update

					0, 0, // Withdraw length
					0, 53, // Total Path Attribute Length

					255,  // Attribute flags
					1,    // Attribute Type code (ORIGIN)
					0, 1, // Length
					2, // INCOMPLETE

					0,      // Attribute flags
					2,      // Attribute Type code (AS Path)
					12,     // Length
					2,      // Type = AS_SEQUENCE
					2,      // Path Segment Length
					59, 65, // AS15169
					12, 248, // AS3320
					1,      // Type = AS_SET
					2,      // Path Segment Length
					59, 65, // AS15169
					12, 248, // AS3320

					0,              // Attribute flags
					3,              // Attribute Type code (Next Hop)
					4,              // Length
					10, 11, 12, 13, // Next Hop

					0,          // Attribute flags
					4,          // Attribute Type code (MED)
					4,          // Length
					0, 0, 1, 0, // MED 256

					0,          // Attribute flags
					5,          // Attribute Type code (Local Pref)
					4,          // Length
					0, 0, 1, 0, // Local Pref 256

					0, // Attribute flags
					6, // Attribute Type code (Atomic Aggregate)
					0, // Length

					0,    // Attribute flags
					7,    // Attribute Type code (Atomic Aggregate)
					6,    // Length
					1, 2, // ASN
					10, 11, 12, 13, // Address

					24, uint8(k), uint8(i), uint8(j), // Prefix
				}
				con.WriteB(update)
				c++
			}
		}
	}

	fmt.Printf("Added routes: %d\n", c)

	buf := bytes.NewBuffer(nil)
	err = pprof.StartCPUProfile(buf)
	if err != nil {
		panic(err)
	}

	peerCfg := config.Peer{
		AdminEnabled:      true,
		LocalAS:           65200,
		PeerAS:            200,
		PeerAddress:       bnet.IPv4FromOctets(172, 17, 0, 3),
		LocalAddress:      bnet.IPv4FromOctets(169, 254, 200, 0),
		ReconnectInterval: time.Second * 15,
		HoldTime:          time.Second * 90,
		KeepAlive:         time.Second * 30,
		Passive:           true,
		RouterID:          b.RouterID(),
		IPv4: &config.AddressFamilyConfig{
			ImportFilter: filter.NewAcceptAllFilter(),
			ExportFilter: filter.NewAcceptAllFilter(),
			AddPathSend: routingtable.ClientOptions{
				MaxPaths: 10,
			},
		},
		RouteServerClient: true,
		VRF:               v,
	}

	b.AddPeer(peerCfg)

	start := time.Now().UnixNano()
	b.ConnectMockPeer(peerCfg, con)

	<-ch
	end := time.Now().UnixNano()

	d := end - start
	pprof.StopCPUProfile()
	fmt.Printf("Learning routes took %d ms\n", d/1000000)

	ioutil.WriteFile("profile.pprof", buf.Bytes(), 0644)
}
