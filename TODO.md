# Runtime Improvements

## Correctness

- [ ] **Unguarded type assertion in `GetTypedParamsFromContext`** — [url/utils.go:36](runtime/url/utils.go#L36)
  `ctx.Value(ParamKey{}).(map[string]string)` panics if the context holds a wrong type. Add a two-value assertion with an `!ok` guard.

- [ ] **Silent parameter conversion failure** — [url/utils.go:103](runtime/url/utils.go#L103)
  When a query/path param can't be converted to the target type (e.g. `"abc"` → `int`), the field is silently set to its zero value. Should at minimum log a warning.

- [ ] **Inconsistent JSON encoding error handling** — [router/router.go:96](runtime/router/router.go#L96) vs [handle/handle.go:63](runtime/handle/handle.go#L63)
  Middleware response encoding errors are only logged; handler response encoding errors are returned to the client. Pick one behavior and apply it consistently.

## Design

- [ ] **Multiple query param values silently dropped** — [url/utils.go:49](runtime/url/utils.go#L49)
  `?tag=a&tag=b` keeps only `"a"`. Either support slice fields for multi-value params, or document the limitation clearly.

## Performance

- [ ] **Reflection on every request** — [url/utils.go:57-85](runtime/url/utils.go#L57)
  `setFieldsFromMap` reflects on struct fields per-request with no caching. Cache field metadata per type using `sync.Map`.

- [ ] **`responseWriter` uses byte slice append** — [router/router.go:80](runtime/router/router.go#L80)
  `rw.body = append(rw.body, b...)` causes repeated allocations for large responses. Replace with `bytes.Buffer`.
