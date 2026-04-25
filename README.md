# galley-fullstack-test

Three-service test app for Galley:

- **frontend** — Vite + React, served by nginx. Builds via the
  user-supplied `Dockerfile` so this exercises Galley's Kaniko path.
- **backend** — Go HTTP server with one endpoint that bumps a counter
  in Postgres. No `Dockerfile` — Galley's railpack autodetects Go and
  builds it via BuildKit.
- **db** — `postgres:16-alpine` image, no build.

`galley.yml` wires the dependencies (frontend → backend → db) and
declares the env vars each service needs.

Open the preview URL after Galley deploys; clicking *Bump counter*
round-trips through nginx → backend Go server → Postgres and back.
