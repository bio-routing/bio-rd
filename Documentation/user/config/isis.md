



## ISIS
ISIS config






<hr />

<div class="dd">

<code>NETs</code>  <i>[]string</i>

</div>
<div class="dt">

Network entity title for this instance

</div>

<hr />

<div class="dd">

<code>level1</code>  <i><a href="#isislevel">ISISLevel</a></i>

</div>
<div class="dt">

Configuration for the Level 1 adjacency

</div>

<hr />

<div class="dd">

<code>level2</code>  <i><a href="#isislevel">ISISLevel</a></i>

</div>
<div class="dt">

Configuration for the level 2 adjacency

</div>

<hr />

<div class="dd">

<code>interfaces</code>  <i>[]<a href="#isisinterface">ISISInterface</a></i>

</div>
<div class="dt">

Interface related configuration

</div>

<hr />

<div class="dd">

<code>lsp_lifetime</code>  <i>uint16</i>

</div>
<div class="dt">

Amount of time a link-state PDU should persist in the network
Expressed in seconds

</div>

<hr />





## ISISLevel
ISISLevel level config

Appears in:


- <code><a href="#isis">ISIS</a>.level1</code>

- <code><a href="#isis">ISIS</a>.level2</code>





<hr />

<div class="dd">

<code>disable</code>  <i>bool</i>

</div>
<div class="dt">

Disables this level for the instance

</div>

<hr />

<div class="dd">

<code>authentication_key</code>  <i>string</i>

</div>
<div class="dt">

Password for authentication

</div>

<hr />

<div class="dd">

<code>no_csnp_authentication</code>  <i>bool</i>

</div>
<div class="dt">

Disable authentication for the Complete Sequence Number PDUs

</div>

<hr />

<div class="dd">

<code>no_hello_authentication</code>  <i>bool</i>

</div>
<div class="dt">

DIsable authentication for hello messages

</div>

<hr />

<div class="dd">

<code>no_psnp_authentication</code>  <i>bool</i>

</div>
<div class="dt">

Disable authentication for the Partial Sequence Number PDUs

</div>

<hr />

<div class="dd">

<code>wide_metrics_only</code>  <i>bool</i>

</div>
<div class="dt">

Enable sending and receiving wide metrics only for this level

</div>

<hr />





## ISISInterface
ISISInterface interface config

Appears in:


- <code><a href="#isis">ISIS</a>.interfaces</code>





<hr />

<div class="dd">

<code>name</code>  <i>string</i>

</div>
<div class="dt">

Name of the interface to configure

</div>

<hr />

<div class="dd">

<code>passive</code>  <i>bool</i>

</div>
<div class="dt">

Configure interface as passive

</div>

<hr />

<div class="dd">

<code>point_to_point</code>  <i>bool</i>

</div>
<div class="dt">

Configure interface as point-to-point

</div>

<hr />

<div class="dd">

<code>level1</code>  <i><a href="#isisinterfacelevel">ISISInterfaceLevel</a></i>

</div>
<div class="dt">

Level 1 configuration parameters for the interface

</div>

<hr />

<div class="dd">

<code>level2</code>  <i><a href="#isisinterfacelevel">ISISInterfaceLevel</a></i>

</div>
<div class="dt">

Level 2 configuration parameters for the interface

</div>

<hr />





## ISISInterfaceLevel
ISISInterfaceLevel interface level config

Appears in:


- <code><a href="#isisinterface">ISISInterface</a>.level1</code>

- <code><a href="#isisinterface">ISISInterface</a>.level2</code>





<hr />

<div class="dd">

<code>disable</code>  <i>bool</i>

</div>
<div class="dt">

Disable this level for the interface

</div>

<hr />

<div class="dd">

<code>hello_interval</code>  <i>uint16</i>

</div>
<div class="dt">

Hello interval
Expressed in seconds

</div>

<hr />

<div class="dd">

<code>hold_time</code>  <i>uint16</i>

</div>
<div class="dt">

Hold time
Expressed in seconds

</div>

<hr />

<div class="dd">

<code>metric</code>  <i>uint32</i>

</div>
<div class="dt">

Metric for the interface in this level

</div>

<hr />

<div class="dd">

<code>passive</code>  <i>bool</i>

</div>
<div class="dt">

Configures interface as passive

</div>

<hr />

<div class="dd">

<code>priority</code>  <i>uint8</i>

</div>
<div class="dt">

Configures the device priority to become a designated router for this level
Value range: 0-127

</div>

<hr />




