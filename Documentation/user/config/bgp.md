



## BGP






<hr />

<div class="dd">

<code>groups</code>  <i>[]<a href="#bgpgroup">BGPGroup</a></i>

</div>
<div class="dt">

List of BGP Peer groups. All peers *must* belong to a group
If a parameter is configured in the group and in the neighbor level, the neighbor is used

</div>

<hr />





## BGPGroup

Appears in:


- <code><a href="#bgp">BGP</a>.groups</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name for the group

</div>

<hr />

<div class="dd">

<code>local_address</code>  <i>string</i>

</div>
<div class="dt">

Local address for the peering

</div>

<hr />

<div class="dd">

<code>ttl</code>  <i>uint8</i>

</div>
<div class="dt">

Maximum allowed TTL for routes from peers belonging to this group

</div>

<hr />

<div class="dd">

<code>authentication_key</code>  <i>string</i>

</div>
<div class="dt">

MD5 secret for the session

</div>

<hr />

<div class="dd">

<code>peer_as</code>  <i>uint32</i>

</div>
<div class="dt">

Peer AS number

</div>

<hr />

<div class="dd">

<code>local_as</code>  <i>uint32</i>

</div>
<div class="dt">

Local AS number

</div>

<hr />

<div class="dd">

<code>hold_time</code>  <i>uint16</i>

</div>
<div class="dt">

Hold timer

</div>

<hr />

<div class="dd">

<code>multipath</code>  <i><a href="#multipath">Multipath</a></i>

</div>
<div class="dt">

Enables multipath routes

</div>

<hr />

<div class="dd">

<code>import</code>  <i>[]string</i>

</div>
<div class="dt">

List of import filters.
Example:
import: ["ACCEPT_ALL"]
# this example assumes that the a policy named ACCEPT_ALL exists in the configuration

</div>

<hr />

<div class="dd">

<code>export</code>  <i>[]string</i>

</div>
<div class="dt">

List of export filters. Syntax is the same as with import

</div>

<hr />

<div class="dd">

<code>route_server_client</code>  <i>bool</i>

</div>
<div class="dt">

Configures the daemon as a route server client

</div>

<hr />

<div class="dd">

<code>route_reflector_client</code>  <i>bool</i>

</div>
<div class="dt">

Configures the daemon as a route reflector client

</div>

<hr />

<div class="dd">

<code>cluster_id</code>  <i>string</i>

</div>
<div class="dt">

Cluster ID for route reflection

</div>

<hr />

<div class="dd">

<code>passive</code>  <i>bool</i>

</div>
<div class="dt">

Configures the client in passive mode

</div>

<hr />

<div class="dd">

<code>neighbors</code>  <i>[]<a href="#bgpneighbor">BGPNeighbor</a></i>

</div>
<div class="dt">

Neighbors that belong to this group. See bgpneighbors.md for details.

</div>

<hr />

<div class="dd">

<code>ipv4</code>  <i><a href="#addressfamilyconfig">AddressFamilyConfig</a></i>

</div>
<div class="dt">

Configuration values for the IPv4 AFI family

</div>

<hr />

<div class="dd">

<code>ipv6</code>  <i><a href="#addressfamilyconfig">AddressFamilyConfig</a></i>

</div>
<div class="dt">

Configuration values for the IPv6 AFI family

</div>

<hr />

<div class="dd">

<code>routing_instance</code>  <i>string</i>

</div>
<div class="dt">

Name of the routing instance this groups belongs to

</div>

<hr />





## Multipath

Appears in:


- <code><a href="#bgpgroup">BGPGroup</a>.multipath</code>

- <code><a href="#bgpneighbor">BGPNeighbor</a>.multipath</code>





<hr />

<div class="dd">

<code>enable</code>  <i>bool</i>

</div>
<div class="dt">

Enable multipath

</div>

<hr />

<div class="dd">

<code>multiple_as</code>  <i>bool</i>

</div>
<div class="dt">

Enable learning multiple paths for routes coming from different AS

</div>

<hr />





## BGPNeighbor

Appears in:


- <code><a href="#bgpgroup">BGPGroup</a>.neighbors</code>





<hr />

<div class="dd">

<code>peer_address</code>  <i>string</i>

</div>
<div class="dt">

Address for the peer

</div>

<hr />

<div class="dd">

<code>local_address</code>  <i>string</i>

</div>
<div class="dt">

Local address for the session

</div>

<hr />

<div class="dd">

<code>disabled</code>  <i>bool</i>

</div>
<div class="dt">

Disable the session with this peer

</div>

<hr />

<div class="dd">

