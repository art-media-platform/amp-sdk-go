# art.media.platform
_a turn-key UX and media deployment solution for 3D experiential apps_


**_art.media.platform_** ("amp") is an SDK for building multi-platform 3D and media-centric apps with pluggable infrastructure where artists and creators control secure environments and "digital-twin" spatial and geographic linking.

The amp client exists as a [Unity](https://unreal.com) and [Unreal](https://unreal.com) multiplatform app powered by an embedded [Golang](https://golang.org) dynamic runtime library built by this repo.

This means amp is a "turn-key", 3D-based user interface solution allowing you rapidly publish a native app on Windows, Mac, Linux, Android, iOS, and most AR / VR headsets while also delivering the benefits of embedded and "headless" Go runtime instances with a footprint in single digit megabytes and rather respectable benchmarks.

## What does this framework solve?

Amp is 3D client-to-infrastructure bridge and is interesting to app developers looking to deploy engaging and memorable visuals while maintaining human engagement securely. This is somewhat available in AAA games but is notoriously difficult to deliver using limited web-based solutions such as [three.js](https://threejs.org/).

The amp stack serves as a bridge that allows app developers to focus on their core value proposition. Easily add your own 2D or 3D custom UI components while you get multi-platform build support out of the box. E.g. data visualization, geographical and spatial linking.

Geographic-centric applications, such as GIS, CAD, and BIM, typically require integrated 3D as part of the user experience.  Amp's 3D client natively integrates [maps and locations](https://infinity-code.com/assets/online-maps), allowing you to unify location-based datasets, spatially precise environments, high-fidelity 3D rendering, and extensible linking.


## Workflow

This repo is lightweight and dependency-free so that it can be added to your project without consequence. At a high level, the development workflow is:

1. Import [amp-sdk-go](https://github.com/art-media-platform/amp-sdk-go) in your Go project and expose your functionally as an `amp.App`.
2. Clone [amp-host-go](https://github.com/art-media-platform/amp-host-go) and embed and expose your Go packages within it.
3. `make build libarthost` (with your packages embedded within it).
4. Rapidly build a native Unity app using one of the amp.App templates, embedding `libarthost` within it.
5. At runtime, your app or resources within it are "pinned" via `amp://{yourAppNameID}/{yourSchema...}` while the amp UX runtime manages the user's perceptual experience of all actively pinned URIs.




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