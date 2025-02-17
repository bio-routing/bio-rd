



## PolicyOptions






<hr />

<div class="dd">

<code>policy_statements</code>  <i>[]<a href="#policystatement">PolicyStatement</a></i>

</div>
<div class="dt">

Policy statements to filter route imports and exports
Example:
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

<code>prefix_lists</code>  <i>[]<a href="#prefixlist">PrefixList</a></i>

</div>
<div class="dt">

prefix lists to be used in the policy statements.
Example:
  prefix-lists:
    prefixes:
      - 2001:db8:0:1::/64

</div>

<hr />





## PrefixList

Appears in:


- <code><a href="#policyoptions">PolicyOptions</a>.prefix_lists</code>





<hr />

<div class="dd">

<code>prefixes</code>  <i>[]string</i>

</div>
<div class="dt">

List of prefixes

</div>

<hr />





## PolicyStatement

Appears in:


- <code><a href="#policyoptions">PolicyOptions</a>.policy_statements</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the policy statement

</div>

<hr />

<div class="dd">

<code>terms</code>  <i>[]<a href="#policystatementterm">PolicyStatementTerm</a></i>

</div>
<div class="dt">

List of terms defining the policy (see example above)

</div>

<hr />





## PolicyStatementTerm

Appears in:


- <code><a href="#policystatement">PolicyStatement</a>.terms</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the term

</div>

<hr />

<div class="dd">

<code>from</code>  <i><a href="#policystatementtermfrom">PolicyStatementTermFrom</a></i>

</div>
<div class="dt">

Filter to match the term

</div>

<hr />

<div class="dd">

<code>then</code>  <i><a href="#policystatementtermthen">PolicyStatementTermThen</a></i>

</div>
<div class="dt">

Action to execute if the filter matches
Available actions are:
  - Accept: accepts the route without modifications
  - Reject: rejects the route
  - MED: sets the MED to the specified value (max 4294967295)
  - LocalPref sets the local preference to the specified value (max 4294967295)
  - AsPathPrepend: prepends AS numbers to the route. Details bellow
  - NextHop: modify the next-hop to the specified address

</div>

<hr />





## PolicyStatementTermFrom

Appears in:


- <code><a href="#policystatementterm">PolicyStatementTerm</a>.from</code>





<hr />

<div class="dd">

<code>route_filters</code>  <i>[]<a href="#routefilter">RouteFilter</a></i>

</div>
<div class="dt">

List of route filters to match incoming packets
Example:
  route_filters:
     - prefix: "198.51.100.0/24"
       matcher: "orlonger"

</div>

<hr />





## RouteFilter

Appears in:


- <code><a href="#policystatementtermfrom">PolicyStatementTermFrom</a>.route_filters</code>





<hr />

<div class="dd">

<code>prefix</code>  <i>string</i>

</div>
<div class="dt">

Prefix to match. Defined in CIDR notation

</div>

<hr />

<div class="dd">

<code>matcher</code>  <i>string</i>

</div>
<div class="dt">

Qualifier for the filter.
Available options:
  - exact: matches only the exact prefix
  - orlonger: matches if the prefix equals the filter, or has a longer prefix length
  - longer: matches if the route has a longer prefix length than the one defined in the filter
  - range: the route falls between the values defined in len_min and len_max

</div>

<hr />

<div class="dd">

<code>len_min</code>  <i>uint8</i>

</div>
<div class="dt">

minimum length of the range

</div>

<hr />

<div class="dd">

<code>len_max</code>  <i>uint8</i>

</div>
<div class="dt">

maximum lange of the range

</div>

<hr />





## PolicyStatementTermThen

Appears in:


- <code><a href="#policystatementterm">PolicyStatementTerm</a>.then</code>





<hr />

<div class="dd">

<code>accept</code>  <i>bool</i>

</div>
<div class="dt">

accept the route

</div>

<hr />

<div class="dd">

<code>reject</code>  <i>bool</i>

</div>
<div class="dt">

reject the route

</div>

<hr />

<div class="dd">

<code>med</code>  <i>uint32</i>

</div>
<div class="dt">

Multi-exit discriminator

</div>

<hr />

<div class="dd">

<code>local_pref</code>  <i>uint32</i>

</div>
<div class="dt">

Local preference

</div>

<hr />

<div class="dd">

<code>as_path_prepend</code>  <i><a href="#aspathprepend">ASPathPrepend</a></i>

</div>
<div class="dt">

ASN to prepend.
  Values:
    - ASN: asn number to prepend
    - count: amount of times that the number will be prepended

</div>

<hr />

<div class="dd">

<code>next_hop</code>  <i><a href="#nexthop">NextHop</a></i>

</div>
<div class="dt">

IP address to be used as a next-hop for the route

</div>

<hr />





## ASPathPrepend

Appears in:


- <code><a href="#policystatementtermthen">PolicyStatementTermThen</a>.as_path_prepend</code>





<hr />

<div class="dd">

<code>asn</code>  <i>uint32</i>

</div>
<div class="dt">

AS number

</div>

<hr />

<div class="dd">

<code>count</code>  <i>uint16</i>

</div>
<div class="dt">

times to prepend

</div>

<hr />





## NextHop

Appears in:


- <code><a href="#policystatementtermthen">PolicyStatementTermThen</a>.next_hop</code>





<hr />

<div class="dd">

<code>address</code>  <i>string</i>

</div>
<div class="dt">

IP address to be used as next hop

</div>

<hr />




