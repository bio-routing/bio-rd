// Copyright 2017 Google Inc. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package bird can lookup IP prefixes and autonomous system numbers and
// add them to flows in case the routers implementation doesn't support this, e.g. ipt-NETFLOW
package bird

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/stats"
)

// QueryResult carries all useful information we extracted from a BIRD querys result
type QueryResult struct {
	// Pfx is the prefix that is being used to forward packets for the IP
	// address from the query
	Pfx net.IPNet

	// As is the ASN that the subject IP is announced by
	AS uint32

	// NhAs is the ASN of the subject IPs associated Next Hop
	NHAS uint32
}

// QueryCache represents a set of QueryResults that have been cached
type QueryCache struct {
	cache map[string]QueryResult
	lock  sync.RWMutex
}

// Query represents a query to BIRD and encapsulates it with a channel where it's result is expected
type Query struct {
	birdQuery string
	retCh     chan *QueryResult
}

// birdCon represents a connection to a BIRD instance
type birdCon struct {
	sock  string
	con   net.Conn
	recon chan bool
	lock  sync.RWMutex
}

// Annotator represents a BIRD based BGP annotator
type Annotator struct {
	queryC chan *Query

	// cache is used to cache query results
	cache *QueryCache

	// connection to BIRD
	bird4 *birdCon

	// connectio to BIRD6
	bird6 *birdCon

	// debug level
	debug int
}

// NewAnnotator creates a new BIRD annotator and get's service started
func NewAnnotator(sock string, sock6 string, debug int) *Annotator {
	a := &Annotator{
		cache:  newQueryCache(),
		queryC: make(chan *Query),
		debug:  debug,
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.bird4 = newBirdCon(sock)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		a.bird6 = newBirdCon(sock6)
	}()

	wg.Wait()
	go a.gateway()

	return a
}

// getConn gets the net.Conn property of the BIRD connection
func (c *birdCon) getConn() *net.Conn {
	return &c.con
}

// newQueryCache creates and initializes a new `QueryCache`
func newQueryCache() *QueryCache {
	return &QueryCache{cache: make(map[string]QueryResult)}
}

// reconnector receives a signal via channel that triggers a connection attempt to BIRD
func (c *birdCon) reconnector() {
	for {
		// wait for signal of a closed connection
		<-c.recon

		// try to connect up to 5 times
		for i := 0; i < 5; i++ {
			tmpCon, err := net.Dial("unix", c.sock)
			if err != nil {
				glog.Warningf("Unable to connect to BIRD on %s: %v", c.sock, err)
				continue
			}

			// Read welcome message we are not interested in
			buf := make([]byte, 1024)
			nbytes, err := tmpCon.Read(buf[:])
			if err != nil || nbytes == 0 {
				if err == nil {
					tmpCon.Close()
				}
				glog.Warning("Reading from BIRD failed: %v", err)
				continue
			}

			c.lock.Lock()
			c.con = tmpCon
			c.lock.Unlock()
			break
		}
	}
}

// Get tries to receive an entry from QueryCache `qc`
func (qc *QueryCache) Get(addr []byte) *QueryResult {
	qc.lock.RLock()
	defer qc.lock.RUnlock()

	res, ok := qc.cache[net.IP(addr).String()]
	if !ok {
		atomic.AddUint64(&stats.GlobalStats.BirdCacheMiss, 1)
		return nil
	}
	atomic.AddUint64(&stats.GlobalStats.BirdCacheHits, 1)
	return &res
}

// Set sets data for `addr` in QueryCache `qc` to `qres`
func (qc *QueryCache) Set(addr []byte, qres *QueryResult) {
	qc.lock.Lock()
	defer qc.lock.Unlock()
	qc.cache[net.IP(addr).String()] = *qres
}

// newBirdCon creates a birdCon to socket `s`
func newBirdCon(s string) *birdCon {
	b := &birdCon{
		sock:  s,
		recon: make(chan bool),
	}
	go b.reconnector()
	b.recon <- true
	return b
}

