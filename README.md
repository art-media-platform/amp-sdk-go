# go-arc-sdk

This repo contains Go types, interfaces, and type support for wrapping your Go package as an [`arc.App`](https://github.com/arcspace/go-arc-sdk/blob/main/apis/arc/api.app.go).  This repo is lightweight and dependency-free so that it can be consumed by a project without consequence.

## Why?

An `arc.App` essentially plugs into AMP, a 3D-engine based "turn-key" UX solution that lets you rapidly publish a native app on Windows, macOS, Linux, Android, iOS, and most XR headsets.

This is interesting to:
  - **Storage and content providers**, centralized or peer-based, looking to offer an alternative (non web-based) UI that offers the user an immersive media experience, complete with compelling UI flexibilty.   Examples are [IPFS](https://www.ipfs.com/), the [Internet Computer](https://dfinity.org/), and Amazon's [S3](https://aws.amazon.com/s3/).
  - **Library and app developers** looking for a flexible UI solution that allows them to focus on their value-proposition while allowing custom UI components as needed (e.g data visualization, geographical and spatial linking).
  
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
