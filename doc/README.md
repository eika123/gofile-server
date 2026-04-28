# Documentation

This folder contains design and architecture notes for the file server.

- `architecture.md` explains the current separation of concerns in the app.
- `podman.md` explains the Podman deployment, reverse proxy setup, and persistent shared volume.

## How to use

Read `architecture.md` to understand:

- the role of `cmd/hello/main.go`
- how authentication is isolated in `internal/auth`
- how filesystem access is isolated in `internal/file_traverse`
- how UI rendering and static assets are isolated in `internal/ui`

For deployment, TLS is terminated at a reverse proxy like nginx, so use that.


You can add additional docs here as the project grows.
