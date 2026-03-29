package route

import (
	"testing"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/protocols/bgp/types"
)

// TestECMP_MixedPathTypes_BGPvsStatic verifies that ECMP does not panic
// when comparing a BGPPathType path against a StaticPathType path.
// This is the primary crash: bio-rd calls p.BGPPath.ECMP(q.BGPPath)
// without checking that q is also BGPPathType. When q is StaticPathType,
// q.BGPPath is nil → nil pointer dereference.
func TestECMP_MixedPathTypes_BGPvsStatic(t *testing.T) {
	bgpPath := &Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{
				NextHop:   bnet.IPv4(1).Ptr(),
				LocalPref: 100,
				Origin:    0,
				Source:    bnet.IPv4(1).Ptr(),
			},
			ASPath:    types.NewASPath([]uint32{64512}),
			ASPathLen: 1,
		},
	}

	staticPath := &Path{
		Type: StaticPathType,
		StaticPath: &StaticPath{
			NextHop: bnet.IPv4(2).Ptr(),
		},
	}

	// Must not panic — should return false (different types are never ECMP-equal)
	result := bgpPath.ECMP(staticPath)
	if result {
		t.Error("ECMP should return false for different path types")
	}
}

// TestECMP_MixedPathTypes_StaticvsBGP is the reverse direction.
func TestECMP_MixedPathTypes_StaticvsBGP(t *testing.T) {
	staticPath := &Path{
		Type: StaticPathType,
		StaticPath: &StaticPath{
			NextHop: bnet.IPv4(1).Ptr(),
		},
	}

	bgpPath := &Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{
				NextHop:   bnet.IPv4(2).Ptr(),
				LocalPref: 100,
				Origin:    0,
				Source:    bnet.IPv4(2).Ptr(),
			},
			ASPath:    types.NewASPath([]uint32{64513}),
			ASPathLen: 1,
		},
	}

	result := staticPath.ECMP(bgpPath)
	if result {
		t.Error("ECMP should return false for different path types")
	}
}

// TestECMP_MixedPathTypes_FIBvsBGP covers the FIBPathType case.
func TestECMP_MixedPathTypes_FIBvsBGP(t *testing.T) {
	fibPath := &Path{
		Type: FIBPathType,
		FIBPath: &FIBPath{
			NextHop: bnet.IPv4(1).Ptr(),
		},
	}

	bgpPath := &Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{
				NextHop:   bnet.IPv4(2).Ptr(),
				LocalPref: 100,
				Origin:    0,
				Source:    bnet.IPv4(2).Ptr(),
			},
			ASPath:    types.NewASPath([]uint32{64512}),
			ASPathLen: 1,
		},
	}

	result := fibPath.ECMP(bgpPath)
	if result {
		t.Error("ECMP should return false for different path types")
	}
}

// TestRouteUpdateEqualPathCount_MixedTypes verifies that
// Route.updateEqualPathCount does not panic when the route has paths
// of different types (the real-world scenario: locally-originated
// StaticPath + BGP-learned BGPPath for the same anycast prefix).
func TestRouteUpdateEqualPathCount_MixedTypes(t *testing.T) {
	pfx := bnet.NewPfx(bnet.IPv4FromOctets(10, 0, 99, 0), 24)

	r := NewRoute(pfx.Ptr(), &Path{
		Type: StaticPathType,
		StaticPath: &StaticPath{
			NextHop: bnet.IPv4(1).Ptr(),
		},
	})

	r.AddPath(&Path{
		Type: BGPPathType,
		BGPPath: &BGPPath{
			BGPPathA: &BGPPathA{
				NextHop:   bnet.IPv4(2).Ptr(),
				LocalPref: 100,
				Origin:    0,
				Source:    bnet.IPv4(2).Ptr(),
			},
			ASPath:    types.NewASPath([]uint32{64513}),
			ASPathLen: 1,
		},
	})

	// Trigger path selection, which in turn calls updateEqualPathCount.
	r.PathSelection()

	// Must not panic and should result in a single ECMP path (static preferred).
	count := r.ECMPPathCount()
	if count != 1 {
		t.Errorf("unexpected ECMP count: got %d, want 1", count)
	}
}

// TestBGPPathPrepend_NilASPath verifies that Prepend does not panic
// when ASPath is nil. This happens when a route is originated locally
// as BGPPathType without initializing the ASPath field.
func TestBGPPathPrepend_NilASPath(t *testing.T) {
	p := &BGPPath{
		BGPPathA: &BGPPathA{
			NextHop:   bnet.IPv4(1).Ptr(),
			LocalPref: 200,
			Origin:    0,
			Source:    bnet.IPv4(1).Ptr(),
		},
		ASPath: nil, // not initialized
	}

	// Must not panic — should handle nil ASPath gracefully
	p.Prepend(64512, 1)

	if p.ASPath == nil {
		t.Error("Prepend should initialize ASPath if nil")
	}
	if p.ASPathLen == 0 {
		t.Error("Prepend should increase ASPathLen")
	}
	if len(*p.ASPath) == 0 {
		t.Error("Prepend should add at least one AS_PATH segment")
		return
	}
	firstSeg := (*p.ASPath)[0]
	if len(firstSeg.ASNs) == 0 {
		t.Error("Prepend should add at least one ASN to the first AS_PATH segment")
	} else if firstSeg.ASNs[0] != 64512 {
		t.Errorf("Prepend should add prepended ASN 64512, got %d", firstSeg.ASNs[0])
	}
}
