## 4go3

A 403 bypasser written in Go. The project now follows a conventional Go layout with the entry point under `cmd/4go3` and reusable packages living in `internal/`.

### Usage

```
go run ./cmd/4go3 -url https://example.org/test1/test2/test3?user=3 -segment test2
```

Key flags:

* `-url`/`-u` – Target URL. Schemes default to `https://` when omitted.
* `-segment`/`-path` – Named path segment to fuzz.
* `-segment-index` – Zero-based index of the segment to fuzz. Useful for duplicate segment names. When neither `-segment` nor `-segment-index` is supplied, the CLI lists the available path segments and lets you pick interactively.
* `-query`/`-q` – Additional query parameters in `key=value` form.
* `-threads`/`-t`, `-rate`, `-timeout` – Worker pool configuration.
* `-header`/`-H` – Repeatable custom headers.

Segments following the chosen target are preserved, so for `https://example.org/test1/test2/test3` you can fuzz `test1`, `test2`, or `test3` independently without losing the surrounding path. Query strings and fragments supplied in the URL are automatically carried into every generated request.
