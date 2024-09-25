# art.media.platform
_a fully secure turn-key solution for media deployment, 3D experiences and UX infrastructure that we all can agree on_


**_art.media.platform_** ("amp") is an SDK for building multi-platform 3D and media-centric apps with pluggable infrastructure where artists and creators control secure environments and "digital-twin" spatial and geographic linking.

The amp client exists as a [Unity](https://unreal.com) and [Unreal](https://unreal.com) cross-platform app powered by an embedded [Golang](https://golang.org) dynamic runtime library built by this and the [amp-host-go](https://github.com/art-media-platform/amp-host-go) repo.

This means amp is a "turn-key", 3D-based user interface solution allowing you rapidly publish a native app on Windows, Mac, Linux, Android, iOS, and most AR / VR headsets while also delivering the benefits of embedded and "headless" Go runtime instances with a footprint in single digit megabytes and rather respectable benchmarks.


## What does this framework solve?

Amp is 3D client-to-infrastructure bridge and is interesting to app developers looking to deploy engaging and memorable visuals while maintaining human engagement securely. This is somewhat available in AAA games but is notoriously difficult to deliver using limited web-based solutions such as [three.js](https://threejs.org/).  Seeing performance benchmarks when running "on the metal" are compelling.

The amp stack serves as a bridge that allows app developers to focus on their core value proposition. Easily add your own 2D or 3D custom UI components while you get multi-platform build support out of the box. -- e.g. data visualization, geographical and spatial linking.

Geographic-centric applications, such as GIS, CAD, and BIM, typically require integrated 3D as part of the user experience.  Amp's 3D client natively integrates [maps and locations](https://infinity-code.com/assets/online-maps), allowing you to unify location-based datasets, spatially precise environments, high-fidelity 3D rendering, and extensible linking -- pulling datasets from device, LAN, WAN, or MAN.


## What about Web?

Web is alive and well under this model.  Software such as [Webview](https://developer.vuplex.com/webview/overview) is embedded in the 3D client, which is essentially the Chromium browser, allowing an amp client to be a multi-function web browser out of the box.  Further, an in-app browser fits naturally when you think about spatial linking, where URLs or assets are anchored to map locations or 3D spatial locations.



## Project Workflow

This repo is lightweight and dependency-free so that it can be added to your project without consequence. At a high level, the development workflow is:

1. Import [amp-sdk-go](https://github.com/art-media-platform/amp-sdk-go) in your Go project and expose your functionally by implementing  [`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) however you like.
2. Clone [amp-host-go](https://github.com/art-media-platform/amp-host-go) (currently private) and add your Go packages to it (like adding a module to C++ project and registering it either statically or dynamically).
3. `make build libarthost` with your packages embedded within it
    - `make build arthost` builds the host as "headless" server or peer node.
4. In your rapidly (or meticulously) native Unity or Unreal app, embed `libarthost` within it.
5. At runtime, your package services URL requests that are "pinned" via the [`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) interface.  Meanwhile, the amp UX runtime manages the user's perceptual experience of all actively pinned URIs in addition to providing a toolbox of extensions and examples.

## Points of Interest

In suggested order of review for newcomers:

|                                                                                                   |                                                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [api.tag.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/tag/api.tag.go)    | small versatile hash tag scheme offering easy interoperability                                                                                                                  |
| [api.task.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/task/api.task.go) | wrapper for goroutines inspired by a conventional parent-child process model                                                                                                    |
| [api.app.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go)           | `amp.App` types and interfaces defining how state is requested, pushed, and merged                                                                                              |
| [api.host.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.host.go)         | `amp.Host` types and interfaces that [`amp-host-go`](https://github.com/art-media-platform/amp-host-go) implements                                                              |

## What is `amp.App`?

[`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) is the plugin interface for amp.

Like a traditional OS service, an `amp.App` responds to queries it recognizes and operates on user data and system state. The stock amp runtime offers essential apps, such as file system access, and user account services. The less obvious power of amp is its extensibility. This is done by implementing the `amp.App` interface and registering it with the Go (or other transpile) based amp runtime. The amp runtime then manages the user's perceptual experience of all actively pinned cells on an amp compatible client.


## Acknowledgements
- [AVPro Video](https://renderheads.com/products/avpro-video/) by RenderHeads
- [Webview](https://developer.vuplex.com/webview/overview) by Vuplex
- [Online Maps](https://infinity-code.com/doxygen/online-maps/) by Infinity Code
- [FMOD](https://www.fmod.com/) by Firelight Technologies