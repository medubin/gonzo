# Gonzo - Claude AI Assistant Guide

This guide helps Claude AI effectively work with the Gonzo code generation project.

## About Gonzo

Gonzo is a standalone Go library and CLI tool that creates type-safe web API servers and clients from a custom API definition language. It consists of:

- **Core Parser/Generator**: Lexer, parser, and code generation engine
- **Language Templates**: Go server and TypeScript client templates
- **Runtime Libraries**: HTTP routing, middleware, error handling, URL parsing
- **CLI Tool**: Command-line interface for code generation

Gonzo is designed to be imported as a library by app repos. The runtime packages (`runtime/`) are imported by generated server code. The CLI (`bin/gonzo-api.go`) is run from the consuming app to generate code.

## Common Bash Commands

**Generate Go server code from API definition:**
```bash
go run bin/gonzo-api.go generate -input code_generator/generator/test_data/test_server.api -language go -stack server -output code_generator/generator/test_data/server -package server
```

**Generate TypeScript client code:**
```bash
go run bin/gonzo-api.go generate -input code_generator/generator/test_data/test_server.api -language typescript -stack client -output code_generator/generator/test_data/client -package client
```

**Run the example server:**
```bash
go run code_generator/generator/test_data/main.go
```

**Run tests:**
```bash
go test ./...
go test -v code_generator/generator/...
```

**Update snapshots after intentional changes:**
```bash
UPDATE_SNAPS=true go test -v -run TestCoreGenerate ./code_generator/generator/...
```

## Code Style Guidelines

- **Go**: Follow standard Go conventions, use `gofmt`
- **Template syntax**: Use Go template syntax in `.yaml` config files
- **File structure**: Generated files should not be manually edited
- **Error handling**: Use the `gerrors` package for structured error responses (see Error Handling section below)
- **Imports**: Only import packages that are actually used to avoid unused import errors
- **Field types**: All struct fields should be pointers with `omitempty` JSON tags
- **Validation**: Required fields validated with simple `== nil` checks

## Error Handling

**Always use the `gerrors` package instead of generic Go errors for consistent HTTP status codes and structured error responses.**

### Available Error Types

| Function | HTTP Status | Use Case |
|----------|-------------|----------|
| `gerrors.MissingArgumentError()` | 400 Bad Request | Missing required fields, parameters, or request body |
| `gerrors.InvalidArgumentError()` | 400 Bad Request | Invalid input format or values |
| `gerrors.UnauthenticatedError()` | 401 Unauthorized | Invalid credentials or missing authentication |
| `gerrors.PermissionDeniedError()` | 403 Forbidden | Authenticated but not authorized to perform the action |
| `gerrors.NotFoundError()` | 404 Not Found | Resource doesn't exist |
| `gerrors.AlreadyExistsError()` | 409 Conflict | Resource already exists |
| `gerrors.InternalError()` | 500 Internal Server Error | Database errors, system failures |
| `gerrors.UnimplementedError()` | 501 Not Implemented | Placeholder for unfinished endpoints |
| `gerrors.RateLimitedError()` | 429 Too Many Requests | Rate limit or quota exceeded |
| `gerrors.UnavailableError()` | 503 Service Unavailable | Downstream dependency temporarily unavailable |

### Error Handling Examples

**✅ Correct - Use gerrors:**
```go
// Missing required parameter
if userID == nil {
    return nil, gerrors.MissingArgumentError("missing user ID")
}

// Resource not found
if err == sql.ErrNoRows {
    return nil, gerrors.NotFoundError("user not found")
}

// Invalid credentials
if !auth.CheckPassword(password, hash) {
    return nil, gerrors.UnauthenticatedError("invalid email or password")
}

// Database error
if err != nil {
    return nil, gerrors.InternalError("database error")
}
```

**❌ Incorrect - Don't use generic errors:**
```go
// Don't do this - no HTTP status code info
return nil, errors.New("user not found")
return nil, fmt.Errorf("missing user ID")
```

### Generated Code

- **Validation methods** automatically use `gerrors.MissingArgumentError()` for required fields
- **Middleware** uses appropriate gerrors types (e.g., RequireBody uses `MissingArgumentError`)
- **Handler templates** should be updated to use gerrors for consistent error responses
- **Enum conversion functions** automatically generated for all enum types:
  - `EnumFromType()` - Convert primitive values to enum constants
  - `EnumToType()` - Convert enum constants to primitive values
  - `EnumIsValid()` - Validate enum values
  - Supports all primitive types (string, int32, int64, float32, float64, bool)

## Workflow Instructions

**When modifying code generation:**
1. Update templates in `code_generator/generator/languages/`
2. Test with: `rm -rf code_generator/generator/test_data/server && [generate command]`
3. Verify generated code compiles: `go build code_generator/generator/test_data/server`
4. Run tests to ensure no regressions: `go test ./...`
5. Update snapshots if output changed intentionally: `UPDATE_SNAPS=true go test -v -run TestCoreGenerate ./code_generator/generator/...`

**When adding new language support:**
1. Create new directory in `code_generator/generator/languages/[language]/`
2. Add `config.yaml` with templates and type mappings
3. Update `code_generator/utils/allowlists.go` to register the new language/stack path
4. Update core generator to handle language-specific patterns

