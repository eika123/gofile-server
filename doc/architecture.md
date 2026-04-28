# Architecture and Separation of Concerns

This project is structured to separate HTTP routing, authentication, filesystem access, and UI rendering so each part can evolve independently.

## `cmd/hello/main.go`

This is the application entry point and HTTP router.

- Loads environment variables and configuration.
- Builds the `chi` router.
- Registers routes:
  - `/hello` for a simple health/demo endpoint.
  - `/files` for the directory browsing handler.
  - `/static/*` for embedded CSS/JS assets.
- Keeps request handling separate from lower-level business logic.

## `internal/auth`

Contains reusable middleware for HTTP Basic Auth.

- `NewBasicAuth(username, password)` returns middleware that wraps any `http.Handler`.
- The middleware is registered in `main.Router`.
- This keeps authentication logic out of the handler and makes it reusable.

## `internal/file_traverse`

Contains filesystem traversal utilities.

- `ListFiles` and `ListSubDirectories` enumerate directory contents.
- `GetFileContent` reads a file for download.
- These functions are independent of HTTP and can be tested separately.

## `cmd/hello/main.go` handler helpers

The `handleDisplayDirectoryContents` handler uses helper functions to keep flow clear:

- `parseRequestedPath` extracts the `path` query parameter.
- `resolveRequestedPath` normalizes the path against `ROOT_PATH` and validates it.
- `getDisplayPath` computes the user-visible path with `ROOT_PATH` stripped.
- `serveFile` streams a regular file to the response.
- `listDirectoryContents` loads directory contents from `internal/file_traverse`.

This means the handler itself is responsible for HTTP semantics, while the helpers encapsulate specific concerns.

## `internal/ui`

Contains the UI renderer and static assets for the directory listing page.

- `renderer.go` builds a template model and executes the page template.
- The template is responsible for markup structure only.
- `static.go` exposes `StaticHandler()` to serve embedded `/static/*` assets.
- CSS and JS live in `internal/ui/css` and `internal/ui/js`.

### Why this separation matters

- The Go template layer is separate from request handling.
- Static assets are embedded and served by a dedicated handler.
- Presentation concerns (CSS/JS) do not leak into the core HTTP logic.
- This makes it easier to replace the UI later with a different renderer or frontend.

## Frontend assets

- `internal/ui/css/dirlist.css` contains minimal styling.
- `internal/ui/js/dirlist.js` inspects file extensions on the client side and applies category classes such as `type-pdf`, `type-image`, and `type-archive`.
- This allows the UI to annotate files with icons and styles without changing server-side logic.

## Deployment note

The server is designed to run behind a reverse proxy like nginx for TLS termination.
- The app binds to `PORT` / `BIND_ADDR`.
- The proxy can terminate TLS and forward traffic to the Go app on plain HTTP.
