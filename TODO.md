# TODO

## Language & API Definition

- [ ] **Nested routes** — Allow grouping endpoints under a common path prefix (e.g., `/users/{id}/...`) so related endpoints share path parameters without repeating them on every definition.

- [ ] **Options/decorators syntax** — Add metadata annotations to endpoints and types (e.g., caching directives, rate limiting hints, deprecation markers). Currently there is no way to attach structured configuration to definitions.

- [ ] **Deprecation markers** — Allow marking endpoints and types as deprecated in the API definition, so generators can emit deprecation warnings in generated code.

- [ ] **API versioning** — Add first-class versioning support to the API definition language so that multiple versions of an API can be defined and generated from a single source.

- [ ] **Request/response header definitions** — Add syntax for specifying required or custom request/response headers as part of an endpoint definition. Currently headers can only be handled through middleware with no type-safe contract.

- [x] **HEAD and OPTIONS HTTP methods** — The parser only recognizes GET, POST, PUT, DELETE, and PATCH. HEAD and OPTIONS are standard HTTP methods with practical uses (health checks, CORS preflight) and should be supported.

- [ ] **Type and field validation** — Add constraint syntax for fields (e.g., min/max length for strings, numeric ranges, regex patterns). The generator would emit validation logic in the target language rather than requiring manual validation in every handler.

- [ ] **Typed cookies** — The runtime `cookies` package exists but there is no way to declare typed cookies in the `.api` language. Add syntax so cookie shapes are part of the API contract and generated code is type-safe.

- [x] **Map key type validation** — The spec documents that non-comparable map key types produce invalid Go code, but the generator does not catch this. Add a validation pass that rejects non-comparable key types at generation time rather than at Go compile time.

## Code Generation

- [ ] **File splitting / modularization** — Currently all types land in a single `types.go`/`types.ts` file and all client methods in a single `client.ts`. For large APIs this becomes hard to navigate. Add an option to split output across multiple files organized by resource or service group.

- [x] **TypeScript request validation** — Go generates a `Validate()` method on request types that checks required fields. TypeScript generates no equivalent, so invalid requests are only caught server-side. Generate TypeScript validation helpers to enable client-side validation before a request is sent.

- [x] **OpenAPI/Swagger export** — Generate an OpenAPI 3.x spec from the `.api` definition. This would enable integration with API explorers (Swagger UI, Redoc), auto-generated docs, and third-party tooling that consumes OpenAPI specs.

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
