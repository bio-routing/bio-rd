package route

import (
	"fmt"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route/api"
)

// BGPPath represents a set of BGP path attributes
type GRPPath struct {
	NextHop  *bnet.IP
	Device   string
	MetaData map[string]string
}

func (g *GRPPath) AddMetaData(metaData map[string]string) {
	if metaData == nil {
		return
	}

	if g.MetaData == nil {
		g.MetaData = make(map[string]string)
	}

	for k, v := range metaData {
		g.MetaData[k] = v
	}
}

func (g *GRPPath) ToProto() *api.GRPPath {
	if g == nil {
		return nil
	}

	return &api.GRPPath{
		NextHop:  g.ToProto().NextHop,
		Device:   g.Device,
		MetaData: g.MetaData,
	}
}

func (g *GRPPath) Compare(c *GRPPath) bool {
	if g.NextHop.Compare(c.NextHop) != 0 {
		return false
	}

	if g.Device != c.Device {
		return false
	}

	// MetaData isn't relavant for routing

	return true
}

func (g *GRPPath) Equal(c *GRPPath) bool {
	return g.Select(c) == 0
}

func (g *GRPPath) Select(c *GRPPath) int8 {
	if g.NextHop.Compare(c.NextHop) == -1 {
		return 1
	}

	if g.NextHop.Compare(c.NextHop) == 1 {
		return -1
	}

	return 0
}

func (g *GRPPath) String() string {
	buf := &strings.Builder{}

	fmt.Fprintf(buf, "NEXT HOP: %s, ", g.NextHop)
	fmt.Fprintf(buf, ", Device %s", g.Device)
	fmt.Fprintf(buf, ", MetaData %s", g.metaDataToString())

	return buf.String()
}

func (g *GRPPath) metaDataToString() string {
	items := make([]string, 0)

	for k, v := range g.MetaData {
		items = append(items, fmt.Sprintf("%s: %s", k, v))
	}

	return strings.Join(items, ", ")
}

func (g *GRPPath) Print() string {
	buf := &strings.Builder{}

	fmt.Fprintf(buf, "\t\tNEXT HOP: %s\n", g.NextHop)
	fmt.Fprintf(buf, "\t\tDevice: %s", g.Device)
	fmt.Fprintf(buf, "\t\tMetaData: %s", g.metaDataToString())

	return buf.String()
}

func (g *GRPPath) Copy() *GRPPath {
	if g == nil {
		return nil
	}

	cp := &GRPPath{
		NextHop:  g.NextHop.Dedup(),
		Device:   g.Device,
		MetaData: make(map[string]string),
	}

	for k, v := range g.MetaData {
		cp.MetaData[k] = v
	}

	return cp
}

func (g *GRPPath) GetNextHop() *bnet.IP {
	if g == nil {
		return nil
	}

	return g.NextHop
}

func (p *Path) redistributeToGRP() (*Path, error) {
	oldType := p.Type

	// TODO: Get device from StaticPath once added there

	p.GRPPath = &GRPPath{
		NextHop:  p.GetNextHop(),
		MetaData: make(map[string]string),
	}
	p.Type = GRPPathType

	p.PurgePathInformation(oldType)

	return p, nil
}
