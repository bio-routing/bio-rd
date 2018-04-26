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

// Package annotation annotates flows with meta data from external sources
package annotation

import (
	"context"
	"sync/atomic"

	"github.com/taktv6/tflow2/annotation/bird"
	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/netflow"
	"github.com/taktv6/tflow2/stats"

	"github.com/golang/glog"
	"google.golang.org/grpc"
)

// Annotator represents an flow annotator
type Annotator struct {
	inputs        []chan *netflow.Flow
	output        chan *netflow.Flow
	numWorkers    int
	bgpAugment    bool
	birdAnnotator *bird.Annotator
	debug         int
	cfg           *config.Config
}

// New creates a new `Annotator` instance
func New(inputs []chan *netflow.Flow, output chan *netflow.Flow, numWorkers int, cfg *config.Config) *Annotator {
	a := &Annotator{
		inputs:     inputs,
		output:     output,
		numWorkers: numWorkers,
		cfg:        cfg,
	}
	if *cfg.BGPAugmentation.Enabled {
		a.birdAnnotator = bird.NewAnnotator(*cfg.BGPAugmentation.BIRDSocket, *cfg.BGPAugmentation.BIRD6Socket, *cfg.Debug)
	}
	a.Init()
	return a
}

// Init get's the annotation layer started, receives flows, annotates them, and carries them
// further to the database module
func (a *Annotator) Init() {
	for _, ch := range a.inputs {
		for i := 0; i < a.numWorkers; i++ {
			go func(ch chan *netflow.Flow) {
				clients := make([]netflow.AnnotatorClient, 0)
				for _, an := range a.cfg.Annotators {
					var opts []grpc.DialOption
					opts = append(opts, grpc.WithInsecure())
					glog.Infof("Connecting to annotator %s at %s", an.Name, an.Target)
					conn, err := grpc.Dial(an.Target, opts...)
					if err != nil {
						glog.Errorf("Failed to dial: %v", err)
					}

					clients = append(clients, netflow.NewAnnotatorClient(conn))
				}

				for {
					// Read flow from netflow/IPFIX module
					fl := <-ch

					// Align timestamp on `aggrTime` raster
					fl.Timestamp = fl.Timestamp - (fl.Timestamp % *a.cfg.AggregationPeriod)

					// Update global statstics
					atomic.AddUint64(&stats.GlobalStats.FlowBytes, fl.Size)
					atomic.AddUint64(&stats.GlobalStats.FlowPackets, uint64(fl.Packets))

					// Send flow to external annotators
					for _, c := range clients {
						tmpFlow, err := c.Annotate(context.Background(), fl)
						if err != nil {
							glog.Errorf("Unable to annotate")
							continue
						}
						fl = tmpFlow
					}

					// Annotate flows with ASN and Prefix information from local BIRD (bird.nic.cz) instance
					if a.bgpAugment {
						a.birdAnnotator.Augment(fl)
					}

					// Send flow over to database module
					a.output <- fl
				}
			}(ch)
		}
	}
}
