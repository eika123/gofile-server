# Podman deployment

This project can be deployed with Podman behind an Nginx reverse proxy.

## Files added

- `Containerfile` — builds the Go application into a minimal container image.
- `podman-compose.yml` — defines the application and reverse proxy services.
- `nginx/nginx.conf` — routes incoming HTTP traffic to the Go app.
- `.containerignore` — excludes local artifacts and source-only files from the build context.

The compose file uses a fully qualified image name for the app service (`localhost/go-home-file-server_file-server`) so Podman does not need unqualified name lookup or registry aliases.

## Service layout

- `file-server` runs the Go app and exposes port `8080` internally.
- `reverse-proxy` runs Nginx and exposes port `80` to the host.
- `shared_data` is a Podman-managed persistent volume mounted into the app at `/app/shared`.

## Environment variables

The compose file sets:

- `ROOT_PATH=/app/shared`
- `BASIC_AUTH_USER`
- `BASIC_AUTH_PASS`

Update `BASIC_AUTH_USER` and `BASIC_AUTH_PASS` before deploying.

## Running

From the repo root, run the helper script to load `.env` and start the stack:

```sh
scripts/podman-up.sh
```

If you are running rootless Podman and cannot bind port `80`, start in development mode:

```sh
scripts/podman-up.sh --dev
```

This will use the default unprivileged ports:

- `PROXY_PORT=8080` for HTTP reverse-proxy access
- `PROXY_SSL_PORT=8443` for HTTPS reverse-proxy access
- `APP_PORT=8081` for direct app access if needed

To override the defaults, export values in your shell or `.env` file:

```sh
PROXY_PORT=8085 PROXY_SSL_PORT=8444 APP_PORT=8086 scripts/podman-up.sh --dev
```

## Running locally without Podman

You can run the app directly for simple development checks without starting the full Nginx stack:

```sh
scripts/run_local.sh
```

This will load values from `.env`, use the current repository as the root path, and write logs to `./logs/file-server/app.log` by default.

If you want to use a different root path or bind address locally:

```sh
ROOT_PATH=./shared BIND_ADDR=":8081" BASIC_AUTH_USER=dev BASIC_AUTH_PASS=dev scripts/run_local.sh
```

To stop the server, press `Ctrl+C`.

To stop and remove containers:

```sh
scripts/podman-down.sh
```

## HTTPS support for Nginx

The reverse proxy can also serve HTTPS using a self-signed certificate.
Generate the certificate pair before starting the stack:

```sh
scripts/generate-self-signed-cert.sh
```

This creates `nginx/certs/selfsigned.crt` and `nginx/certs/selfsigned.key`.

If you are running the compose stack in development mode, access it at:

```sh
https://localhost:8443
```

For regular mode, use:

```sh
https://localhost
```

Your browser will warn about the self-signed certificate unless you trust it locally.

## Copying files into the shared volume

The `shared_data` volume is managed by Podman and is not automatically populated from the repo source tree.
Use the helper script to copy a host directory into the shared volume:

```sh
scripts/podman-copy-to-shared.sh ./shared
```

To replace the existing contents of the shared volume with the source directory, use:

```sh
scripts/podman-copy-to-shared.sh --clean ./shared
```

If you need to target a different volume name, set `VOLUME_NAME` before running the helper:

```sh
VOLUME_NAME=shared_data scripts/podman-copy-to-shared.sh ./shared
```

This is the neat way to populate the volume with any directory before starting or after recreating the stack.

## Viewing logs

Use the helper script to view logs from the Podman stack:

```sh
scripts/podman-logs.sh
```

Follow logs in real time:

```sh
scripts/podman-logs.sh --follow
```

View logs for a specific service:

```sh
scripts/podman-logs.sh file-server
scripts/podman-logs.sh reverse-proxy
```

### Persisted logs

The application now writes logs to a host-mounted directory under `./logs/file-server`, and Nginx writes logs to `./logs/nginx`.

This means logs are preserved even when the containers are stopped or recreated.

Nginx is configured to serve the shared volume mounted at `/app/shared`, and it falls back to the Go app only when a requested path is not present in the shared folder.

To read the persisted log files directly, use the file-backed helper:

```sh
scripts/podman-logs.sh --file
```

Or for a specific service:

```sh
scripts/podman-logs.sh --file file-server
scripts/podman-logs.sh --file reverse-proxy
```

## Notes

- The app is intended to be accessed through the reverse proxy on port `80`.
- The `shared_data` volume is persistent, so file uploads and contents remain after reboot.
- The application does not mount the source directory directly, keeping the source tree separate from persisted data.
