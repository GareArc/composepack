# Runtime Writer

This document summarizes how `internal/core/runtime.Writer` materializes release directories, based on the runtime layout described in PRD.md / CLAUDE.md.

## Layout

Given a base directory (defaults to `.cpack-releases/` in config) and release name, the writer produces:

```
<cpackBase>/<release>/
  docker-compose.yaml    # merged Compose file
  files/                 # rendered file assets (scripts/configs, etc.)
    ...
```

Helper templates never appear here; only rendered/ static assets are copied into `files/`.

## Implementation Notes

* Uses `internal/util/fsutil` helpers for directory creation and atomic file writes.
* Existing `files/` dir gets wiped (via `os.RemoveAll`) before writing, ensuring stale files disappear.
* Paths from `WriteOptions.Files` must be relative; `Writer` rejects absolute paths or ones containing `..`.
* Files are written with `0644` permissions, compose file as well.
* Returns the full runtime path so callers can hand it to docker-compose commands.
