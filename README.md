# Buoy

A static file server for developers on MacOS.

- Minimal
- Starts automatically through launchd
- Static file server

## Getting Started

### Serve files from another directory via symlink
You can point Buoy’s `www` folder at another location by creating a symlink inside `~/.local/share/buoy/www`:

```sh
ln -s ~/sandbox/example ~/.local/share/buoy/www/example
```

Now requests to `http://localhost:42869/example` will serve files from `~/sandbox/example`.
