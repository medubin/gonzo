# TODO

## Language & API Definition

- [x] **Nested routes** — Allow grouping endpoints under a common path prefix (e.g., `/users/{id}/...`) so related endpoints share path parameters without repeating them on every definition.

- [x] **Options/decorators syntax** — `@name(args, kw: value, ...)` decorators can be stacked above any endpoint or `group`. Group decorators cascade onto every endpoint inside (last-wins on conflicts). Open-vocabulary: parser collects them verbatim and the full set is emitted into `RouteInfo.Decorators` so middleware can dispatch on arbitrary names without a code-generator change. Typed shortcuts exist for the few first-class cases — `@auth` populates `RouteInfo.AuthScheme` and OpenAPI security; `@deprecated` emits Go/TS deprecation comments and OpenAPI `deprecated: true`. Decorators on types and fields are not yet supported (will land alongside the validation / OpenAPI examples TODOs).

- [x] **Deprecation markers** — `@deprecated` / `@deprecated("message")` decorator on endpoints emits a Go `// Deprecated:` doc comment, a TypeScript `/** @deprecated */` JSDoc, and `deprecated: true` on the OpenAPI operation. Type-level / field-level deprecation requires lifting decorator support to those nodes (tracked under "Options/decorators syntax").

- [x] **API versioning** — URL-based versioning is composed from `group` (path prefix) plus namespaced `import` (per-version types). See "API Versioning" in `API_SPEC.md`. Remaining gaps are tracked separately: per-version package layout (file splitting TODO), deprecation signaling (deprecation markers TODO), and header/content-negotiation versioning (header definitions TODO).

- [ ] **Request/response header definitions** — Add syntax for specifying required or custom request/response headers as part of an endpoint definition. Currently headers can only be handled through middleware with no type-safe contract.

- [x] **HEAD and OPTIONS HTTP methods** — The parser only recognizes GET, POST, PUT, DELETE, and PATCH. HEAD and OPTIONS are standard HTTP methods with practical uses (health checks, CORS preflight) and should be supported.

- [ ] **Type and field validation** — Add constraint syntax for fields (e.g., min/max length for strings, numeric ranges, regex patterns). The generator would emit validation logic in the target language rather than requiring manual validation in every handler.

- [ ] **Typed cookies** — The runtime `cookies` package exists but there is no way to declare typed cookies in the `.api` language. Add syntax so cookie shapes are part of the API contract and generated code is type-safe.

- [x] **Map key type validation** — The spec documents that non-comparable map key types produce invalid Go code, but the generator does not catch this. Add a validation pass that rejects non-comparable key types at generation time rather than at Go compile time.

## Code Generation

- [ ] **File splitting / modularization** — Currently all types land in a single `types.go`/`types.ts` file and all client methods in a single `client.ts`. For large APIs this becomes hard to navigate. Add an option to split output across multiple files organized by resource or service group.

- [x] **TypeScript request validation** — Go generates a `Validate()` method on request types that checks required fields. TypeScript generates no equivalent, so invalid requests are only caught server-side. Generate TypeScript validation helpers to enable client-side validation before a request is sent.

- [x] **OpenAPI/Swagger export** — Generate an OpenAPI 3.x spec from the `.api` definition. This would enable integration with API explorers (Swagger UI, Redoc), auto-generated docs, and third-party tooling that consumes OpenAPI specs.

### OpenAPI follow-ups

- [ ] **OpenAPI tags / server grouping** — Today every endpoint lives in a flat `paths` block. Group endpoints by their parent `server` declaration via OpenAPI `tags` so Swagger UI / Redoc render them in sections.

- [x] **OpenAPI security schemes** — Driven by `@auth("<scheme>")` decorators. The OpenAPI generator collects every used scheme, emits per-operation `security:` requirements, and declares the schemes under `components.securitySchemes` with sensible defaults (`bearer` → http+JWT, `apiKey` → header `X-API-Key`). User-overridable scheme declarations (custom header names, OAuth flows, etc.) remain a future extension.

- [ ] **OpenAPI multi-status responses** — Every endpoint declares a single 200/204 plus a `default` error. Allow declaring additional response codes per endpoint (e.g. 201 Created, 404 Not Found) in the `.api` language and emit each.

- [ ] **OpenAPI examples** — `requestBody` / `responses` schemas have no `example` or `examples`. Add a way to attach example payloads to types or endpoints so the spec renders runnable samples in API explorers.

- [ ] **OpenAPI request/response headers** — Depends on the broader header-definition TODO. Once headers are first-class in the `.api` language, surface them as `parameters: { in: header }` and per-response `headers:` blocks in the spec.

- [ ] **OpenAPI info block** — `version` is hardcoded to `0.0.0` and `title` is the `-package` flag. Add `-api-version` plus optional `description`, `contact`, `license` flags so the generated `info` block can be populated without post-processing.

- [ ] **Mock implementation generation** — Generate mock server implementations and stub client instances for use in tests. Currently every consumer must hand-roll their own mocks.

## Runtime & Middleware

- [ ] **Streaming endpoints** — Add support for endpoints that stream a sequence of values rather than returning a single response. This likely requires both new API definition syntax and new runtime plumbing for server-sent events or chunked transfer encoding.

- [ ] **WebSocket endpoints** — Add support for declaring WebSocket endpoints in the API definition and generating the corresponding connection-handling scaffolding on server and client.

- [ ] **Validation decorators / middleware** — Add built-in middleware for enforcing schema-level constraints (field lengths, value ranges, enum membership) automatically, without requiring manual validation in each handler.

### Correctness

- [x] **Unguarded type assertion in `GetTypedParamsFromContext`** — [url/utils.go:36](runtime/url/utils.go#L36)
  `ctx.Value(ParamKey{}).(map[string]string)` panics if the context holds a wrong type. Add a two-value assertion with an `!ok` guard.

- [ ] **Silent parameter conversion failure** — [url/utils.go:103](runtime/url/utils.go#L103)
  When a query/path param can't be converted to the target type (e.g. `"abc"` → `int`), the field is silently set to its zero value. Should at minimum log a warning.

- [x] **Inconsistent JSON encoding error handling** — [router/router.go:96](runtime/router/router.go#L96) vs [handle/handle.go:63](runtime/handle/handle.go#L63)
  Middleware response encoding errors are only logged; handler response encoding errors are returned to the client. Pick one behavior and apply it consistently.

### Design

- [x] **Multiple query param values silently dropped** — [url/utils.go:49](runtime/url/utils.go#L49)
  `?tag=a&tag=b` keeps only `"a"`. Either support slice fields for multi-value params, or document the limitation clearly.
