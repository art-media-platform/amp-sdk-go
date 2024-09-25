# art.media.platform
_A fully secure and provisioned solution for media deployment, 3D experiences and UX infrastructure that we all can agree on._


**_art.media.platform_** ("amp") is an SDK for building multi-platform 3D and media-centric apps with pluggable infrastructure where artists and creators control secure environments and "digital-twin" spatial and geographic linking.

The amp client exists as a [Unity](https://unreal.com) and [Unreal](https://unreal.com) any-platform app, powered by an embedded [Golang](https://golang.org) dynamic runtime library (specified by this repo).

This means amp is a "turn-key" 3D-based user experience solution allowing you to publish a native app on Windows, Mac, Linux, Android, iOS, and most AR / VR headsets.  Further, the underlying embedded or "headless" Go runtime has a footprint measured in single digit megabytes while achieving impressive benchmarks that scale.


## What does this framework solve?

Amp is 3D client-to-infrastructure bridge and is interesting to app developers looking to deploy engaging and memorable experiences -- all while maintaining human engagement securely. This is somewhat available in AAA games but is notoriously difficult to deliver using limited web-based solutions such as [three.js](https://threejs.org/).  For example, 3D experiences on the web are problematic once assets are more than several _megabytes_.  Meanwhile, most 3D titles have deployments in _double-digit gigabytes._    Worse, the web browser sandbox can pose major blocks that a publisher has no ability to control.

This stack serves as a bridge that allows app developers to focus on their core value proposition. Easily add your own 2D or 3D custom UI components while you get multi-platform build support and content deployment out of the box.

Geographic-centric applications, such as GIS, CAD, and BIM, typically require integrated 3D as part of the user experience.  Amp's 3D client natively integrates [maps and locations](https://infinity-code.com/assets/online-maps), allowing you to unify location-based datasets, spatially precise environments, high-fidelity 3D rendering, and extensible linking -- pulling datasets from device, LAN, WAN, or MAN.


## What about web?

Web is alive and well in this stack.  Software such as [Webview](https://developer.vuplex.com/webview/overview) is embedded in the 3D client, which is essentially the Chromium browser, allowing your app to have an embedded web browser out of the box.  _An in-app integrated web browser is optimal when you think about spatial linking, where URLs or assets are anchored to map or 3D / spatial locations._



## Integration Workflow

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