# TODO

## Language & API Definition

- [x] **Nested routes** — Allow grouping endpoints under a common path prefix (e.g., `/users/{id}/...`) so related endpoints share path parameters without repeating them on every definition.

- [x] **Options/decorators syntax** — `@name(args, kw: value, ...)` decorators can be stacked above any endpoint or `group`. Group decorators cascade onto every endpoint inside (last-wins on conflicts). Open-vocabulary: parser collects them verbatim and the full set is emitted into `RouteInfo.Decorators` so middleware can dispatch on arbitrary names without a code-generator change. Typed shortcuts exist for the few first-class cases — `@auth` populates `RouteInfo.AuthScheme` and OpenAPI security; `@deprecated` emits Go/TS deprecation comments and OpenAPI `deprecated: true`. Decorators on types and fields are not yet supported (will land alongside the validation / OpenAPI examples TODOs).

- [x] **Deprecation markers** — `@deprecated` / `@deprecated("message")` decorator on endpoints emits a Go `// Deprecated:` doc comment, a TypeScript `/** @deprecated */` JSDoc, and `deprecated: true` on the OpenAPI operation. Type-level / field-level deprecation requires lifting decorator support to those nodes (tracked under "Options/decorators syntax").

- [x] **API versioning** — URL-based versioning is composed from `group` (path prefix) plus namespaced `import` (per-version types). See "API Versioning" in `API_SPEC.md`. Remaining gaps are tracked separately: per-version package layout (file splitting TODO), deprecation signaling (deprecation markers TODO), and header/content-negotiation versioning (header definitions TODO).

- [x] **Request/response header definitions** — Endpoint-level `@header(name, required: bool, description: string)` decorator. Documents request headers in OpenAPI and makes them reachable from server middleware via `RouteInfo.Decorators` for enforcement. Decided against typed handler-signature plumbing after triage — most headers (`X-Request-Id`, tenant IDs, ETags, idempotency keys) are middleware concerns, not handler concerns; type safety in handlers buys very little. Response headers and multi-value headers remain out of scope (will pair with the OpenAPI multi-status TODO).

- [x] **HEAD and OPTIONS HTTP methods** — The parser only recognizes GET, POST, PUT, DELETE, and PATCH. HEAD and OPTIONS are standard HTTP methods with practical uses (health checks, CORS preflight) and should be supported.

- [x] **Type and field validation** — Field-level `@validation(...)` decorator with `min`, `max`, `minLength`, `maxLength`, `pattern`, and `format` kwargs. Constraints are enforced at runtime by the generated Go `Validate()` method (returns `gerrors.InvalidArgumentError`) and the generated TypeScript `validate*` helpers, and surfaced in OpenAPI as `minimum`/`maximum`/`minLength`/`maxLength`/`pattern`/`format`. No parser-side check yet that a constraint matches its field's type (e.g., `min` on a string is ignored at runtime); treat as user discipline.

- [x] **Typed cookies** — Endpoint-level `@cookie(name, required, description, httpOnly, secure, sameSite, maxAge, path, domain)` decorator. Drives three things: (1) OpenAPI `parameters: { in: cookie }` entries; (2) a generated `cookies.go` in the consuming package with name constants + `SetXxx` helpers that bake declared security attributes into the call site; (3) middleware-readable contract via `RouteInfo.Decorators`. Decided against typing cookie *values* — real cookies are opaque session IDs or signed/encrypted blobs whose decode is app-specific (signing key, scheme, DB lookup) and which gonzo cannot generate. The ergonomic and security wins (typo-proof names + un-forgettable HttpOnly/Secure/SameSite) come from the constants + Set helpers without that pretense. Runtime gained `cookies.Opt` plus `MaxAge`, `Path`, `Domain` opts so call sites can extend a declared cookie without disabling its security flags.

- [x] **Map key type validation** — The spec documents that non-comparable map key types produce invalid Go code, but the generator does not catch this. Add a validation pass that rejects non-comparable key types at generation time rather than at Go compile time.

## Code Generation

- [ ] **File splitting / modularization** — Currently all types land in a single `types.go`/`types.ts` file and all client methods in a single `client.ts`. For large APIs this becomes hard to navigate. Add an option to split output across multiple files organized by resource or service group.

- [x] **TypeScript request validation** — Go generates a `Validate()` method on request types that checks required fields. TypeScript generates no equivalent, so invalid requests are only caught server-side. Generate TypeScript validation helpers to enable client-side validation before a request is sent.

