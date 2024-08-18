# art.media platform

_- a turn-key UXR solution for creating high-fidelity, multi-platform 3D experiences -_


_art.media.platform_ ("amp") is an SDK for building multi-platform 3D and media-centric apps with pluggable infrastructure where artists and creators control and secure environments. The amp client exists as a [Unity](https://unreal.com) and [Unreal](https://unreal.com) app driven by an embedded [Golang](https://golang.org) dynamic runtime library (output by this repo). 

This means amp is a "turn-key" 3D-based user interface solution allowing you rapidly publish a native app on Windows, Mac, Linux, Android, iOS, and most XR headsets while also delivering the benefits of embedded and "headless" Go runtime instances with a footprint in single digit megabytes while also running native on SoC metal has been very encouraging.

This is interesting to:

- **Storage and content providers** can deploy visually stunning user experiences using amp, delivering an immersive media experience while attracting users. This is otherwise only available in AAA games and profoudly more painful through a pure web-based solution -- e.g. [IPFS](https://www.ipfs.com/), Amazon's [S3](https://aws.amazon.com/s3/).
- **Library and app developers** regard amp as a flexible UI framework that allows them to focus on their core value proposition. Easily add your own 2D or 3D custom UI components while you get multi-platform build support out of the box. E.g. data visualization, geographical and spatial linking.
- **Geo/Spatial workspaces** are common in geo/spatial centric applications, such as GIS, CAD, and BIM, where integrated 3D visualization is a core part of the user experience.  amp's Unity client natively integrates [Online Maps XR](https://infinity-code.com/assets/online-maps) mapping, allowing you to unify location-based datasets, spatially precise environments, high-fidelity 3D rendering, and extensible linking.

This repo is lightweight and dependency-free so that it can be added to your project without consequence. At a high level, the development workflow is:

1. Import [amp-sdk-go](https://github.com/art-media-platform/amp-sdk-go) in your Go project and expose your functionally as an `amp.App`.
2. Clone [amp-host-go](https://github.com/art-media-platform/amp-host-go) and embed and expose your Go packages within it.
3. `make build libarthost` (with your packages embedded within it).
4. Rapidly build a native Unity app using one of the amp.App templates, embedding `libarthost` within it.
5. At runtime, your app or resources within it are "pinned" via `amp://{yourAppNameID}/{yourSchema...}` while the amp UX runtime manages the user's perceptual experience of all actively pinned URIs.


## Acknowledgements & Props

- 

## Points of Interest

In suggested order of review for newcomers:

|                                                                                                   |                                                                                                                                                                                 |
| ------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [api.tag.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/tag/api.tag.go) | TODO                                                                 |
| [api.task.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/stdlib/task/api.task.go) | A wrapper for goroutines inspired by a conventional parent-child process model and is used throughout this SDK.                                                                 |
| [api.app.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go)           | Interfaces for developers looking to implement an `amp.App`, defining how state is requested, pushed, and merged.                                                               |
| [api.host.go](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.host.go)         | Defines `amp.Host` and its related types, what [`amp-host-go`](https://github.com/art-media-platform/amp-host-go) implements, and the abstraction that an `amp.App` plugs into. |

## What is `amp.App`?

[`amp.App`](https://github.com/art-media-platform/amp-sdk-go/blob/main/amp/api.app.go) is the plugin interface for amp.

Like a traditional OS service, an `amp.App` responds to queries it recognizes and operates on user data and system state. The stock amp runtime offers essential apps, such as file system access, and user account services. The less obvious power of amp is its extensibility. This is done by implementing the `amp.App` interface and registering it with the Go (or other transpile) based amp runtime. The amp runtime then manages the user's perceptual experience of all actively pinned cells on an amp compatible client.
