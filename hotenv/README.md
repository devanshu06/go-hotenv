# hotenv

`hotenv` is a lightweight Go package for reading environment-style configuration files that **auto-reload** when the file changes.  
It’s designed for use in containerized environments like Kubernetes, where secrets/configs are mounted as files and updated automatically (e.g., via Doppler, ESO, or AWS Secrets Manager).

---

## Features

- Reads `.env`-style files (`KEY=VALUE` or multi-line quoted values)
- Hot-reloads automatically when the file changes
- API similar to Go’s built-in `os.Getenv`
- Works seamlessly with mounted Kubernetes Secret or ConfigMap volumes
- Thread-safe and efficient (uses `fsnotify`)

---

## Installation

```bash
go get github.com/devanshu06/go-hotenv/hotenv@latest
```
---

## Usage

### Basic example

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/devanshu06/go-hotenv/hotenv"
)

func main() {
	// Optional: start the watcher explicitly (otherwise it starts lazily)
	hotenv.Init("")

	http.HandleFunc("/hi", func(w http.ResponseWriter, r *http.Request) {
		message := hotenv.Getenv("GREETING_TEXT", "Hello")
		fmt.Fprintln(w, message)
	})

	port := hotenv.Getenv("PORT", "8080")
	http.ListenAndServe(":"+port, nil)
}
```
---
### File format

A `.env` file mounted at `/app/secrets/.env` might look like this:

```
GREETING_TEXT="hello from devanshu06"
```

`hotenv` supports:
- Single-line key/value pairs  
- Multi-line values wrapped in `'` or `"` quotes  
- Comments starting with `#`

---

### How hot reload works

- `hotenv` uses [`fsnotify`](https://github.com/fsnotify/fsnotify) to **watch the directory** of your `.env` file (default `/app/secrets/.env`).
- When the file or directory emits a change event (`Write`, `Create`, `Rename`, etc.), the watcher waits **800 ms** (a *debounce*) and reloads the file once.
- This covers the Kubernetes Secret update pattern (atomic symlink swap).
- The reload updates an in-memory map. All subsequent `hotenv.Getenv()` calls instantly return the new values — no restart needed.

---

### Configuration

You can tweak defaults **before** the first `Getenv` call:

```go
hotenv.WithDefaultPath("/custom/path/.env")
hotenv.WithFallbackToProcessEnv(false) // disables os.Getenv fallback
hotenv.WithLogger(func(f string, v ...any) { fmt.Printf(f, v...) })
hotenv.Init("") // start watcher early
```

---