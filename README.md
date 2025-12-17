# Buoy

A static file server for web developers on MacOS.

- Minimal
- Starts automatically through launchd
- Static file server

## Getting Started

```
go install joeldare.com/buoy@latest
```

The installer places the _Buoy_ binary at `~/.local/bin/buoy`.

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
