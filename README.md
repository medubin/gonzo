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
- [ ] nested routes
- [ ] add file splitting
- [ ] add options
- [ ] validate types and fields
- [ ] type cookies

## New Features

- [ ] OpenAPI/Swagger export
- [ ] Generate mock implementations
- [ ] Add validation decorators
- [ ] Support for streaming endpoints
- [ ] Support for WebSocket endpoints
