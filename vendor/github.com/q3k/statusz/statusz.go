// Copyright 2018 Serge Bazanski
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Based off github.com/youtube/doorman/blob/master/go/status/status.go

// Copyright 2016 Google, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package statusz

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"html"
	"html/template"
	"io"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"sync"
	"time"

	"github.com/golang/glog"
	"github.com/shirou/gopsutil/load"
)

var (
	binaryName  = filepath.Base(os.Args[0])
	binaryHash  string
	hostname    string
	username    string
	serverStart = time.Now()

	lock     sync.Mutex
	sections []section
	tmpl     = template.Must(reparse(nil))
	funcs    = make(template.FuncMap)

	DefaultMux = true
)

type section struct {
	Banner   string
	Fragment string
	F        func() interface{}
}

var statusHTML = `<!DOCTYPE html>
<html>
<head>
<title>Status for {{.BinaryName}}</title>
<style>
body {
font-family: sans-serif;
background: #fff;
}
h1 {
clear: both;
width: 100%;
text-align: center;
font-size: 120%;
background: #eeeeff;
}
.lefthand {
float: left;
width: 80%;
}
.righthand {
text-align: right;
}
</style>
</head>
<h1>Status for {{.BinaryName}}</h1>
<div>
<div class=lefthand>
Started at {{.StartTime}}<br>
Current time {{.CurrentTime}}<br>
SHA256 {{.BinaryHash}}<br>
</div>
<div class=righthand>
Running as {{.Username}} on {{.Hostname}}<br>
Load {{.LoadAvg}}<br>
View <a href=/debug/status>status</a>,
	<a href=/debug/requests>requests</a>
</div>
</div>`

func reparse(sections []section) (*template.Template, error) {
	var buf bytes.Buffer

	io.WriteString(&buf, `{{define "status"}}`)
	io.WriteString(&buf, statusHTML)

	for i, sec := range sections {
		fmt.Fprintf(&buf, "<h1>%s</h1>\n", html.EscapeString(sec.Banner))
		fmt.Fprintf(&buf, "{{$sec := index .Sections %d}}\n", i)
		fmt.Fprintf(&buf, `{{template "sec-%d" call $sec.F}}`+"\n", i)
	}
	fmt.Fprintf(&buf, `</html>`)
	io.WriteString(&buf, "{{end}}\n")

	for i, sec := range sections {
		fmt.Fprintf(&buf, `{{define "sec-%d"}}%s{{end}}\n`, i, sec.Fragment)
	}
	return template.New("").Funcs(funcs).Parse(buf.String())
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	lock.Lock()
	defer lock.Unlock()

	loadavg := "unknown"
	l, err := load.AvgWithContext(r.Context())
	if err == nil {
		loadavg = fmt.Sprintf("%.2f %.2f %.2f", l.Load1, l.Load5, l.Load15)
	}

	data := struct {
		Sections    []section
		BinaryName  string
		BinaryHash  string
		Hostname    string
		Username    string
		StartTime   string
		CurrentTime string
		LoadAvg     string
	}{
		Sections:    sections,
		BinaryName:  binaryName,
		BinaryHash:  binaryHash,
		Hostname:    hostname,
		Username:    username,
		StartTime:   serverStart.Format(time.RFC1123),
		CurrentTime: time.Now().Format(time.RFC1123),
		LoadAvg:     loadavg,
	}

	if err := tmpl.ExecuteTemplate(w, "status", data); err != nil {
		glog.Errorf("servenv: couldn't execute template: %v", err)
	}
}

func init() {
	var err error
	hostname, err = os.Hostname()
	if err != nil {
		glog.Fatalf("os.Hostname: %v", err)
	}

	user, err := user.Current()
	if err != nil {
		glog.Fatalf("user.Current: %v", err)
	}
	username = fmt.Sprintf("%s (%s)", user.Username, user.Uid)

	f, err := os.Open(os.Args[0])
	if err != nil {
		glog.Fatalf("os.Open(%q): %v", os.Args[0], err)
	}
	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		glog.Fatalf("io.Copy: %v", err)
	}
	binaryHash = fmt.Sprintf("%x", h.Sum(nil))

	if DefaultMux {
		http.HandleFunc("/debug/status", StatusHandler)
	}
}

// AddStatusPart adds a new section to status. frag is used as a
// subtemplate of the template used to render /debug/status, and will
// be executed using the value of invoking f at the time of the
// /debug/status request. frag is parsed and executed with the
// html/template package. Functions registered with AddStatusFuncs
// may be used in the template.
func AddStatusPart(banner, frag string, f func(context.Context) interface{}) {
	lock.Lock()
	defer lock.Unlock()

	secs := append(sections, section{
		Banner:   banner,
		Fragment: frag,
		F:        func() interface{} { return f(context.Background()) },
	})

	var err error
	tmpl, err = reparse(secs)
	if err != nil {
		secs[len(secs)-1] = section{
			Banner:   banner,
			Fragment: "<code>bad status template: {{.}}</code>",
			F:        func() interface{} { return err },
		}
	}
	tmpl, _ = reparse(secs)
	sections = secs
}

// AddStatusSection registers a function that generates extra
// information for /debug/status. If banner is not empty, it will be
// used as a header before the information. If more complex output
// than a simple string is required use AddStatusPart instead.
func AddStatusSection(banner string, f func(context.Context) string) {
	AddStatusPart(banner, `{{.}}`, func(ctx context.Context) interface{} { return f(ctx) })
}
