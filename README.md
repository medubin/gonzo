# Gonzo

Gonzo is a code generation library and CLI tool that creates type-safe web API servers and clients from a custom API definition language. Define your API once and generate consistent, type-safe code across multiple languages.

## What Gonzo Does

Gonzo takes API definitions written in a custom DSL and generates:

- **Go server code** with type-safe handlers, routing, and parameter extraction
- **TypeScript client code** for frontend applications with full type safety
- **Type definitions** that ensure consistency between client and server
- **HTTP routing setup** with automatic path and query parameter handling
- **Request/response marshaling** with proper error handling

## How It Works

1. **Define your API** using Gonzo's `.api` format:

   ```api
   type UserID int64
   type Email string

   enum UserRole string {
     ADMIN = "admin"
     USER = "user"
   }

   type User {
     required ID UserID
     Username string
     Email Email
     Role UserRole
   }

   type CreateUserRequest {
     required Username string
     required Email Email
     Role UserRole
   }

   type UserListParams {
     page int32
     pageSize int32
   }

   server UserService {
     GetUser GET /users/{id UserID} returns(User)
     CreateUser POST /users body(CreateUserRequest) returns(User)
     ListUsers GET /users parameters(UserListParams) returns(repeated(User))
   }
   ```

2. **Generate code** using the Gonzo CLI:

   ```bash
   go run bin/gonzo-api.go generate -input server.api -language go -stack server -output ./server -package server
   ```

3. **Implement your business logic** by satisfying the generated interface:

   ```go
   func (s *UserServiceImpl) GetUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, GetUserUrl]) (*User, error) {
     userID := *url.PathParams.Id
     // Your business logic here
     return &User{ID: &userID, Username: &username}, nil
   }
   ```

## Project Structure

```
bin/                          # CLI tool
code_generator/               # Parser, generator, and templates
runtime/                      # Runtime libraries imported by generated code
  ├── cookies/                # Cookie handling
  ├── gerrors/                # Structured error handling with HTTP status codes
  ├── handle/                 # Generic type-safe request handler
  ├── middleware/             # Middleware system
  ├── router/                 # HTTP routing
  ├── types/                  # Shared types
  └── url/                    # URL parameter parsing
api-definition-language/      # VSCode syntax highlighting for .api files
```

## Language Specification

For detailed information about the API definition language syntax, see [API_SPEC.md](./API_SPEC.md).

## CLI Usage

### Basic Syntax

```bash
go run bin/gonzo-api.go generate [flags]
```

### Flags

- `-input <path>`: Path to the `.api` file. Required.
- `-output <directory>`: Output directory for generated files. Required.
- `-language <lang>`: Target language (`go` or `typescript`). Required.
- `-stack <type>`: Generate for `server` or `client`. Required.
- `-package <name>`: Package name for generated code. Required.

### Examples

**Generate Go server code:**
```bash
go run bin/gonzo-api.go generate -input server.api -language go -stack server -output ./server -package server
```

**Generate TypeScript client code:**
```bash
go run bin/gonzo-api.go generate -input server.api -language typescript -stack client -output ./client -package client
```

### What Gets Generated

**For Go servers (`-stack server`):**
- `types.go` - Type definitions, enums, and URL parameter structs
- `server.go` - Server interface and router setup
- `server_impl.go` - Implementation struct (generated once, not overwritten)
- Individual endpoint files (e.g., `get_user.go`) - Implementation stubs

**For TypeScript clients (`-stack client`):**
- `types.ts` - Type definitions, interfaces, and enums
- `client.ts` - Client class with methods for each endpoint

## Using as a Library

Add Gonzo as a dependency in your app:

```bash
go get github.com/medubin/gonzo
```

Generated server code imports from the runtime packages:

```go
import "github.com/medubin/gonzo/runtime/gerrors"
import "github.com/medubin/gonzo/runtime/router"
import "github.com/medubin/gonzo/runtime/middleware"
```

For local development of both Gonzo and your app simultaneously, use Go workspaces:

```bash
go work init
go work use ../gonzo
go work use .
```

# TODO

## api

- [x] error to http code handling
- [x] export to typescript
- [x] remove line removal
- [x] add generated to files
- [x] prevent duplicate types
- [x] use templates
- [x] handle nested maps and arrays
- [x] add required type with check
- [x] upcase first letter for the getter function
- [x] flip the typing so that golang complains about the incorrect endpoint function instead of complaining in main
- [x] simplify server parameters (now uses URL[params, pathParams])

