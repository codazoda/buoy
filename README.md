# Buoy

A static file server for developers on MacOS.

- Minimal
- Starts automatically through launchd
- Static file server

## Getting Started

### Install

One-liner (downloads a prebuilt macOS binary when available):

```sh
curl -fsSL https://raw.githubusercontent.com/codazoda/buoy/main/scripts/install.sh | bash
```

The installer places the `Buoy` binary in `~/.local/bin` and clears the macOS quarantine attribute so it runs without prompts.

### Serve files from another directory via symlink

You can add another static site to Buoy by creating a symlink inside `~/.local/share/buoy/www`:

```sh
ln -s ~/sandbox/example ~/.local/share/buoy/www/example
```

Now requests to `http://localhost:42869/example` will serve files from `~/sandbox/example`.
