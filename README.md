# AMP / Arc Media Platform

_A turnkey solution for creating high-fidelity, multi-platform 3D experiences_

## What is AMP?

This repo contains Go interfaces and support for the _Arc Media Platform_ ("AMP"), a SDK for building multi-platform 3D and media-centric apps with pluggable infrastructure. The AMP client is a [Unity](https://unity.com) based app driven by an embedded [Go](https://golang.org) runtime.  This means AMP is a "turnkey" 3D-based user interface solution allowing you rapidly publish a native app on Windows, macOS, Linux, Android, iOS, and most XR headsets while also delivering the benefits of Go.

This is interesting to:
  - **Storage and content providers**: AMP allows storage and content providers to create visually stunning user experience that deliver an immersive media experience, attracting and engaging users like never before, otherwise only seen in AAA games and generally not possible through a pure web-based solution.  E.g. [IPFS](https://www.ipfs.com/), the [Internet Computer](https://dfinity.org/), and Amazon's [S3](https://aws.amazon.com/s3/).  
  - **Library and app developers**:  AMP uniquely improves app development for library and app developers by providing a flexible UI framework that allows them to focus on their core value proposition. Easily integrate custom UI components and get multi-platform support out of the box.  E.g. data visualization, geographical and spatial linking.
  
This repo is lightweight and dependency-free so that it can be added to your project without consequence.  At a high level, the development workflow is:

  1. Import [`go-arc-sdk`](https://github.com/arcspace/go-arc-sdk) in your Go project and expose your functionally as an `arc.App`.
  2. Clone [`go-archost`](https://github.com/arcspace/go-archost), register your app alongside other `arc` apps you wish to ship with, and build `libarchost` (with your package embedded within it).
  3. Rapidly build a native Unity app using one of the AMP app templates, with `libarchost` embedded within it.
  4. At runtime, the Unity client any root `arc.Cell` of your app is "pinned" via `arc://{yourAppUID}/{yourSchema...}` while the AMP UX runtime manages the user's perceptual experience of all actively pinned cells.
  
## Points of Interest

In suggested order of review for newcomers:

|                          |                                                                   |
|------------------------- | ------------------------------------------------------------------|
| [api.task.go](https://github.com/arcspace/go-arc-sdk/blob/main/stdlib/task/api.task.go)        | A wrapper for goroutines inspired by a conventional parent-child process model and is used throughout this SDK.              |
| [api.app.go](https://github.com/arcspace/go-arc-sdk/blob/main/apis/arc/api.app.go)    | Interfaces for developers looking to implement an `arc.App`, defining how state is requested, pushed, and merged.                |
| [api.host.go](https://github.com/arcspace/go-arc-sdk/blob/main/apis/arc/api.host.go) | Defines `arc.Host` and its related types, what [`go-archost`](https://github.com/arcspace/go-archost) implements, and the abstraction that an `arc.App` plugs into.             |

## What is `arc.App`?

[`arc.App`](https://github.com/arcspace/go-arc-sdk/blob/main/apis/arc/api.app.go) is the plugin interface for AMP.  

Like a traditional OS service, an `arc.App` responds to queries it recognizes and operates on user data and system state.   The stock AMP runtime offers essential apps, such as file system access, and user account services.  The less obvious power of AMP its extensibility. This is done by implementing the `arc.App` interface and registering it with the Go-based AMP runtime.  The AMP runtime then manages the user's perceptual experience of all actively pinned cells on an AMP-compatible client.