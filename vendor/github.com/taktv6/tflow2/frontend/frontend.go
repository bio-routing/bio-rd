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

// Package frontend provides services via HTTP
package frontend

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	_ "net/http/pprof" // Needed for profiling only
	"strings"

	"github.com/golang/glog"
	"github.com/taktv6/tflow2/config"
	"github.com/taktv6/tflow2/database"
	"github.com/taktv6/tflow2/iana"
	"github.com/taktv6/tflow2/intfmapper"
	"github.com/taktv6/tflow2/stats"
)

// Frontend represents the web interface
type Frontend struct {
	protocols  map[string]string
	indexHTML  string
	flowDB     *database.FlowDatabase
	intfMapper *intfmapper.Mapper
	iana       *iana.IANA
	config     *config.Config
}

// New creates a new `Frontend`
func New(fdb *database.FlowDatabase, intfMapper *intfmapper.Mapper, iana *iana.IANA, config *config.Config) *Frontend {
	fe := &Frontend{
		flowDB:     fdb,
		intfMapper: intfMapper,
		iana:       iana,
		config:     config,
	}
	fe.populateIndexHTML()
	http.HandleFunc("/", fe.httpHandler)
	go http.ListenAndServe(*fe.config.Frontend.Listen, nil)
	return fe
}

// populateIndexHTML copies tflow2.html into indexHTML variable
func (fe *Frontend) populateIndexHTML() {
	html, err := ioutil.ReadFile("tflow2.html")
	if err != nil {
		glog.Errorf("Unable to read tflow2.html: %v", err)
		return
	}

	fe.indexHTML = string(html)
}

func (fe *Frontend) agentsHandler(w http.ResponseWriter, r *http.Request) {
	type routerJSON struct {
		Name       string
		Interfaces []string
	}
	type routersJSON struct {
		Agents []routerJSON
	}

	data := routersJSON{
		Agents: make([]routerJSON, 0),
	}

	for _, agent := range fe.config.Agents {
		a := routerJSON{
			Name:       *agent.Name,
			Interfaces: make([]string, 0),
		}

		intfmap := fe.intfMapper.GetInterfaceIDByName(a.Name)
		for name := range intfmap {
			a.Interfaces = append(a.Interfaces, name)
		}

		data.Agents = append(data.Agents, a)
	}

	b, err := json.Marshal(data)
	if err != nil {
		http.Error(w, fmt.Sprintf("Marshal failed: %v", err), 500)
	}

	fmt.Fprintf(w, "%s", string(b))
}

func (fe *Frontend) httpHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	parts := strings.Split(r.URL.Path, "?")
	path := parts[0]
	switch path {
	case "/":
		fe.indexHandler(w, r)
	case "/query":
		fe.queryHandler(w, r)
	case "/metrics":
		stats.Metrics(w)
	case "/protocols":
		fe.getProtocols(w, r)
	case "/promquery":
		fe.prometheusHandler(w, r)
	case "/agents":
		fe.agentsHandler(w, r)
	case "/tflow2.css":
		fileHandler(w, r, "tflow2.css")
	case "/tflow2.js":
		fileHandler(w, r, "tflow2.js")
	case "/papaparse.min.js":
		fileHandler(w, r, "vendors/papaparse/papaparse.min.js")
	}
}

func (fe *Frontend) getProtocols(w http.ResponseWriter, r *http.Request) {
	output, err := json.Marshal(fe.iana.GetIPProtocolsByName())
	if err != nil {
		glog.Warningf("Unable to marshal: %v", err)
		http.Error(w, "Unable to marshal data", 500)
	}
	fmt.Fprintf(w, "%s", output)
}

func fileHandler(w http.ResponseWriter, r *http.Request, filename string) {
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		glog.Warningf("Unable to read file: %v", err)
		http.Error(w, "Unable to read file", 404)
	}
	fmt.Fprintf(w, "%s", string(content))
}

func (fe *Frontend) indexHandler(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("query")
	if query == "" {
		query = "{}"
	}

	output := strings.Replace(fe.indexHTML, "VAR_QUERY", query, -1)
	fmt.Fprintf(w, output)
}

func (fe *Frontend) queryHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")

	query, errors := fe.translateQuery(r.URL.Query())
	if errors != nil {
		http.Error(w, "Unable to parse query:", 422)
		for _, err := range errors {
			fmt.Fprintln(w, err.Error())
		}
		return
	}

	result, err := fe.flowDB.RunQuery(&query)
	if err != nil {
		http.Error(w, fmt.Sprintf("Query failed: %v", err), 500)
		return
	}

	if len(result.Data) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	result.WriteCSV(w)
}
