# Gonzo

Gonzo is a code generation tool that creates web API servers from a custom API definition language. It allows developers to define their API structure using a simple, declarative syntax and automatically generates type-safe server code.

## What Gonzo Does

Gonzo takes API definitions written in a custom Domain Specific Language (DSL) and generates:

- **Go server code** with type-safe handlers and routing
- **TypeScript client code** for frontend applications
- **Type definitions** that ensure consistency between client and server
- **HTTP routing setup** with proper method and path handling

## How It Works

1. **Define your API** using Gonzo's `.api` format:

   ```
   type User {
     ID int32
     Name string
     Email string
   }

   type SignupBody {
     User User
     Password string
   }

   server MyServer {
     Signup POST /user/new body(SignupBody) returns(User)
     GetUser GET /user/<UserID> returns(User)
   }
   ```

2. **Generate code** using the Gonzo CLI:

   ```bash
   gonzo-api generate -input server -output ./generated -language go -stack server
   ```

3. **Implement your business logic** by satisfying the generated interface:
   ```go
   func (s *ServerImpl) Signup(ctx context.Context, body *SignupBody, cookie cookies.Cookies, url url.URL[SignupUrl]) (*User, error) {
     // Your implementation here
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

- `api/generate/` - Core code generation engine (lexer, parser, generators)
- `api/src/` - Runtime libraries for generated servers (routing, error handling, etc.)
- `server/` - Generated server code and manual implementations
- `db/` - Database migrations and queries
- `internal/` - Main application entry point

This project includes a complete example implementation of a user authentication API with signup, signin, signout, and user management endpoints.

## CLI Usage

The Gonzo CLI provides a `generate` command to create code from your API definitions:

### Basic Syntax

```bash
gonzo-api generate [flags]
```

### Available Flags

- `-input <name>`: Input file name (without .api extension). Required.
- `-output <directory>`: Output directory for generated files. Required.
- `-language <lang>`: Target language (`go` or `typescript`). Default: `go`
- `-stack <type>`: Generate for `server` or `client`. Default: `server`

### Examples

**Generate Go server code:**

```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -output api/code_generator/generator/test_data/server -language go -stack server -package server
```

**Generate TypeScript client code:**

```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -output api/code_generator/generator/test_data/client -language typescript -stack client -package client
```

**Generate from a different API file:**

```bash
gonzo-api generate -input my-api -output ./generated -language go -stack server
```

### What Gets Generated

When you run the generate command, Gonzo creates:

**For Go servers (`-stack server`):**

- `types.go` - Type definitions and interfaces
- `server.go` - Server implementation struct (if it doesn't exist)
- Individual endpoint files (e.g., `signup.go`, `get_user.go`)

**For TypeScript clients (`-stack client`):**

- `types.ts` - Type definitions
- Client code for making API calls

### Example Workflow

1. Create your API definition file (e.g., `server.api`)
2. Run the generator:
   ```bash
   gonzo-api generate -input server -output ./server -language go -stack server
   ```
3. Implement the generated interface methods in your server struct
4. Build and run your application

# TODO

## service

- [] finish auth
- [] dockerize
- [] cdk

## api

- [x] error to http code handling
- [] nested routes
- [] add required type with check
- [] export to typescript
- [x] remove line removal
- [x] add generated to files
- [] add file splitting
- [x] prevent duplicate types
- [] simplify server parameters
- [] add options
- [] add imports
- [] middleware
- [] validate types and fields
- [] type cookies?
- [] use templates
- [] flip the typing so that golang complains about the incorrect endpoint function instead of complaining in main
- [] upcase first letter for the getter function
- [] handle nested maps and arrays

## frontend

- [] react
