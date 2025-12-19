# Buoy

A static file server for web developers on macOS. Open source and distributed under the MIT license.

- Minimal


## Getting Started

### Prerequisites

- Go toolchain installed (`go` in PATH)

### Installation

Install Buoy from source:

```
go install github.com/codazoda/buoy@latest
```

Run it with:

```
buoy
```

Point your web browser at [http://localhost:42869](http://localhost:42869/). Content is served from `~/.local/share/buoy/www` and the port can be overridden via the `PORT` environment variable.

### Serve files from another directory via symlink

You can add another static site to Buoy by creating a symlink inside `~/.local/share/buoy/www`:

```sh
ln -s ~/sandbox/example ~/.local/share/buoy/www/example
```

Now requests to `http://localhost:42869/example` will serve files from `~/sandbox/example`.


## Acknowledgements

Created by Joel Dare

Copyright (c) Dare Companies Dotcom LLC
