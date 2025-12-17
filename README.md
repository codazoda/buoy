# Buoy

A static file server for web developers on MacOS.

- Minimal
- Starts automatically through launchd


## Getting Started

### Prerequisites

- macOS (uses launchd for auto-start)
- Go toolchain installed (`go` in PATH)

### Installation

Install the Buoy.

```
go install github.com/codazoda/buoy@latest
```

The installer places the _Buoy_ binary at `~/.local/bin/buoy`. Run it with the following command. It will offer to install it via LaunchD.

```
buoy
```

Point your web browser at [http://localhost:42869](http://localhost:42869/).

### Serve files from another directory via symlink

You can add another static site to Buoy by creating a symlink inside `~/.local/share/buoy/www`:

```sh
ln -s ~/sandbox/example ~/.local/share/buoy/www/example
```

Now requests to `http://localhost:42869/example` will serve files from `~/sandbox/example`.

### Uninstall

Remove the launchd service and installed binary:

```sh
buoy -uninstall
# or, from source:
go run buoy.go -uninstall
```

To also clear settings and content, delete `~/.local/share/buoy`.
