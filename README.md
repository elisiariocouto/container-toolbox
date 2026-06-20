# container-toolbox

A small terminal UI for managing [Docker Compose](https://docs.docker.com/compose/)
stacks on a single server. Built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

Each direct subdirectory of `~/stacks/` that contains a compose file
(`compose.yaml`, `compose.yml`, `docker-compose.yaml`, or `docker-compose.yml`)
is treated as one stack. The directory name is the stack name.

## How it works

It is a hybrid client:

- The **Docker Go SDK** talks to the local socket to read live container status
  (running / stopped / partial), grouping containers by the
  `com.docker.compose.project` label.
- Lifecycle actions shell out to the **`docker compose` CLI**, run inside each
  stack directory so Compose auto-detects the file and project name.

## Features

- List stacks discovered in `~/stacks/` with their running/stopped status.
- Start (`up -d`), stop (`down`), and restart a stack.
- Pull newer images and recreate (`pull` then `up -d`).
- Stream a stack's logs.

## Requirements

- Go 1.25+ (to build)
- Docker Engine with the `docker compose` (v2) CLI plugin
- Access to the local Docker socket

## Build & run

```sh
make build        # produces ./bin/container-toolbox
make run          # go run .
./bin/container-toolbox --stacks-dir /path/to/stacks
```

By default it looks in `~/stacks`. Override with `--stacks-dir`.

## Keybindings

| Key            | Action                         |
| -------------- | ------------------------------ |
| `j` / `k`, ↑/↓ | move selection                 |
| `enter` / `l`  | view logs for the stack        |
| `u`            | start (`docker compose up -d`) |
| `d`            | stop (`docker compose down`)   |
| `r`            | restart                        |
| `p`            | pull + recreate                |
| `R`            | refresh status                 |
| `?`            | toggle help                    |
| `q` / `ctrl+c` | quit                           |

In the logs view: `esc` / `q` returns to the list (stopping the stream),
arrows / `ctrl+u` / `ctrl+d` scroll.

## License

MIT — see [LICENSE](LICENSE).
