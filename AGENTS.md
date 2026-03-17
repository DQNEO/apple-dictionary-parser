# Go Agent Notes

- When running Go commands in this project (for example `go test ./...`), set `GOCACHE` to a writable directory such as `/tmp/go-build` to avoid sandbox permission errors:

  ```sh
  GOCACHE=/tmp/go-build go test ./...
  ```

- The default cache location under `~/Library/Caches/go-build` may be read-only in this environment, causing `operation not permitted` failures.