## Language & API Definition

- [ ] **Nested routes** — Allow grouping endpoints under a common path prefix (e.g., `/users/{id}/...`) so related endpoints share path parameters without repeating them on every definition.

- [ ] **Import/module system** — Enable splitting large API definitions across multiple `.api` files and importing shared types. Right now everything must live in a single monolithic file, which becomes unwieldy for large APIs.

- [ ] **Options/decorators syntax** — Add metadata annotations to endpoints and types (e.g., caching directives, rate limiting hints, deprecation markers). Currently there is no way to attach structured configuration to definitions.

- [ ] **Deprecation markers** — Allow marking endpoints and types as deprecated in the API definition, so generators can emit deprecation warnings in generated code.

- [ ] **API versioning** — Add first-class versioning support to the API definition language so that multiple versions of an API can be defined and generated from a single source.

- [ ] **Request/response header definitions** — Add syntax for specifying required or custom request/response headers as part of an endpoint definition. Currently headers can only be handled through middleware with no type-safe contract.

- [ ] **HEAD and OPTIONS HTTP methods** — The parser only recognizes GET, POST, PUT, DELETE, and PATCH. HEAD and OPTIONS are standard HTTP methods with practical uses (health checks, CORS preflight) and should be supported.

- [ ] **Custom HTTP success status codes** — Endpoints currently always return 200. There should be a way to declare a different success code (e.g., 201 for resource creation, 204 for deletion) in the API definition, which the generator then uses in both server and client output.

- [ ] **Type and field validation** — Add constraint syntax for fields (e.g., min/max length for strings, numeric ranges, regex patterns). The generator would emit validation logic in the target language rather than requiring manual validation in every handler.

- [ ] **Typed cookies** — The runtime `cookies` package exists but there is no way to declare typed cookies in the `.api` language. Add syntax so cookie shapes are part of the API contract and generated code is type-safe.

- [ ] **Map key type validation** — The spec documents that non-comparable map key types produce invalid Go code, but the generator does not catch this. Add a validation pass that rejects non-comparable key types at generation time rather than at Go compile time.

## Code Generation

- [ ] **File splitting / modularization** — Currently all types land in a single `types.go`/`types.ts` file and all client methods in a single `client.ts`. For large APIs this becomes hard to navigate. Add an option to split output across multiple files organized by resource or service group.

- [ ] **TypeScript structured error types** — The Go server uses the `gerrors` package to return typed errors with distinct HTTP status codes (400/401/404/409/500). The TypeScript client throws a generic `Error` with only the status code as a string, so callers cannot distinguish error types. Generate typed error classes that mirror the `gerrors` hierarchy.

- [ ] **TypeScript enum helpers** — Go generates `EnumFromType`, `EnumToType`, and `EnumIsValid` conversion functions for every enum. TypeScript only generates a string-union type with no validation or conversion utilities. Generate equivalent helper functions in TypeScript.

- [ ] **TypeScript request validation** — Go generates a `Validate()` method on request types that checks required fields. TypeScript generates no equivalent, so invalid requests are only caught server-side. Generate TypeScript validation helpers to enable client-side validation before a request is sent.

- [ ] **OpenAPI/Swagger export** — Generate an OpenAPI 3.x spec from the `.api` definition. This would enable integration with API explorers (Swagger UI, Redoc), auto-generated docs, and third-party tooling that consumes OpenAPI specs.

- [ ] **Mock implementation generation** — Generate mock server implementations and stub client instances for use in tests. Currently every consumer must hand-roll their own mocks.

## Runtime & Middleware

- [ ] **Streaming endpoints** — Add support for endpoints that stream a sequence of values rather than returning a single response. This likely requires both new API definition syntax and new runtime plumbing for server-sent events or chunked transfer encoding.

- [ ] **WebSocket endpoints** — Add support for declaring WebSocket endpoints in the API definition and generating the corresponding connection-handling scaffolding on server and client.

- [ ] **Validation decorators / middleware** — Add built-in middleware for enforcing schema-level constraints (field lengths, value ranges, enum membership) automatically, without requiring manual validation in each handler.
