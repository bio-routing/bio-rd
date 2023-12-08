package config

import (
	"testing"
	"time"

	bnet "github.com/bio-routing/bio-rd/net"
	"github.com/bio-routing/bio-rd/routingtable/filter"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v2"
)

const (
	BGPGroupTestFile = `
groups:
  - name: Group1
    local_address: 100.64.0.1
    route_server_client: true
    route_reflector_client: true
    passive: true
    ttl: 10
    local_as: 65200
    peer_as: 65300
    authentication_key: secret
    hold_time: 180
    routing_instance: main
    import: ["ACCEPT_ALL"]
    export: ["REJECT_ALL"]
    neighbors:
      - peer_address: 100.64.0.2
        cluster_id: 100.64.0.0
        disabled: true
        ipv4:
          add_path:
            receive: true
            send:
              path_count: 5
      - peer_address: 100.64.1.2
        local_address: 100.64.1.1
        local_as: 65400
        peer_as: 65401
        hold_time: 90
        ttl: 1
        routing_instance: test
        export: ["ACCEPT_ALL"]
        import: ["REJECT_ALL"]
        authentication_key: top-secret
        passive: false
        route_reflector_client: false
        route_server_client: false
        ipv6:
          add_path:
            receive: true
            send:
              path_count: 2
    ipv6:
      add_path:
        receive: false
        send:
          path_count: 10
  `
)

func TestBGPLoad(t *testing.T) {
	policyOptions := &PolicyOptions{}
	accept_all := filter.NewFilter("ACCEPT_ALL", nil)
	reject_all := filter.NewFilter("REJECT_ALL", nil)
	policyOptions.PolicyStatementsFilter = []*filter.Filter{
		accept_all,
		reject_all,
	}

	b := []byte(BGPGroupTestFile)
	var bgp *BGP
	err := yaml.Unmarshal(b, &bgp)
	if err != nil {
		t.Fatalf("unexpected error while parsing: %s", err)
	}

	assert.Equal(t, 1, len(bgp.Groups), "group count")
	group := bgp.Groups[0]

	assert.Equal(t, 2, len(group.Neighbors), "neighbor count")

	err = bgp.load(64900, policyOptions)
	if err != nil {
		t.Fatalf("unexpected error while loading group config: %s", err)
	}

	n1 := group.Neighbors[0]
	assert.Equal(t, bnet.IPv4FromOctets(100, 64, 0, 1).Dedup(), n1.LocalAddressIP, "neighbor 1 local address")
	assert.Equal(t, bnet.IPv4FromOctets(100, 64, 0, 2).Dedup(), n1.PeerAddressIP, "neighbor 1 peer address")
	assert.True(t, *n1.RouteServerClient, "neighbor 1 route server client")
	assert.True(t, *n1.RouteReflectorClient, "neighbor 1 route reflector client")
	assert.True(t, *n1.Passive, "neighbor 1 passive")
	assert.Equal(t, uint8(10), n1.TTL, "neighbor 1 TTL")
	assert.Equal(t, uint32(65200), n1.LocalAS, "neighbor 1 local ASN")
	assert.Equal(t, uint32(65300), n1.PeerAS, "neighbor 1 peer ASN")
	assert.Equal(t, "secret", n1.AuthenticationKey, "neighbor 1 auth")
	assert.Equal(t, 180*time.Second, n1.HoldTimeDuration, "neighbor 1 hold")
	assert.Equal(t, "main", n1.RoutingInstance, "neighbor 1 VRF")
	assert.Equal(t, filter.Chain{accept_all}, n1.ImportFilterChain, "neighbor 1 import")
	assert.Equal(t, filter.Chain{reject_all}, n1.ExportFilterChain, "neighbor 1 export")
	assert.Equal(t, bnet.IPv4FromOctets(100, 64, 0, 0).Dedup(), n1.ClusterIDIP, "neighbor 1 cluster ID")
	assert.True(t, n1.IPv4.AddPath.Receive, "neighbor 1 IPv4 add path receive")
	assert.False(t, n1.IPv6.AddPath.Receive, "neighbor 1 IPv6 add path receive")
	assert.Equal(t, uint8(5), n1.IPv4.AddPath.Send.PathCount, "neighbor 1 IPv4 add path send count")
	assert.Equal(t, uint8(10), n1.IPv6.AddPath.Send.PathCount, "neighbor 1 IPv6 add path send count")
	assert.True(t, n1.Disabled, "neighbor 1 disabled")

	n2 := group.Neighbors[1]
	assert.Equal(t, bnet.IPv4FromOctets(100, 64, 1, 1).Dedup(), n2.LocalAddressIP, "neighbor 2 local address")
	assert.Equal(t, bnet.IPv4FromOctets(100, 64, 1, 2).Dedup(), n2.PeerAddressIP, "neighbor 2 peer address")
	assert.False(t, *n2.RouteServerClient, "neighbor 2 route server client")
	assert.False(t, *n2.RouteReflectorClient, "neighbor 2 route reflector client")
	assert.False(t, *n2.Passive, "neighbor 2 passive")
	assert.Equal(t, uint8(1), n2.TTL, "neighbor 2 TTL")
	assert.Equal(t, uint32(65400), n2.LocalAS, "neighbor 2 local ASN")
	assert.Equal(t, uint32(65401), n2.PeerAS, "neighbor 2 peer ASN")
	assert.Equal(t, "top-secret", n2.AuthenticationKey, "neighbor 2 auth")
	assert.Equal(t, 90*time.Second, n2.HoldTimeDuration, "neighbor 2 hold")
	assert.Equal(t, "test", n2.RoutingInstance, "neighbor 2 VRF")
	assert.Equal(t, filter.Chain{reject_all}, n2.ImportFilterChain, "neighbor 2 import")
	assert.Equal(t, filter.Chain{accept_all}, n2.ExportFilterChain, "neighbor 2 export")
	assert.Nil(t, n2.ClusterIDIP, "neighbor 2 cluster ID")
	assert.Nil(t, n2.IPv4, "neighbor 2 IPv4")
	assert.True(t, n2.IPv6.AddPath.Receive, "neighbor 2 IPv6 add path receive")
	assert.Equal(t, uint8(2), n2.IPv6.AddPath.Send.PathCount, "neighbor 2 IPv6 add path send count")
}
