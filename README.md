# Buoy

An always-on static file server for web developers on macOS. See https://buoy.joelryan.com.


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

Buoy writes a default `index.html` into the `www` directory when it doesn't exist. If the existing `index.html` still includes the `<!-- Buoy default index -->` marker comment, Buoy will replace it on startup. Remove that comment (or replace the file) to keep a custom version.

### Serve files from another directory via symlink

You can add another static site to Buoy by creating a symlink inside `~/.local/share/buoy/www`:

```sh
ln -s ~/sandbox/example ~/.local/share/buoy/www/example
```

Now requests to `http://localhost:42869/example` will serve files from `~/sandbox/example`.


## Acknowledgements

Created by Joel Dare

Copyright (c) Dare Companies Dotcom LLC
