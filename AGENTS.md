# AGENTS.md

## Project Overview

This repository contains `rcm`, a Go CLI for managing Redis Cluster and Redis master-slave deployments. It can inspect topology/status information and execute Redis commands against selected members. Some high-risk Redis commands are forbidden.

Primary documented commands:

- `rcm cluster status <seed-node>` displays Redis Cluster status grouped by shard.
- `rcm cluster status <seed-node> -s` also shows slot ranges.
- `rcm cluster exec <seed-node> -- <cmd> [args...]` executes a Redis command.
- `rcm cluster exec` can target specific nodes with `-n`, roles with `-r master|slave|all`, or the seed node by default.
- The Redis command is positional. There is no `-c` flag.

Global flags include:

- `-a, --password` for the Redis password.
- `-t, --timeout` for Redis operation timeout.
- `--cpupprof` and `--mempprof` for profiling output.

Forbidden Redis commands are `DEBUG`, `FLUSHALL`, `FLUSHDB`, `SHUTDOWN`, and `MONITOR`.

## Repository Layout

- `main.go`: CLI entrypoint.
- `cmd/`: Cobra root command, version command, and top-level command registration.
- `cmd/subcmd/cluster/`: Cluster subcommands such as `status`, `exec`, and `slowlog`.
- `cmd/subcmd/instance/`: Instance-level commands such as `monitor` and `keymap`.
- `redis/`: Redis client, connection, command, slot, and instance logic.
- `vars/`: Build-time metadata and shared runtime configuration.
- `perf/`: Profiling helpers.
- `README.md`: English user documentation.
- `README_zh.md`: Chinese user documentation.
- `Makefile`: Linux build, deploy, clean, and smoke-run targets.

## Build And Test

Use Go modules. The module path is `redis-cluster-manager`.

Common commands:

```sh
go test ./...
go build -o rcm main.go
```

The Makefile builds a Linux AMD64 static binary named `rcm` and injects version metadata:

```sh
make build
```

Be careful with:

```sh
make
```

The default `all` target runs `clean build deploy run`; `deploy` moves the binary to `/usr/local/bin/`, which may require elevated permissions and changes the local system.

## Development Notes

- Keep CLI behavior consistent with the README examples and Cobra help text.
- Prefer adding new user-facing commands under `cmd/subcmd/...` and registering them from the nearest existing command initializer.
- Keep Redis protocol and topology logic in the `redis/` package instead of embedding it in Cobra command handlers.
- Preserve the documented targeting behavior for `cluster exec`: explicit nodes, role-based targets, all nodes, or seed-only fallback.
- Preserve `cluster exec` command parsing: Redis command tokens come from positional args after the seed node, commonly separated from CLI flags with `--`.
- `-n` and `-r` are mutually exclusive for `cluster exec`; maintain this contract when changing target selection.
- Cluster status output is expected to be grouped by shard, ordered by master address, and include keys, client, slot count, and slot range information.
- Do not assume a live Redis Cluster is available in automated tests. Prefer unit tests around parsing, command construction, target selection, and formatting.

## Before Finishing Changes

Run:

```sh
go test ./...
```

For CLI changes, also run a local build:

```sh
go build -o rcm main.go
```

If changing Makefile behavior, check whether the target writes outside the repository before running it.

## Future Work Mentioned In README

README tracks these planned or partially implemented areas:

- Cluster slowlog collection and unified display.
- Instance monitor parsing and command distribution display.
- Instance keymap histogram support with configurable bucket boundaries and sampling rate.

When working on these areas, update README examples and help text together with code changes.
