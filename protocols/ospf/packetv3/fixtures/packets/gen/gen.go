package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"text/template"

	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3/fixtures"
	"github.com/google/gopacket/pcapgo"
)

func main() {
	cwd := ""
	var filename string
	for depth := 0; filename != "gen.go"; depth++ {
		_, currentPath, _, ok := runtime.Caller(depth)
		if !ok {
			return
		}
		filename = filepath.Base(currentPath)
		cwd = currentPath
	}
	dir := filepath.Dir(cwd) + "/../"

	files := []string{
		"OSPFv3_multipoint_adjacencies.cap",
		"OSPFv3_broadcast_adjacency.cap",
		"OSPFv3_NBMA_adjacencies.cap",
	}

	for _, path := range files {
		fmt.Printf("Processing infile %s\n", path)
		f, err := os.Open(dir + "/../" + path)
		if err != nil {
			panic(err)
		}
		defer f.Close()

		r, err := pcapgo.NewReader(f)
		if err != nil {
			panic(err)
		}

		var packetCount int
		tempBuf := bytes.NewBufferString("")
		funcs := make([]string, 0)
		for {
			data, _, err := r.ReadPacketData()
			if err == io.EOF {
				break
			}
			if err != nil {
				panic(err)
			}

			pl, src, dst, err := fixtures.Payload(data)
			if err != nil {
				panic(err)
			}

			funcName := serializePacket(tempBuf, path, pl, src, dst, packetCount)
			funcs = append(funcs, funcName)
			packetCount++
		}

		args := &GenTemplArgs{File: path, PacketFuncs: funcs}
		outBuf := bytes.NewBufferString("")
		if err := genTemplate.Execute(outBuf, args); err != nil {
			panic(err)
		}

		tempBuf.WriteTo(outBuf)
		file, err := os.Create(dir + path + ".go")
		defer file.Close()
		if err != nil {
			panic(err)
		}
		outBuf.WriteTo(file)
	}
}

var genTemplate = template.Must(template.New("Gen").Parse(`
// GENERATED FILE - do not edit!
// to regenerate this, run "go run ./protocols/ospf/packetv3/fixtures/packets/gen/"

package packets

import (
	"github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/ospf/packetv3"
)

func init() {
	filePkts := make([]*packetv3.OSPFv3Message, {{ len .PacketFuncs }})
	{{ range $index, $func := .PacketFuncs -}}
	filePkts[{{ $index }}] = {{ $func }}()
	{{ end -}}
	Packets["{{ .File }}"] = filePkts
}
`))

type GenTemplArgs struct {
	File        string
	PacketFuncs []string
}

var packetTemplate = template.Must(template.New("packet").Parse(`
func {{ .FuncName }}() *packetv3.OSPFv3Message {
	packet := {{ .Msg }}
	body := {{ .Body }}
	{{ $lsafield := .LSAField -}}
	{{ if gt (len .LSAs) 0 -}}
		body.{{ $lsafield }} = make([]*packetv3.LSA, {{ len .LSAs }})
	{{ range $index, $lsa := .LSAs -}}
		body.{{ $lsafield }}[{{ $index }}] = {{ $lsa }}
	{{ end -}}
	{{ end -}}
	packet.Body = body
	return packet
}
`))

type PacketTemplArgs struct {
	FuncName string
	Msg      string
	Body     string
	LSAField string
	LSAs     []string
}

var pointerRegex = regexp.MustCompile(`\(\*packetv3\.\w+\)\(0x[0-9a-f]+\)`)
var netRegex = regexp.MustCompile(`net.IP\{higher:(0x[0-9a-f]+), lower:(0x[0-9a-f]+), isLegacy:false\}`)

func cleanSerialized(in string) string {
	clean := pointerRegex.ReplaceAllString(in, "nil")
	clean = strings.ReplaceAll(clean, "_:0x0, ", "")
	clean = netRegex.ReplaceAllString(clean, "net.IPv6($1, $2)")
	return clean
}

func serializeLSAs(items []*packetv3.LSA) []string {
	out := make([]string, len(items))
	for i := range items {
		ser := fmt.Sprintf("%#v", items[i])
		bodyStr := "nil"
		if items[i].Body != nil {
			bodyStr = cleanSerialized(fmt.Sprintf("%#v", items[i].Body))
		}
		ser = pointerRegex.ReplaceAllString(ser, bodyStr)
		out[i] = cleanSerialized(ser)
	}
	return out
}

func serializePacket(out *bytes.Buffer, file string, payload []byte, src, dst net.IP, count int) string {
	buf := bytes.NewBuffer(payload)
	msg, _, err := packetv3.DeserializeOSPFv3Message(buf, src, dst)
	if err != nil {
		panic(err)
	}

	args := &PacketTemplArgs{}
	args.FuncName = fmt.Sprintf("packet_%s_%03d", strings.ReplaceAll(file, ".cap", ""), count+1)
	args.Msg = cleanSerialized(fmt.Sprintf("%#v", msg))
	args.Body = cleanSerialized(fmt.Sprintf("%#v", msg.Body))

	switch t := msg.Body.(type) {
	case *packetv3.DatabaseDescription:
		args.LSAField = "LSAHeaders"
		args.LSAs = serializeLSAs(t.LSAHeaders)
	case *packetv3.LinkStateUpdate:
		args.LSAField = "LSAs"
		args.LSAs = serializeLSAs(t.LSAs)
	case *packetv3.LinkStateAcknowledgement:
		args.LSAField = "LSAHeaders"
		args.LSAs = serializeLSAs(t.LSAHeaders)
	}

	if err := packetTemplate.Execute(out, args); err != nil {
		panic(err)
	}

	return args.FuncName
}
