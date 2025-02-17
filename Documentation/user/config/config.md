



## Config






<hr />

<div class="dd">

<code>policy_options</code>  <i>PolicyOptions</i>

</div>
<div class="dt">

A list of policy statements, filters and prefix lists that can be used to filter route imports and exports.
<a href="policy.md">Detailed documentation</a>
Example:
policy_statements:
  - name: "PeerA-In"
      terms:
        - name: "Reject_certain_stuff"
          from:
            route_filters:
               - prefix: "198.51.100.0/24"
                 matcher: "orlonger"
               - prefix: "203.0.113.0/25"
                 matcher: "exact"
               - prefix: "203.0.113.128/25"
                 matcher: "exact"

</div>

<hr />

<div class="dd">

<code>routing_instances</code>  <i>[]RoutingInstance</i>

</div>
<div class="dt">

List of routing instances
<a href="routing_instance.md">Configuration parameters</a>

</div>

<hr />

<div class="dd">

<code>routing_options</code>  <i>RoutingOptions</i>

</div>
<div class="dt">

Routing options
<a href="routing_options.md">parameter documentation</a>
Allowed values:
- <a href="static_route.md">static_routes</a>
- router_id
- autonomous_system

</div>

<hr />

<div class="dd">

<code>protocols</code>  <i>Protocols</i>

</div>
<div class="dt">

Here you define the specific configuration parameters for each protocol.
<a href="protocols.md">documentation</a>
Available protocols:
  - <a href="bgp.md">bgp</a>
  - <a href="isis.md">is-is</a>

</div>

<hr />




