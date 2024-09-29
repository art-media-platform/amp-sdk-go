# ART.MEDIA.PLATFORM

## Allowing for community driven 3D experiences
<u>_bridging UX infrastructure with community input to allow 
a network of users to collaborate seamlessly in real time on dynamic 3D environments._</u>

#### A Open Source SDK for 3D media deployment.

**AMP** ("**_art.media.platform_**") is an SDK for building community oriented 3D asset management solutions. The SDK is multi-platform, and media-centric. It can have infrastructure both plug in their own assets and embeds, as well share it with other artists and networks using the SDK. 

The AMP SDK has been integrated with clients in [Unity](https://unreal.com) and [Unreal](https://unreal.com), this is done by embedding the powered by the embedded [Golang](https://golang.org) dynamic runtime library created by this SDK (specified by this repo).

Having the embed be written in **Go** allows developers to you to publish a native app on Windows, Mac, Linux, Android, iOS, and most AR / VR headsets. 

While the packages was created for integrating with 3D clients, you can also use this SDK in a "headless" Go runtime, allowing you to integrate the SDK in a distributed server environment independant of your client as well. 

With Go's minimal footprint measured in single digit megabytes, and impressive impressive performance benchmarks, allow it to scale for concurrent distributed systems.

## _What does this framework solve?_

***AMP's primary goal is to act as a bridge between 3D clients to system, network, and infrastructure resources.***

It impliments a handful of opinionated custom integrations that make it easy to spin up distributed data solutions for the 3D interfaces that you choose to integrate with. 

The core of the AMP SDK is a tag system that makes it possible to create a seperate descriptive metadata for scenes and assets, making it possible to reuse graphic assets already downloaded in the system as components described by the markup to create scenes dynamically. This limits the ammount of times needed to actually transfer heavy 3D asset information across the network to the client. 

We use this mark up to store the state for the client on the server/ This paradigm allows for security, performance, and community driven user experiences. 

The SDK also comes with a convenient set of utilities, concurrancy support with the amp.task package.

We hope to evolve this SDK to a compelling if opinionated bridge for app developers looking to deploy engaging and memorable experiences -- all while maintaining human engagement securely. Delivering what is has somewhat been solved in a handful of AAA games but is notoriously difficult to deliver using limited web-based solutions such as [three.js](https://threejs.org/).  

In most cases assets can be multiple gigabytes, forcing a refetch on every page load when most community wide 3D titles have deployments in _double-digit gigabytes._ Many developers also find themselves being locked in by the web browser sandbox will having to go through major leaps or waiting on browser developers lock them out of the control they may need for features that the AMP sdk makes readily available.

We hope to serve the devs in this ecosystem serves as a bridge that allows app developers to focus on their core value proposition, and empower developers everywhere to add your own 2D or 3D custom UI components while you get multi-platform build support and content deployment out of the box.

### _What about web?_

Web is alive and well in this stack.  We embedded [Webview](https://developer.vuplex.com/webview/overview) in the 3D client, to aid allowing your app to have an embedded web browser out of the box to in bridging with your existing web 2 ecosystem. 

We hope that through we can make the transition to web 3 easier with consideration for rich tooling that has been developed for the web ecoysystem, and are thinking of more ways to integrate with existing ecosystems to make it possible for industries to really evolve and adopt these innovative solutions without needing to abandon all the tech they have already developed.

## GIS, ERP, CRM but in 3D

Beyond social applications, we find that the paradigm also serves to bring these richer UX experiences to enterprise system domains GIS, ERP, and CRM system data points to act as nodes for a part of the metadata describing these representational objects. 

To aid with this we have added native support [maps and locations](https://infinity-code.com/assets/online-maps) in AMP's 3D client. This allowed us to unify location-based datasets, spatially precise environments, with high-fidelity 3D rendering, and extensible linking -- pulling datasets from device, LAN, WAN, or MAN.

## Modeling, CAD, AI and Data Analytics but in 3D

We are constantly discovering more domains where this paradigm could apply. In the same way the web dev ecosystems discovered that the react component paradigm was able to describe components on the mobile ecosystem, we want to be able to bridge this generic representational data layer to do this. 

Right now we are exploring integrations with CAD systems where multiple nodes can to be used to describe different parts of a CAD model with a lot of needed meta data information separate from the presentation layer. In this process we are discovering that this data model can have applications in scientific modeling, AI, and data analytics spaces. 



## Integration Specifications

This repo is lightweight and dependency-free so that it can be added to your project without consequence. At a high level:

1. Add [amp-sdk-go](https://github.com/art-media-platform/amp-sdk-go) to your Go project and expose your functionally by implementing [`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) however you like.
2. Clone / Fork [amp-host-go](https://github.com/art-media-platform/amp-host-go) (currently private) and consume your `amp.App`, similar to adding a library to a C++ project and importing it statically or dynamically.
3. `make build libarthost` (with your packages embedded within it)
    - Or, `make build arthost` which builds the host as "headless" server or peer node
4. In your rapidly (or meticulously) designed native Unity or Unreal app, link `libarthost`
5. At runtime, your `amp.App` services URL requests that are "pinned" via the [`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) interface.  Meanwhile, the amp UX runtime manages the user's perceptual experience of all actively pinned URLs in addition to providing a toolbox of extensions and examples.

## Points of Interest

|                                                                                                   |                                                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [api.tag.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/tag/api.tag.go)    | small versatile hash tag scheme offering easy interoperability                                                                                                                  |
| [api.task.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/task/api.task.go) | wrapper for goroutines inspired by a conventional parent-child process model                                                                                                    |
| [api.app.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go)           | `amp.App` types and interfaces defining how state is requested, pushed, and merged                                                                                              |
| [api.host.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.host.go)         | `amp.Host` types and interfaces that [`amp-host-go`](https://github.com/art-media-platform/amp-host-go) implements                                                              |

## What is `amp.App`?

[`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) is the primary plugin interface for system runtime support as well as for third parties.  On startup, [`amp.Host`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.host.go) and instantiates registered `amp.App` instances as needed.

Similar to a traditional OS service, an `amp.App` responds to queries it recognizes and operates on user data and system state. The stock amp runtime offers essential apps, such as file system access, media service, and user account services.

The less obvious power of amp is its _extensibility_. The [`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) interface is flexible and unrestricted, allowing you to wrap anything that Go can link against.  This means any Go, C, C++, or any native static or dynamic executable can be wrapped and push a 3D-native UX (with stock or custom assets).


## Acknowledgements
- [AVPro Video](https://renderheads.com/products/avpro-video/) by RenderHeads
- [Webview](https://developer.vuplex.com/webview/overview) by Vuplex
- [Online Maps](https://infinity-code.com/doxygen/online-maps/) by Infinity Code
- [FMOD](https://www.fmod.com/) by Firelight Technologies