<code>ttl</code>  <i>uint8</i>

</div>
<div class="dt">

Maximum allowed TTL for routes from peers belonging to this group

</div>

<hr />

<div class="dd">

<code>authentication_key</code>  <i>string</i>

</div>
<div class="dt">

MD5 secret for the session

</div>

<hr />

<div class="dd">

<code>peer_as</code>  <i>uint32</i>

</div>
<div class="dt">

Peer AS number

</div>

<hr />

<div class="dd">

<code>local_as</code>  <i>uint32</i>

</div>
<div class="dt">

Local AS number

</div>

<hr />

<div class="dd">

<code>hold_time</code>  <i>uint16</i>

</div>
<div class="dt">

Hold timer

</div>

<hr />

<div class="dd">

<code>multipath</code>  <i><a href="#multipath">Multipath</a></i>

</div>
<div class="dt">

Enables multipath routes
Allowed values:
- enable: enables the feature
- multiple_as: enables learning multiple paths for routes coming from different AS

</div>

<hr />

<div class="dd">

<code>import</code>  <i>[]string</i>

</div>
<div class="dt">

List of import filters.
Example:
import: ["ACCEPT_ALL"]
# this example assumes that the a policy named ACCEPT_ALL exists in the configuration

</div>

<hr />

<div class="dd">

<code>export</code>  <i>[]string</i>

</div>
<div class="dt">

List of export filters. Syntax is the same as with import

</div>

<hr />

<div class="dd">

<code>route_server_client</code>  <i>bool</i>

</div>
<div class="dt">

Configures the daemon as a route server client

</div>

<hr />

<div class="dd">

<code>route_reflector_client</code>  <i>bool</i>

</div>
<div class="dt">

Configures the daemon as a route reflector client

</div>

<hr />

<div class="dd">

<code>passive</code>  <i>bool</i>

</div>
<div class="dt">

Configures the client in passive mode

</div>

<hr />

<div class="dd">

<code>cluster_id</code>  <i>string</i>

</div>
<div class="dt">

Cluster ID for route reflection

</div>

<hr />

<div class="dd">

<code>ipv4</code>  <i><a href="#addressfamilyconfig">AddressFamilyConfig</a></i>

</div>
<div class="dt">

Configuration values for the IPv4 AFI family

</div>

<hr />

<div class="dd">

<code>ipv6</code>  <i><a href="#addressfamilyconfig">AddressFamilyConfig</a></i>

</div>
<div class="dt">

Configuration values for the IPv6 AFI family

</div>

<hr />

<div class="dd">

<code>advertise_ipv4_multiprotocol</code>  <i>bool</i>

</div>
<div class="dt">

Advertise the multiprotocol capability for the IPv4 AFI

</div>

<hr />

<div class="dd">

<code>routing_instance</code>  <i>string</i>

</div>
<div class="dt">

Name of the routing instance this groups belongs to

</div>

<hr />





## AddressFamilyConfig

Appears in:


- <code><a href="#bgpgroup">BGPGroup</a>.ipv4</code>

- <code><a href="#bgpgroup">BGPGroup</a>.ipv6</code>

- <code><a href="#bgpneighbor">BGPNeighbor</a>.ipv4</code>

- <code><a href="#bgpneighbor">BGPNeighbor</a>.ipv6</code>





<hr />

<div class="dd">

<code>add_path</code>  <i><a href="#addpathconfig">AddPathConfig</a></i>

</div>
<div class="dt">

Enable add_path for send and receive

</div>

<hr />

<div class="dd">

<code>next_hop_extended</code>  <i>bool</i>

</div>
<div class="dt">

Enable extended next hop for the address family

</div>

<hr />





## AddPathConfig

Appears in:


- <code><a href="#addressfamilyconfig">AddressFamilyConfig</a>.add_path</code>





<hr />

<div class="dd">

<code>receive</code>  <i>bool</i>

</div>
<div class="dt">

Enable receive add_path

</div>

<hr />

<div class="dd">

<code>send</code>  <i><a href="#addpathsendconfig">AddPathSendConfig</a></i>

</div>
<div class="dt">

Enable send add_path

</div>

<hr />





## AddPathSendConfig

Appears in:


- <code><a href="#addpathconfig">AddPathConfig</a>.send</code>





<hr />

<div class="dd">

<code>multipath</code>  <i>bool</i>

</div>
<div class="dt">

Enable multipath by add_path

</div>

<hr />

<div class="dd">

<code>path_count</code>  <i>uint8</i>

</div>
<div class="dt">

Maximum allowed path count for add_path

</div>

<hr />