// Augment function provides the main interface to the external world to consume service of this module
func (a *Annotator) Augment(fl *netflow.Flow) {
	srcRes := a.cache.Get(fl.SrcAddr)
	if srcRes == nil {
		srcRes = a.query(net.IP(fl.Router), fl.SrcAddr)
		a.cache.Set(fl.SrcAddr, srcRes)
	}

	dstRes := a.cache.Get(fl.DstAddr)
	if dstRes == nil {
		dstRes = a.query(net.IP(fl.Router), fl.DstAddr)
		a.cache.Set(fl.DstAddr, dstRes)
	}

	fl.SrcPfx = &netflow.Pfx{
		IP:   srcRes.Pfx.IP,
		Mask: srcRes.Pfx.Mask,
	}

	fl.DstPfx = &netflow.Pfx{
		IP:   dstRes.Pfx.IP,
		Mask: dstRes.Pfx.Mask,
	}

	fl.SrcAs = srcRes.AS
	fl.DstAs = dstRes.AS
	fl.NextHopAs = dstRes.NHAS
}

// query forms a query, sends it to the processing engine, reads the result and returns it
func (a *Annotator) query(rtr net.IP, addr net.IP) *QueryResult {
	q := Query{
		birdQuery: fmt.Sprintf("show route all for %s protocol nf_%s\n", addr.String(), strings.Replace(rtr.String(), ".", "_", -1)),
		retCh:     make(chan *QueryResult),
	}
	a.queryC <- &q
	return <-q.retCh
}

// gateway starts the main service routine
func (a *Annotator) gateway() {
	buf := make([]byte, 1024)
	for {
		var res QueryResult
		query := <-a.queryC
		if query == nil {
			continue
		}
		data := []byte(query.birdQuery)

		// Determine if we are being queried for an IPv4 or an IPv6 address
		bird := a.bird4
		if strings.Contains(query.birdQuery, ":") {
			bird = a.bird6
		}

		// Skip annotation if we're not connected to bird yet
		bird.lock.RLock()
		if bird.con == nil {
			glog.Warningf("skipped annotating flow: BIRD is not connected yet")
			bird.lock.RUnlock()
			query.retCh <- &res
			continue
		}

		// Send query to BIRD
		_, err := bird.con.Write(data)
		if err != nil {
			bird.lock.RUnlock()
			glog.Errorf("Unable to write to BIRD: %v", err)
			bird.recon <- true
			continue
		}
		bird.lock.RUnlock()

		// Read reply from BIRD
		n, err := bird.con.Read(buf[:])
		if err != nil {
			bird.lock.RUnlock()
			glog.Errorf("unable to read from BIRD: %v", err)
			bird.recon <- true
			continue
		}

		// Parse BIRDs output
		output := string(buf[:n])
		lines := strings.Split(output, "\n")
		for i, line := range lines {
			// Take the first line as that should contain the prefix
			if i == 0 {
				parts := strings.Split(line, " ")
				if len(parts) == 0 {
					glog.Warningf("unexpected empty output for query '%v'", query)
					continue
				}
				pfx := parts[0]
				parts = strings.Split(pfx, "-")
				if len(parts) != 2 {
					glog.Warningf("unexpected split results for query '%v'", query)
					continue
				}
				pfx = parts[1]

				_, tmpNet, err := net.ParseCIDR(pfx)
				res.Pfx = *tmpNet
				if err != nil {
					glog.Warningf("unable to parse CIDR from BIRD: %v (query '%v')", err, query)
					continue
				}
				continue
			}

			// Find line that contains the AS Path
			if strings.Contains(line, "BGP.as_path: ") {
				// Remove curly braces from BIRD AS path (ignores aggregators), e.g. BGP.as_path: 25291 3320 20940 { 16625 }
				line = strings.Replace(line, "{ ", "", -1)
				line = strings.Replace(line, " }", "", -1)

				parts := strings.Split(line, "BGP.as_path: ")
				pathParts := strings.Split(parts[1], " ")

				if len(parts) < 2 || parts[1] == "" {
					break
				}

				AS, err := strconv.ParseUint(pathParts[len(pathParts)-1], 10, 32)
				if err != nil {
					glog.Warningf("unable to parse ASN")
				}

				NHAS, err := strconv.ParseUint(pathParts[0], 10, 32)
				if err != nil {
					glog.Warningf("unable to parse next hop ASN")
				}

				res.AS = uint32(AS)
				res.NHAS = uint32(NHAS)
				break
			}
		}
		if res.AS == 0 && a.debug > 2 {
			glog.Warningf("unable to find AS path for '%v'", query)
		}
		query.retCh <- &res
	}
}
