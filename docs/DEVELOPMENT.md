# Development

## Prerequisites

- Go 1.21+

## Local Run

```bash
go run ./cmd/app
```

## Tests

```bash
make test
```

Note: when running in constrained environments, set cache dirs:

```bash
XDG_CACHE_HOME=/work/pghapp/.cache GOCACHE=/work/pghapp/.gocache GOTMPDIR=/work/pghapp/.gotmp make test
```

## Webhook Fixtures

Test fixtures live in `test/fixtures/webhook/`. Add a fixture when you add a new
event handler.
