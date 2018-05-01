// Copyright 2017 Google Inc, EXARING AG. All Rights Reserved.
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//     http://www.apache.org/licenses/LICENSE-2.0
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package main is the main package of tflow2
package main

import (
	"flag"
	"runtime"
	"sync"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/annotation"
	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/database"
	"github.com/taktv6/tflow2/frontend"
	"github.com/taktv6/tflow2/iana"
	"github.com/taktv6/tflow2/ifserver"
	"github.com/taktv6/tflow2/intfmapper"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/nfserver"
	"github.com/taktv6/tflow2/sfserver"
	"github.com/taktv6/tflow2/srcache"
	"github.com/taktv6/tflow2/stats"
)

var (
	protoNums     = flag.String("protonums", "protocol_numbers.csv", "CSV file to read protocol definitions from")
	sockReaders   = flag.Int("sockreaders", 24, "Num of go routines reading and parsing netflow packets")
	channelBuffer = flag.Int("channelbuffer", 1024, "Size of buffer for channels")
	dbAddWorkers  = flag.Int("dbaddworkers", 24, "Number of workers adding flows into database")
	nAggr         = flag.Int("numaggr", 12, "Number of flow aggregator workers")

	configFile = flag.String("config", "config.yml", "tflow2 configuration file")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())
	flag.Parse()

	cfg, err := config.New(*configFile)
	if err != nil {
		glog.Exitf("Unable to get configuration: %v", err)
	}

	// Initialize statistics module
	stats.Init()

	inftMapper, err := intfmapper.New(cfg.Agents, *cfg.AggregationPeriod)
	if err != nil {
		glog.Exitf("Unable to initialize interface mappper: %v", err)
	}

	chans := make([]chan *netflow.Flow, 0)

	// Sample Rate Cache
	srcache := srcache.New(cfg.Agents)

	// Netflow v9 Server
	if *cfg.NetflowV9.Enabled {
		nfs := nfserver.New(*sockReaders, cfg, srcache)
		chans = append(chans, nfs.Output)
	}

	// IPFIX Server
	if *cfg.IPFIX.Enabled {
		ifs := ifserver.New(*sockReaders, cfg, srcache)
		chans = append(chans, ifs.Output)
	}

	// sFlow Server
	if *cfg.Sflow.Enabled {
		sfs := sfserver.New(*sockReaders, cfg, srcache)
		chans = append(chans, sfs.Output)
	}

	// Get IANA instance
	iana := iana.New()

	// Start the database layer
	flowDB := database.New(
		*cfg.AggregationPeriod,
		*cfg.CacheTime,
		*dbAddWorkers,
		*cfg.Debug,
		*cfg.CompressionLevel,
		cfg.DataDir,
		*cfg.Anonymize,
		inftMapper,
		cfg.AgentsNameByIP,
		iana,
	)

	// Start the annotation layer
	annotation.New(
		chans,
		flowDB.Input,
		*nAggr,
		cfg,
	)

	// Frontend
	if *cfg.Frontend.Enabled {
		frontend.New(
			flowDB,
			inftMapper,
			iana,
			cfg,
		)
	}

	var wg sync.WaitGroup
	wg.Add(1)
	wg.Wait()
}
