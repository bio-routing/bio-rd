package lg

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/bio-routing/bio-rd/routingtable/vrf"
	"github.com/bio-routing/bio-rd/util/log"
	"google.golang.org/grpc"

	risAPI "github.com/bio-routing/bio-rd/cmd/ris/api"
	bnet "github.com/bio-routing/bio-rd/net"
)

var (
	//go:embed resources
	Res   embed.FS
	pages = map[string]string{
		"/":       "resources/index.gohtml",
		"/routes": "resources/routes.gohtml",
	}
)

type LookingGlass struct {
	risClient risAPI.RoutingInformationServiceClient
}

func New(risAddr string) (*LookingGlass, error) {
	cc, err := grpc.Dial(risAddr, grpc.WithInsecure())
	if err != nil {
		return nil, fmt.Errorf("grpc.Dial failed: %v", err)
	}

	lg := &LookingGlass{
		risClient: risAPI.NewRoutingInformationServiceClient(cc),
	}

	return lg, nil
}

func (l *LookingGlass) Index(w http.ResponseWriter, r *http.Request) {
	page, ok := pages[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	tpl, err := template.ParseFS(Res, page)
	if err != nil {
		log.Errorf("unable to parse %s: %v", r.RequestURI, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html")

	routers, err := l.getRouters()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Errorf("unable to get routers: %v", err)
		return
	}

	w.WriteHeader(http.StatusOK)
	data := map[string]interface{}{
		"routers": routers,
	}
	if err := tpl.Execute(w, data); err != nil {
		log.Errorf("template execution failed: %v", err)
		return
	}
}

func getVRF(r *http.Request) string {
	vrf := r.URL.Query().Get("vrf")
	if vrf != "" {
		return vrf
	}

	return "0:0"
}

func getOriginASN(r *http.Request) (uint32, error) {
	originASNStr := r.URL.Query().Get("origin_asn")
	if originASNStr == "" {
		return 0, nil
	}

	asn, err := strconv.Atoi(originASNStr)
	if err != nil {
		return 0, err
	}

	return uint32(asn), nil
}

func getAFI(r *http.Request) (int, error) {
	afiStr := r.URL.Query().Get("afi")
	if afiStr == "" {
		return 4, nil
	}

	afi, err := strconv.Atoi(afiStr)
	if err != nil {
		return 0, err
	}

	return afi, nil
}

func getFunction(r *http.Request) string {
	fn := r.URL.Query().Get("function")
	if fn != "" {
		return fn
	}

	return "all"
}

func getPrefix(r *http.Request) *bnet.Prefix {
	pfxStr := r.URL.Query().Get("prefix")

	if pfxStr == "" {
		return nil
	}

	if strings.Contains(pfxStr, ".") && !strings.Contains(pfxStr, "/") {
		pfxStr += "/32"
	}

	if strings.Contains(pfxStr, ":") && !strings.Contains(pfxStr, "/") {
		pfxStr += "/128"
	}

	pfx, err := bnet.PrefixFromString(pfxStr)
	if err != nil {
		return nil
	}

	return pfx
}

func (l *LookingGlass) Routes(w http.ResponseWriter, r *http.Request) {
	page, ok := pages[r.URL.Path]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	tpl, err := template.ParseFS(Res, page)
	if err != nil {
		writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to parse %s: %v", r.RequestURI, err))
		return
	}
	w.Header().Set("Content-Type", "text/html")

	router := r.URL.Query().Get("router")
	if router == "" {
		writeAndLogError(w, http.StatusBadRequest, "router not specified")
		return
	}
	vrf := getVRF(r)

	originASN, err := getOriginASN(r)
	if err != nil {
		writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get origin_asn: %v", err))
		return
	}

	afi, err := getAFI(r)
	if err != nil {
		writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get afi: %v", err))
		return
	}

	pfx := getPrefix(r)
	var routes []Route
	switch getFunction(r) {
	case "dump":
		routes, err = l.dumpRIB(router, vrf, afi, originASN)
		if err != nil {
			writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get routes: %v", err))
			return
		}
	case "lpm":
		if pfx == nil {
			writeAndLogError(w, http.StatusBadRequest, "lpm not possible without prefix")
			return
		}
		routes, err = l.lpm(router, vrf, pfx)
		if err != nil {
			writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get routes: %v", err))
			return
		}
	case "get":
		if pfx == nil {
			writeAndLogError(w, http.StatusBadRequest, "get not possible without prefix")
			return
		}
		routes, err = l.getRoute(router, vrf, pfx)
		if err != nil {
			writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get routes: %v", err))
			return
		}
	case "getLonger":
		if pfx == nil {
			writeAndLogError(w, http.StatusBadRequest, "get longer not possible without prefix")
			return
		}
		routes, err = l.getRouteOrLonger(router, vrf, pfx)
		if err != nil {
			writeAndLogError(w, http.StatusInternalServerError, fmt.Sprintf("unable to get routes: %v", err))
			return
		}
	}

	w.WriteHeader(http.StatusOK)
	data := map[string]interface{}{
		"routes": routes,
	}
	if err := tpl.Execute(w, data); err != nil {
		log.Errorf("template execution failed: %v", err)
		return
	}
}

func writeAndLogError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Write([]byte(msg))
	log.Error(msg)
}

type Router struct {
	Address string
	SysName string
	VRFs    []string
}

func (l *LookingGlass) getRouters() ([]Router, error) {
	getRouterResp, err := l.risClient.GetRouters(context.Background(), &risAPI.GetRoutersRequest{})
	if err != nil {
		return nil, fmt.Errorf("unable to get routers: %v", err)
	}

	ret := make([]Router, 0, len(getRouterResp.Routers))
	for _, r := range getRouterResp.Routers {
		ret = append(ret, Router{
			Address: r.Address,
			SysName: r.SysName,
			VRFs:    humanReadableVRFs(r.VrfIds),
		})

	}

	return ret, nil
}

func humanReadableVRFs(vrfs []uint64) []string {
	ret := make([]string, 0, len(vrfs))
	for _, v := range vrfs {
		ret = append(ret, vrf.RouteDistinguisherHumanReadable(v))
	}

	return ret
}

func afiToProtoAFI(afi int) risAPI.DumpRIBRequest_AFISAFI {
	if afi == 4 {
		return risAPI.DumpRIBRequest_IPv4Unicast
	} else if afi == 6 {
		return risAPI.DumpRIBRequest_IPv6Unicast
	}

	return 0
}

func (l *LookingGlass) dumpRIB(router string, vrf string, afi int, originASN uint32) ([]Route, error) {
	streamClient, err := l.risClient.DumpRIB(context.Background(), &risAPI.DumpRIBRequest{
		Router:  router,
		Vrf:     vrf,
		Afisafi: afiToProtoAFI(afi),
		Filter: &risAPI.RIBFilter{
			OriginatingAsn: originASN,
		},
	})

	if err != nil {
		return nil, fmt.Errorf("rpc failed: %v", err)
	}

	res := make([]Route, 0)
	for {
		repl, err := streamClient.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		res = append(res, convertRoute(repl.Route))
	}

	return res, nil
}

func (l *LookingGlass) lpm(router string, vrf string, pfx *bnet.Prefix) ([]Route, error) {
	resp, err := l.risClient.LPM(context.Background(), &risAPI.LPMRequest{
		Router: router,
		Vrf:    vrf,
		Pfx:    pfx.ToProto(),
	})
	if err != nil {
		return nil, fmt.Errorf("rpc failed: %v", err)
	}

	return convertRoutes(resp.Routes), nil
}

func (l *LookingGlass) getRoute(router string, vrf string, pfx *bnet.Prefix) ([]Route, error) {
	resp, err := l.risClient.Get(context.Background(), &risAPI.GetRequest{
		Router: router,
		Vrf:    vrf,
		Pfx:    pfx.ToProto(),
	})
	if err != nil {
		return nil, fmt.Errorf("rpc failed: %v", err)
	}

	return convertRoutes(resp.Routes), nil
}

func (l *LookingGlass) getRouteOrLonger(router string, vrf string, pfx *bnet.Prefix) ([]Route, error) {
	resp, err := l.risClient.GetLonger(context.Background(), &risAPI.GetLongerRequest{
		Router: router,
		Vrf:    vrf,
		Pfx:    pfx.ToProto(),
	})
	if err != nil {
		return nil, fmt.Errorf("rpc failed: %v", err)
	}

	return convertRoutes(resp.Routes), nil
}
