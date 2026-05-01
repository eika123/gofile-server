# Go Home File Server

A minimal Go file server using `chi`, nginx reverse proxy, and Podman for local/home deployment.

I wanted to make this because.
- I'm sick of the hyperscaler cowboys.
- I want to access my files while I rome the town from home to uni and back, and other places.
- It gives me the possibility to make exactly what I want, add the features I want, and so on.
- It's a project about something simple (serving files) suitable for a hobby project while 
  deploying it securely still provides learning opportunities. Extending features to e.g file sharing
  between several users and adding fault tolerance also provides some nice possibilities for learning new things.
- I'm not very motivated by complete toy projects. I would like to use this thing. It solves problems I actually have.


## How to run it
```sh
./scripts podman-up.sh
```

### Shut it down

```sh
./scripts podman-down.sh
```

## Prerequisites / dependencies

Currently, credentials for the https is just stored plaintext in `.env`.
Copy `.env-example` to `.env` before you start developing. This way of doing it is subject to change anytime soon.

For deployment, we use nginx as a reverse proxy. You can try running the go app locally without nginx using 
```sh
./scripts run-local.sh
```

We use podman and podman-compose for container management. Install using e.g 
```sh
sudo apt install podman podman-compose
```


We use OpenSSL to generate self-signed certificates.
Generate them using:

```sh
./scripts/generate-self-signed-cert.sh
```


Go dependencies are listed in `go.mod`. External libraries are

* Chi for routing and some middleware like logging, measuring request latencies etc.
* godotenv for reading env files.


## What it does

- Serves a directory listing through a Go web app
- Adds HTTP Basic Auth middleware
- Uses nginx as a reverse proxy for `HTTP` and `HTTPS`
- Serves static assets directly through nginx
- Supports a shared host volume for persistent files

## Current status

- `favicon.ico` is served through nginx
- Static assets are routed through `/static/`
- The app is currently configured for a Podman-based local deployment
- Basic auth credentials are currently stored in `.env`

## Quick start

1. Copy `.env.example` to `.env` and customize values.
2. Start the stack with `podman-compose up -d`.
3. Open the app in your browser on the exposed port.

## TODO

The current work plan is stored below.

### Immediate follow-up tasks

- [ ] Investigate and fix 504 gateway timeout on fresh server restart
  - Confirm whether nginx upstream DNS/Podman service discovery is causing stale upstream failures
  - Add retry/backoff or startup healthcheck handling so nginx waits for the app service

- [ ] Restore missing icon styling and underlines in the app UI
  - Verify `dirlist.js` and CSS assets are being loaded correctly through nginx
  - Confirm icon classes are present in rendered HTML and static CSS selectors

- [ ] Add hardening and security improvements
  - Harden Basic Auth and password handling
  - Add salted password storage logic and avoid plaintext storage, use hashes etc.
  - Consider env-based secret management and read-only file permissions

- [ ] Add tests
  - Add unit tests for request handlers and auth middleware
  - Add integration tests for static asset serving and favicon behavior
  - Add coverage for path validation and directory listing logic

### Home server deployment planning

- [ ] Plan home server deployment architecture
  - Choose host OS and container runtime (Podman is current candidate)
  - Define network routing and firewall rules for HTTP/HTTPS access
  - Decide on certificate management strategy (self-signed, ACME, internal CA)
  - Plan persistent volume mounts for shared data and logs

- [ ] Define deployment steps
  - Create a `deploy.sh` or provisioning script for the home server environment
  - Document required ports, volumes, and service names
  - Add rollback/restart instructions for the app and nginx stack

### Password security plan

- [ ] Design stronger password handling
  - Use bcrypt or a secure password hashing algorithm
  - Introduce a random salt per password and store only the hash
  - Do not store or log plaintext Basic Auth credentials
  - Add a mechanism for rotating or updating credentials safely
  - Migrate auth credentials from `.env` into e.g sqlite database

### Notes for next time

- The current favicon issue is resolved, static file serving is now working.
- Consider if the favicon logic should rather be provided by the Go application.
- Performance has improved after the latest nginx/static routing changes.
- The remaining gap is startup reliability and security hardening.
