# Gonzo - Claude AI Assistant Guide

This guide helps Claude AI effectively work with the Gonzo code generation project.

## About Gonzo

Gonzo is a modern code generation tool that creates type-safe web API servers and clients from a custom API definition language. It consists of:

- **Core Parser/Generator**: Lexer, parser, and code generation engine
- **Language Templates**: Go server and TypeScript client templates  
- **Runtime Libraries**: HTTP routing, middleware, error handling, URL parsing
- **CLI Tool**: Command-line interface for code generation
- **Example Implementation**: Complete user authentication API

## Common Bash Commands

**Generate Go server code from API definition:**
```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -language go -stack server -output api/code_generator/generator/test_data/server -package server
```

**Generate TypeScript client code:**
```bash
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -language typescript -stack client -output api/code_generator/generator/test_data/client -package client
```

**Run the example server:**
```bash
go run api/code_generator/generator/test_data/main.go
```

**Run the main application:**
```bash
make run
# OR
go run internal/main.go
```

**Run tests:**
```bash
go test ./...
go test -v api/code_generator/generator/...
```

**Database operations:**
```bash
make db-migration          # Run migrations
make db-migration-down     # Rollback migrations
make generate-sqlc         # Generate database code
make new-db-migration name=create_table  # Create new migration
```

## Code Style Guidelines

- **Go**: Follow standard Go conventions, use `gofmt`
- **Template syntax**: Use Go template syntax in `.yaml` config files
- **File structure**: Generated files should not be manually edited
- **Error handling**: Use the `gerrors` package for consistent error responses
- **Imports**: Only import packages that are actually used to avoid unused import errors
- **Field types**: All struct fields should be pointers with `omitempty` JSON tags
- **Validation**: Required fields validated with simple `== nil` checks

## Workflow Instructions

**When modifying code generation:**
1. Update templates in `api/code_generator/generator/languages/`
2. Test with: `rm -rf api/code_generator/generator/test_data/server && [generate command]`
3. Verify generated code compiles: `go build api/code_generator/generator/test_data/server`
4. Run tests to ensure no regressions

**When adding new language support:**
1. Create new directory in `api/code_generator/generator/languages/[language]/`
2. Add `config.yaml` with templates and type mappings
3. Update core generator to handle language-specific patterns

**When working with middleware:**
- Use `RouteWithInfo()` for generated routes to provide proper endpoint metadata
- Test middleware with example server in `api/code_generator/generator/test_data/main.go`
- Middleware should avoid HTTP primitives, use abstracted request/response types

## Project Structure

```
api/
├── bin/gonzo-api.go              # CLI tool entry point
├── code_generator/
│   ├── character_reader/         # Lexer utilities
│   ├── fileio/                   # File I/O operations
│   ├── generator/
│   │   ├── core_generator.go     # Core generation engine
│   │   ├── json_generator.go     # JSON parsing
│   │   ├── languages/
│   │   │   ├── go/server/        # Go templates and config
│   │   │   └── typescript/client/ # TypeScript templates
│   │   └── test_data/            # Test API definitions and generated code
│   └── utils/                    # Language configuration utilities
└── src/                          # Runtime libraries
    ├── cookies/                  # Cookie handling
    ├── gerrors/                  # Error handling
    ├── handle/                   # Generic request handler
    ├── middleware/               # Middleware system
    ├── router/                   # HTTP routing
    └── url/                      # URL parameter parsing

server/                           # Generated server implementation
db/                              # Database migrations and queries (sqlc)
internal/                        # Main application entry point
```

## Development Environment

- **Go Version**: 1.18+ (see go.mod)
- **Database**: PostgreSQL (for example app)
- **Dependencies**: Uses Go modules, run `go mod tidy` if needed
- **Templates**: YAML-based Go templates in `languages/` directories

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
- **Test Files**: Use `api/code_generator/generator/test_data/test_server.api` as reference
- **Types**: All struct fields are pointers with `omitempty` for proper JSON semantics
- **Validation**: Required fields use simple `nil` checks
- **Collections**: Arrays (`repeated(Type)`) and maps (`map(Key: Value)`) work anywhere
- **Map Keys**: Must be comparable types (primitives, enums, comparable structs)

## Debugging Tips

**Generation Issues:**
- Check template syntax in `languages/[lang]/[stack]/config.yaml`
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
go test api/code_generator/generator/
go test api/src/url/
go test api/code_generator/fileio/
go test api/code_generator/utils/
```

**All Tests:**
```bash
go test ./...
go test -v api/code_generator/...
```

**Snapshot Testing:**
Gonzo uses [go-snaps](https://github.com/gkampitakis/go-snaps) for snapshot testing to ensure generated code consistency:

```bash
# Run snapshot tests (compares against existing snapshots)
go test -v -run TestCoreGenerate
go test -v -run TestJSONGenerate

# Update snapshots when making intentional changes
# Snapshots are auto-generated on first run or when missing
```

**Snapshot test files:**
- `core_generator_test.go` - Tests Go server and TypeScript client generation
- `json_generator_test.go` - Tests API parsing to JSON structure
- Snapshots stored in `__snapshots__/` directories
- Use `snaps.MatchSnapshot()` for text content, `snaps.MatchJSON()` for JSON data

**Integration Tests:**
```bash
# Generate code and test compilation
rm -rf api/code_generator/generator/test_data/server
go run api/bin/gonzo-api.go generate -input api/code_generator/generator/test_data/test_server -language go -stack server -output api/code_generator/generator/test_data/server -package server
go build api/code_generator/generator/test_data/server

# Test server functionality
go run api/code_generator/generator/test_data/main.go &
curl -X POST http://localhost:8080/users -H "Content-Type: application/json" -d '{"username":"test","email":"test@example.com","password":"password"}'
```