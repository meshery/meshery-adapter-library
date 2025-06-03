# Meshery Adapter Library
The Meshery Adapter Library provides a common and consistent set of functionality that Meshery adapters use for managing the lifecycle, configuration, operation, and performance of cloud native infrastructure. See [Introducing MeshKit and the Meshery Adapter Library](https://layer5.io/blog/meshery/introducing-meshkit-and-the-meshery-adapter-library) for more.

## Purpose 

The main purpose of the meshery-adapter-library is to 
* provide a set of interfaces, some with default implementations, to be used and extended by adapters.
* implement common cross cutting concerns like logging, errors, and tracing
* provide a mini framework implementing the gRPC server that allows plugging in the mesh specific configuration and 
    operations implemented in the adapters.

### Overview and usage 

The library consists of interfaces and default implementations for the main and common functionality of an adapter. 
It also provides a mini-framework that runs the gRPC adapter service, calling the functions of handlers injected by the 
adapter code. This is represented in an UML-ish style in the figure below. The library is used in the Consul adapter, 
and others will follow. 

<img alt="Overview and usage of meshery-adapter-library" src="./doc/meshery-adapter-library-overview.png" align="center"/>

### Package dependencies hierarchy
A clear picture of dependencies between packages in a module helps avoid circular dependencies (import cycles), 
understand where to put code, design coherent packages etc.

Referring to the figure below, the packages `config` and `meshes` (which contains the adapter service proto definition) 
are at the top of the dependency hierarchy and can be used by any other package. Thinking in layers (L), `config`  
would be in the top layer, L1, `adapter` in L2, and `config/provider`  in L3. Packages can always be imported and used in lower layers.

<img alt="Package dependencies hierarchy" src="./doc/mesher-adapter-library-package-dependencies.png" align="center"/>

<div>&nbsp;</div>

## Join the Meshery community!

<a name="contributing"></a><a name="community"></a>
Our projects are community-built and welcome collaboration. 👍 Be sure to see the <a href="https://docs.meshery.io/project/contributing#not-sure-where-to-start">Contributor Welcome Guide</a> for a tour of resources available to you and jump into our <a href="http://slack.meshery.io">Slack</a>!
<p style="clear:both;">

<h3>Find your MeshMate</h3>

<p>MeshMates are experienced community members, who will help you learn your way around, discover live projects and expand your community network. 
Become a <b>Meshtee</b> today!</p>

Find out more on the <a href="https://meshery.io/community#meshmates">Meshery community</a>. <br />
<br /><br /><br /><br />
</p>

<div>&nbsp;</div>

<a href="https://meshery.io/community"><img alt="Meshery Cloud Native Community" src=".github/readme/images//slack-128.png" style="margin-left:10px;padding-top:5px;" width="110px" align="right" /></a>

<a href="http://slack.meshery.io"><img alt="Meshery Cloud Native Community" src=".github/readme/images//community.svg" style="margin-right:8px;padding-top:5px;" width="140px" align="left" /></a>

<p>
✔️ <em><strong>Join</strong></em> any or all of the weekly meetings on <a href="https://meshery.io/calendar">community calendar</a>.<br />
✔️ <em><strong>Watch</strong></em> community <a href="https://www.youtube.com/@mesheryio?sub_confirmation=1">meeting recordings</a>.<br />
✔️ <em><strong>Fill-in</strong></em> a <a href="https://meshery.io/newcomers">community member form</a> to gain access to community resources.
<br />
✔️ <em><strong>Discuss</strong></em> in the <a href="https://meshery.io/community#discussion-forums">Community Forum</a>.<br />
✔️ <em><strong>Explore more</strong></em> in the <a href="https://meshery.io/community#handbook">Community Handbook</a>.<br />
</p>
<p align="center">
<i>Not sure where to start?</i> Grab an open issue with the <a href="https://github.com/issues?q=is%3Aopen+is%3Aissue+archived%3Afalse+org%meshery+org%3Ameshery+org%3Aservice-mesh-performance+org%3Aservice-mesh-patterns+label%3A%22help+wanted%22+">help-wanted label</a>.
</p>
