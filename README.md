# hotenv

`hotenv` is a Go package for loading `.env` files with built-in hot reload, designed for containerized deployments where configuration files change at runtime. This repository contains the full source for the `github.com/devanshu06/go-hotenv/hotenv` module so that the community can track development, file issues, and contribute improvements.

---

## Why hotenv?

- Watches the backing file and refreshes values automatically (no restart).
- Mirrors the ergonomics of `os.Getenv` while adding sensible defaults and fallbacks.
- Plays nicely with Kubernetes Secrets, ConfigMaps, and other volume-mounted configs.
- Lightweight, production-tested, and safe for concurrent access.

---

## Install

```bash
go get github.com/devanshu06/go-hotenv/hotenv@latest
```

The module source lives in [`./hotenv`](./hotenv). See [`hotenv/README.md`](./hotenv/README.md) for API details and usage examples.

---

## Development workflow

1. Make changes under `hotenv/`.
2. Run `go test ./...` from inside the `hotenv/` directory (or at repo root) to keep the build green.
3. Format code with `go fmt ./...` before sending a pull request.

---

## Releasing a new version

1. Commit the changes in the `hotenv/` module directory.
2. Tag the release with the module path prefix and semantic version:
   ```bash
   git tag hotenv/v1.0.1
   git push origin hotenv/v1.0.1
   ```
3. Consumers can upgrade with:
   ```bash
   go get -u github.com/devanshu06/go-hotenv/hotenv@hotenv/v1.0.1
   ```
   Replace `v1.0.1` with the tag you just pushed.

---
