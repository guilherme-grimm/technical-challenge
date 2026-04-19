# technical-challenge

Small Devices REST API in Go with mongoDB.

Handles devices (id, name, brand, state, created_at) plus a version field for optimistic locking. You can create, get, fully or partially update, list with brand/state filters and cursor pagination, and delete.

Few rules the service enforces:

- created_at is set server side, clients cant touch it
- name and brand cant be changed while state is `in-use` (409)
- `in-use` devices cant be deleted (409)
- version mismatch on update/patch returns 409

## Before you start

You need docker and docker-compose.
Populate the .env file with your mongo connection string.

    cp .env.example .env

Make sure you have a mongo instance running. Be it via docker compose, be it locally.

openapi stubs are generated automatically. `make build`, `make test` and `make run` all depend on `make generate`, which nukes `internal/api/openapi/` and rebuilds it from `internal/api/openapi.yaml`. the Dockerfile does the same on build, so `docker compose up --build` works on a fresh clone.

## Running

Easiest way:

    docker compose up --build

mongo and the api come up together, api waits on mongos healthcheck. Once its up:

- http://localhost:8080 for the api itself
- /docs for the swagger ui
- /openapi.yaml for the raw spec
- /healthz for the health probe

To nuke everything including mongodb volume:

    docker compose down -v

If you'd rather run the api outside docker just bring up mongo on its own:

    docker compose up -d mongo
    cp .env.example .env
    make run

Defaults point at localhost:27017, db `devices`, collection `devices`, listens on :8080.

## Config

Only `MONGO_URI` is actually required, the rest fall back to sensible defaults.

- MONGO_DATABASE: devices
- MONGO_COLLECTION: devices
- ADDR: :8080
- LOG_LEVEL: info
- READ_TIMEOUT, WRITE_TIMEOUT: 15s
- IDLE_TIMEOUT: 60s
- SHUTDOWN_TIMEOUT: 15s

Durations are go duration strings so `500ms`, `2m` etc all work.

## Tests

`make test` runs everything. Heads up the mongo tests spin up a real mongo via testcontainers so docker has to be running or that package wont even boot. If you want just the fast test:

    go test ./internal/service/... ./internal/api/...

## Misc

ids are ULIDs, server assigned. pagination is cursor based, grab the `next_cursor` from a page and pass it as `cursor` on the next call, empty next_cursor means you hit the end.

openapi server stubs under `internal/api/openapi/` are generated and gitignored, `make generate` rebuilds them from `internal/api/openapi.yaml`.

cors is wide open right now, fine for this but obviously wouldnt ship it like that.

## Following up

There are things that could be improved, such as:
- instrumentation w victoriametrics, combined with the tracing logs
- properly add different databases to check the performance-
- stress testing
- better error handling w/ a proper error type and custom metrics logic


