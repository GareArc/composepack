# Release Metadata Store

`internal/core/release.Store` is responsible for persisting `release.json` inside a runtime directory. It records chart + values context so later commands (`up`, `down`, `logs`) can retrieve the history for a release.

## File Layout

```
.cpack-releases/<release>/
  docker-compose.yaml
  files/
  release.json        # managed by release.Store
```

## Metadata Fields

* `releaseName`: user-specified release id.
* `chartName` / `chartVersion`: from `Chart.yaml`.
* `chartDigest`: optional checksum of the packaged chart.
* `runtimePath`: absolute path to the runtime directory (set automatically when saving).
* `createdAt`: UTC timestamp (set when saving if zero).
* `values`: merged values map.
* `valuesSources`: list of value files / CLI overrides used to construct `.Values`.
* `composeFiles`: ordered list of compose fragment files merged together.

## Store Behavior

* `Load` returns `(*Metadata, nil)` when `release.json` exists, `nil, nil` when missing, and wraps other IO errors.
* `Save` ensures the runtime directory exists, sets `RuntimePath` / `CreatedAt`, and writes JSON using a temp file + rename for durability.
* Both methods honor `context.Context` cancellation prior to IO.
