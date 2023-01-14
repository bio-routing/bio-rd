package route

import (
	"fmt"
	"strings"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/route/api"
)

// StaticPath represents a static path of a route
type StaticPath struct {
	NextHop *bnet.IP
}

func (r *Route) staticPathSelection() {
	if len(r.paths) == 0 {
		return
	}

	r.ecmpPaths = uint(len(r.paths))
}

// Select returns negative if s < t, 0 if paths are equal, positive if s > t
func (s *StaticPath) Select(t *StaticPath) int8 {
	return s.NextHop.Compare(t.NextHop)
}

// Compare checks if paths a and t are the same
func (s *StaticPath) Compare(t *StaticPath) bool {
	return s.Equal(t)
}

// Equal returns true if s and t are euqal
func (s *StaticPath) Equal(t *StaticPath) bool {
	if s == nil || t == nil {
		return false
	}

	return s.NextHop.Compare(t.NextHop) == 0
}

// ECMP determines if path s and t are equal in terms of ECMP
func (s *StaticPath) ECMP(t *StaticPath) bool {
	return true
}

func (s *StaticPath) Copy() *StaticPath {
	if s == nil {
		return nil
	}

	cp := *s
	return &cp
}

// Print all known information about a route in logfile friendly format
func (s *StaticPath) String() string {
	return fmt.Sprintf("Next hop: %s, ", s.NextHop)
}

// Print all known information about a route in human readable form
func (s *StaticPath) Print() string {
	buf := &strings.Builder{}

	fmt.Fprintf(buf, "\t\tNext hop: %s\n", s.NextHop)

	return buf.String()
}

// ToProto converts StaticPath to proto static path
func (s *StaticPath) ToProto() *api.StaticPath {
	if s == nil {
		return nil
	}

	return &api.StaticPath{
		NextHop: s.NextHop.ToProto(),
	}
}

// StaticPathFromProtoStaticPath converts a proto StaticPath to StaticPath
func StaticPathFromProtoStaticPath(pb *api.StaticPath, dedup bool) *StaticPath {
	return &StaticPath{
		NextHop: bnet.IPFromProtoIP(pb.NextHop).Ptr(),
	}
}

func (s *StaticPath) GetNextHop() *bnet.IP {
	if s == nil {
		return nil
	}

	return s.NextHop
}