**When working with middleware:**
- Use `RouteWithInfo()` for generated routes to provide proper endpoint metadata
- Test middleware with example server in `code_generator/generator/test_data/main.go`
- Middleware should avoid HTTP primitives, use abstracted request/response types

**Automatic middleware features:**
- **RequireBody middleware**: Automatically applied to POST/PUT/PATCH endpoints based on `RequiresBody: true/false` in RouteInfo
- **Unified execution**: All routes follow the same middleware order: BeforeRouting → RequireBody → BeforeHandler → Handler → AfterHandler
- **Error handling**: Middleware uses gerrors package for consistent HTTP status codes

## Project Structure

```
bin/
└── gonzo-api.go              # CLI tool entry point

code_generator/
├── fileio/                   # File I/O operations
├── generator/
│   ├── core_generator.go     # Core generation engine
│   ├── json_generator.go     # Lexer and parser for .api files
│   ├── languages/
│   │   ├── go/server/        # Go server templates and config
│   │   └── typescript/client/ # TypeScript client templates
│   └── test_data/            # Test API definitions and generated code
└── utils/                    # Language/stack config path registry

runtime/                      # Runtime libraries imported by generated code
├── cookies/                  # Cookie handling
├── gerrors/                  # Structured error handling with HTTP status codes
├── handle/                   # Generic type-safe request handler
├── middleware/               # Middleware system
├── router/                   # HTTP routing
├── types/                    # Shared types (RouteInfo, etc.)
└── url/                      # URL parameter parsing

api-definition-language/      # VSCode syntax highlighting for .api files
```

## Development Environment

- **Go Version**: 1.22+ (see go.mod)
- **Dependencies**: Uses Go modules, run `go mod tidy` if needed
- **Templates**: YAML-based Go templates in `code_generator/generator/languages/` directories

## Using Gonzo from an App Repo

Consuming repos import gonzo as a Go module:

```go
import "github.com/medubin/gonzo/runtime/gerrors"
import "github.com/medubin/gonzo/runtime/router"
```

During local development of both gonzo and an app simultaneously, use Go workspaces:
```bash
go work init
go work use ../gonzo
go work use .
```

## Repository Etiquette

- **Generated Code**: Never manually edit generated files (types.go, server.go, etc.)
- **Tests**: Run tests before committing changes to core generator
- **API Definitions**: Use `.api` extension for API definition files
- **Templates**: Test template changes by regenerating test data
- **Commits**: Use descriptive commit messages, especially for template/generator changes
- **PRs**: Include example of generated code changes when modifying templates
- **CLAUDE.md**: Update this file when adding new commands, changing workflows, or modifying project structure

## API Definition Language

- **File Extension**: `.api`
- **Syntax**: See `API_SPEC.md` for complete language specification
- **Test Files**: Use `code_generator/generator/test_data/test_server.api` as reference
- **Types**: All struct fields are pointers with `omitempty` for proper JSON semantics
- **Validation**: Required fields use simple `nil` checks
- **Collections**: Arrays (`repeated(Type)`) and maps (`map(Key: Value)`) work anywhere
- **Map Keys**: Must be comparable types (primitives, enums, comparable structs)

## Debugging Tips

**Generation Issues:**
- Check template syntax in `code_generator/generator/languages/[lang]/[stack]/config.yaml`
- Verify template data structure matches what templates expect
- Use `rm -rf [output]` before regenerating to avoid stale files

**Runtime Issues:**
- Check middleware order and configuration
- Verify route registration uses `RouteWithInfo` for proper metadata
- Test with curl or example client calls

**Type Issues:**
- Ensure all fields are pointers for proper JSON unmarshaling
- Check that required field validation uses `== nil`
- Verify imports are only added when actually needed

## Testing

**Unit Tests:**
```bash
go test ./code_generator/...
go test ./runtime/...
```

**All Tests:**
```bash
go test ./...
```

**Snapshot Testing:**
Gonzo uses [go-snaps](https://github.com/gkampitakis/go-snaps) for snapshot testing to ensure generated code consistency:

```bash
# Run snapshot tests (compares against existing snapshots)
go test -v -run TestCoreGenerate ./code_generator/generator/...
go test -v -run TestJSONGenerate ./code_generator/generator/...

# Update snapshots when making intentional changes
UPDATE_SNAPS=true go test -v -run TestCoreGenerate ./code_generator/generator/...
```

**Snapshot test files:**
- `core_generator_test.go` - Tests Go server and TypeScript client generation
- `json_generator_test.go` - Tests API parsing to JSON structure
- Snapshots stored in `__snapshots__/` directories
- Use `snaps.MatchSnapshot()` for text content, `snaps.MatchJSON()` for JSON data

**Integration Tests:**
```bash
# Generate code and test compilation
rm -rf code_generator/generator/test_data/server
go run bin/gonzo-api.go generate -input code_generator/generator/test_data/test_server.api -language go -stack server -output code_generator/generator/test_data/server -package server
go build ./code_generator/generator/test_data/server/...

# Test server functionality
go run code_generator/generator/test_data/main.go &
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"username":"test","email":"test@example.com","password":"password"}'
```
