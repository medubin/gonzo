# Gonzo

Gonzo is a modern code generation tool that creates type-safe web API servers and clients from a custom API definition language. It enables developers to define their complete API structure using a simple, declarative syntax and automatically generates consistent, type-safe code across multiple languages.

## What Gonzo Does

Gonzo takes API definitions written in a custom Domain Specific Language (DSL) and generates:

- **Go server code** with type-safe handlers, routing, and parameter extraction
- **TypeScript client code** for frontend applications with full type safety
- **Type definitions** that ensure consistency between client and server
- **HTTP routing setup** with automatic path and query parameter handling
- **Request/response marshaling** with proper error handling

## How It Works

1. **Define your API** using Gonzo's `.api` format:

   ```api
   // Type aliases and primitives
   type UserID int64
   type Email string

   // Enums with underlying types
   enum UserRole string {
     ADMIN = "admin"
     USER = "user"
   }

   // Structured types
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

   // Query parameter types
   type UserListParams {
     page int32
     pageSize int32
     sortBy string
   }

   // Server definition with multiple endpoint types
   server UserService {
     // Path parameters and response types
     GetUser GET /users/{id UserID} returns(User)
     
     // Request bodies
     CreateUser POST /users body(CreateUserRequest) returns(User)
     
     // Query parameters
     ListUsers GET /users parameters(UserListParams) returns(repeated(User))
     
     // Combined path and query parameters
     GetUsersByRole GET /users/role/{role UserRole} parameters(UserListParams) returns(repeated(User))
   }
   ```

2. **Generate code** using the Gonzo CLI:

   ```bash
   go run api/bin/gonzo-api.go generate -input server -language go -stack server -output server -package server
   ```

3. **Implement your business logic** by satisfying the generated interface:
   ```go
   func (s *UserServiceImpl) GetUser(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[struct{}, GetUserUrl]) (*User, error) {
     userID := *url.PathParams.Id
     // Your business logic here
     return &User{ID: &userID, Username: &username}, nil
   }

   func (s *UserServiceImpl) ListUsers(ctx context.Context, body *struct{}, cookie cookies.Cookies, url url.URL[UserListParams, struct{}]) (*[]User, error) {
     page := *url.Params.Page
     pageSize := *url.Params.PageSize
     // Your pagination logic here
   }
   ```

## Features

- **Type Safety**: Generated code ensures type consistency between API definitions and implementations
- **Multiple Language Support**: Generate Go servers and TypeScript clients from the same definition
- **Custom DSL**: Simple, readable syntax for defining APIs
- **Built-in HTTP Handling**: Automatic routing, request/response marshaling, and error handling
- **Database Integration**: Works with SQL databases using tools like sqlc
- **Authentication Support**: Built-in patterns for session management and user authentication

## Project Structure

- `api/code_generator/generator/` - Core code generation engine (lexer, parser, generators)  
- `api/code_generator/generator/languages/` - Language-specific templates (Go, TypeScript)
- `api/src/` - Runtime libraries for generated servers (routing, error handling, URL parsing)
- `api/bin/` - CLI tool for code generation
- `server/` - Generated server code and manual implementations
- `db/` - Database migrations and queries (using sqlc)
- `internal/` - Main application entry point

This project includes a complete example implementation of a user authentication API with signup, signin, signout, and user management endpoints.

## Language Specification

For detailed information about the API definition language syntax, see [API_SPEC.md](./API_SPEC.md).

## CLI Usage

The Gonzo CLI provides a `generate` command to create code from your API definitions:

### Basic Syntax

```bash
gonzo-api generate [flags]
```

### Available Flags

- `-input <name>`: Input file name (without .api extension). Required.
- `-output <directory>`: Output directory for generated files. Required.
- `-language <lang>`: Target language (`go` or `typescript`). Required.
- `-stack <type>`: Generate for `server` or `client`. Required.
- `-package <name>`: Package name for generated code. Required.

### Examples

**Generate Go server code:**

```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -language go -stack server -output api/code_generator/generator/test_data/server -package server
```

**Generate TypeScript client code:**

```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -language typescript -stack client -output api/code_generator/generator/test_data/client -package client
```

**Generate from your own API file:**

```bash
go run api/bin/gonzo-api.go generate -input server -language go -stack server -output ./generated -package myapi
```

### What Gets Generated

When you run the generate command, Gonzo creates:

**For Go servers (`-stack server`):**

- `types.go` - Type definitions, enums, and URL parameter structs
- `server.go` - Server interface and router setup with type-safe handlers
- `server_impl.go` - Implementation struct (generated once, not overwritten)
- Individual endpoint files (e.g., `get_user.go`, `create_user.go`) - Implementation stubs with proper signatures

**For TypeScript clients (`-stack client`):**

- `types.ts` - Type definitions, interfaces, and enums
- `client.ts` - Client classes with methods for each endpoint
- Full type safety with request/response types

### Example Workflow

1. Create your API definition file (e.g., `server.api`)
2. Run the generator:
   ```bash
   go run api/bin/gonzo-api.go generate -input server -language go -stack server -output ./server -package server
   ```
3. Implement the generated interface methods in your `*ServerImpl` struct
4. Start your server using the generated router:
   ```go
   func main() {
       impl := &MyServerImpl{}
       router := &router.Router{}
       StartMyServer(impl, router)
       http.ListenAndServe(":8080", router)
   }
   ```

# TODO

## service

- [ ] finish auth
- [ ] dockerize
- [ ] cdk

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
- [ ] add imports
- [ ] middleware
- [ ] validate types and fields
- [ ] type cookies?

## New Features

- [ ] OpenAPI/Swagger export
- [ ] Generate mock implementations
- [ ] Add validation decorators
- [ ] Support for streaming endpoints
- [ ] Add rate limiting templates
- [ ] Support for WebSocket endpoints

## frontend

- [ ] react