- [x] **OpenAPI/Swagger export** — Generate an OpenAPI 3.x spec from the `.api` definition. This would enable integration with API explorers (Swagger UI, Redoc), auto-generated docs, and third-party tooling that consumes OpenAPI specs.

### OpenAPI follow-ups

- [x] **OpenAPI tags / server grouping** — Each `server` declaration emits a top-level `tags` entry, and every operation references its parent server via `tags: [<ServerName>]`. Swagger UI / Redoc now render a section per server.

- [x] **OpenAPI security schemes** — Driven by `@auth("<scheme>")` decorators. The OpenAPI generator collects every used scheme, emits per-operation `security:` requirements, and declares the schemes under `components.securitySchemes` with sensible defaults (`bearer` → http+JWT, `apiKey` → header `X-API-Key`). User-overridable scheme declarations (custom header names, OAuth flows, etc.) remain a future extension.

- [x] **OpenAPI multi-status responses** — Endpoint-level `@response(code, type?, description: ...)` decorator. Bodyless codes supported (omit the type arg for 204/304/etc.). Composes with `returns(...)`: declaring any explicit 2xx suppresses the implicit 200 default, so a 201-only create no longer carries a stale 200 entry. Description defaults to the standard HTTP reason phrase. Doc-only — handlers continue to set status freely on `*handle.Response[T]`. Decorator parser also gained a `reference`-kind arg so type names can appear bare (`@response(201, User)`) without quoting.

- [x] **OpenAPI examples** — Field-level `@example(value)` decorator emits `example:` on the field schema. Accepts string/number/bool literal. Skipped on `$ref` fields to keep the spec conservatively valid.

- [ ] **OpenAPI response headers** — Request-side is done: `@header(...)` decorators emit `parameters: { in: header }` entries on each operation. Response headers (per-status `headers:` blocks) remain — they are coupled to per-status response declarations and will land alongside the multi-status responses TODO.

- [x] **OpenAPI info block** — Added an `info { ... }` block to the `.api` language with title, version, description, contact, and license fields. The OpenAPI generator now populates the spec's `info` object from it; the CLI `-package` arg is the title fallback. Version/contact/license are content of the API and belong next to the routes — putting them in the language rather than CLI flags keeps any tool that reads the `.api` file alone fully informed.

- [ ] **Mock implementation generation** — Generate mock server implementations and stub client instances for use in tests. Currently every consumer must hand-roll their own mocks.

## Runtime & Middleware

- [ ] **Streaming endpoints** — Add support for endpoints that stream a sequence of values rather than returning a single response. This likely requires both new API definition syntax and new runtime plumbing for server-sent events or chunked transfer encoding.

- [ ] **WebSocket endpoints** — Add support for declaring WebSocket endpoints in the API definition and generating the corresponding connection-handling scaffolding on server and client.

- [x] **Validation decorators / middleware** — Subsumed by field-level `@validation(...)`: each request body type's generated `Validate()` method enforces field-length, value-range, and pattern constraints automatically. Handlers don't write any of the checks; the existing `RequireBody`-shaped path that already calls `Validate()` returns `gerrors.InvalidArgumentError` on a constraint failure. (Enum membership is already enforced via the generated `<Enum>IsValid` helpers.)

### Correctness

- [x] **Unguarded type assertion in `GetTypedParamsFromContext`** — [url/utils.go:36](runtime/url/utils.go#L36)
  `ctx.Value(ParamKey{}).(map[string]string)` panics if the context holds a wrong type. Add a two-value assertion with an `!ok` guard.

- [x] **Silent parameter conversion failure** — [url/utils.go:103](runtime/url/utils.go#L103) Conversion failures now log a warning (`gonzo: param "X": cannot convert "abc" to int32; field left at zero value`) instead of dropping the failure entirely. Field still ends up at zero value for backwards compatibility.
  When a query/path param can't be converted to the target type (e.g. `"abc"` → `int`), the field is silently set to its zero value. Should at minimum log a warning.

- [x] **Inconsistent JSON encoding error handling** — [router/router.go:96](runtime/router/router.go#L96) vs [handle/handle.go:63](runtime/handle/handle.go#L63)
  Middleware response encoding errors are only logged; handler response encoding errors are returned to the client. Pick one behavior and apply it consistently.

### Design

- [x] **Multiple query param values silently dropped** — [url/utils.go:49](runtime/url/utils.go#L49)
  `?tag=a&tag=b` keeps only `"a"`. Either support slice fields for multi-value params, or document the limitation clearly.
