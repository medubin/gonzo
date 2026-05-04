# Gonzo API Language Specification

Gonzo API Language is a domain-specific language (DSL) for defining type-safe REST APIs. It generates server interfaces and client code for multiple target languages.

## File Extension

API definitions use the `.api` file extension.

## Comments

The language supports both single-line and multi-line comments:

```api
// Single-line comment
/* 
Multi-line comment
Can span multiple lines
*/
```

Comments can appear at the end of lines:
```api
type UserID int64 // This is also a comment
```

## Info Block

A single optional `info { ... }` block at the top level carries human-facing metadata about the API — version, description, contact, license. It is consumed by the OpenAPI generator to populate the spec's `info` object; Go and TypeScript targets ignore it.

```api
info {
  title "Customer API"
  version "1.4.2"
  description "Public-facing customer endpoints"
  contact "api@example.com"
  license "Apache-2.0"
}
```

**Recognized fields** (all optional, all string-valued):
- `title` — overrides the CLI `-package` flag as the OpenAPI title
- `version` — emitted as the OpenAPI `info.version` (default `0.0.0`)
- `description`
- `contact` — rendered as `contact.email` if it contains `@`, otherwise `contact.name`
- `license` — rendered as `license.name`

Unknown fields are an error (typos surface immediately rather than silently dropping metadata). At most one `info` block is allowed per file; duplicates are rejected.

The motivation for putting these in the language file rather than in CLI flags: version and contact are *content of the API*, not parameters of how it's generated. They belong next to the routes, version-controlled with them, so any tool that reads the `.api` file alone has enough context.

## Primitive Types

The language supports the following built-in primitive types:

- `string` - Text data
- `int32` - 32-bit signed integer  
- `int64` - 64-bit signed integer
- `float32` - 32-bit floating point
- `float64` - 64-bit floating point
- `bool` - Boolean (true/false)

## Type Definitions

### Type Aliases

Create aliases for primitive types:

```api
type UserID int64
type Email string  
type Timestamp int64
type Score float32
type IsActive bool
```

### Enums

Define enumerated values with an underlying primitive type:

```api
enum UserRole string {
  ADMIN = "admin"
  MODERATOR = "moderator"
  USER = "user"
}

enum UserStatus string {
  ACTIVE = "active"
  SUSPENDED = "suspended"
  PENDING = "pending"
}
```

**Enum Rules:**
- Must specify an underlying primitive type (`string`, `int32`, `int64`, etc.)
- Values must match the underlying type
- String values support escape sequences: `\"`, `\\`, `\n`, `\t`, `\r`

### Collections

Collections (arrays and maps) can be used anywhere in the API definition - as type definitions, struct fields, parameter types, return types, etc.

#### Repeated Types (Arrays/Lists)

```api
type UserList repeated(User)
type UserRoleHistory repeated(repeated(string))  // Nested arrays
type NestedUserGroups repeated(repeated(UserList))
```

#### Maps

```api
type UserPermissions map(string: bool)
type UserActivityByMonth map(string: map(int64: repeated(int32)))  // Complex nested maps
```

**Map Rules:**
- Key type is specified before the colon
- Value type is specified after the colon
- Value types can be any type (primitives, enums, structs, collections)
- Key types must be comparable (see constraints below)

**Map Key Type Constraints:**
Map keys must be comparable types (based on Go language requirements):
- ✅ **Allowed**: Primitives (`string`, `int32`, `int64`, `float32`, `float64`, `bool`)
- ✅ **Allowed**: Enums (comparable since they're based on primitives)  
- ✅ **Allowed**: Structs containing only comparable fields
- ❌ **Not Allowed**: `repeated(Type)` (slices are not comparable)
- ❌ **Not Allowed**: `map(KeyType: ValueType)` (maps are not comparable)

*The parser enforces these constraints: defining a map with a non-comparable key type (e.g. a `repeated`, another `map`, or a struct that transitively contains either) is rejected at parse time with a descriptive error.*

### Structs

Define structured data types:

```api
type User {
  ID UserID
  Username string
  Email Email
  Role UserRole
}
```

#### Required Fields

Fields are optional by default. Mark required fields with the `required` keyword:

```api
type CreateUserRequest {
  required Username string
  required Email Email
  required Password string
  Role UserRole        // Optional
  Profile UserProfile  // Optional
}
```

#### Supported Field Types

Struct fields can contain:
- Any primitive types
- Type aliases
- Enums
- Repeated fields: `repeated(Type)`
- Map fields: `map(KeyType: ValueType)`
- Other struct types
- Nested combinations

Example:
```api
type DetailedUser {
  required ID UserID
  Usernames repeated(string)
  LoginCount map(string: int32)
  Profile UserProfile
}
```

### Forward Declarations

Types can be referenced before they are defined:

```api
type UserList repeated(User)  // User is defined later
// ...
type User {
  ID UserID
  Username string
}
```

## Imports

API definition files can import other `.api` files to share types across multiple definitions.

### Flat Import

Merges all types, enums, and servers from the imported file into the current namespace:

```api
import "common/types.api"

// Types from types.api are now available directly
type UserProfile {
  ID UserID  // UserID defined in types.api
}
```

### Namespaced Import

Imports definitions under a namespace prefix to avoid name conflicts. All imported names are prefixed with the capitalized namespace:

```api
import "common/types.api" as "common"

// Reference imported types with namespace.TypeName syntax
type UserProfile {
  ID common.UserID  // becomes CommonUserID internally
}
```

**Import Rules:**
- Paths are relative to the importing file's directory
- Circular imports are silently skipped
- The same file cannot be imported both with and without a namespace, or under two different namespaces
- Name conflicts between imports and the current file are errors
- Namespaced imports prefix all type/enum/server names with `capitalize(namespace)` (e.g., `as "common"` → `Common` prefix)

## Server Definitions

Define REST API servers with endpoints:

```api
server UserService {
  GetUser GET /users/{id UserID} returns(DetailedUser)
  CreateUser POST /users body(CreateUserRequest) returns(User)
  UpdateUser PUT /users/{id UserID} body(UpdateUserRequest) returns(User)
  DeleteUser DELETE /users/{id UserID} body(DeleteUserRequest) returns(User)
  
  ListUsers GET /users parameters(UserListParams) returns(UserCollection)
  SearchUsers GET /users/search parameters(UserSearchParams) returns(UserCollection)
  GetUsersByRole GET /users/role/{role UserRole} parameters(UserListParams) returns(UserCollection)
}
```

### Endpoint Syntax

```
EndpointName METHOD /path/to/resource [body(Type)] [parameters(Type)] [returns(Type)]
```

#### Components:

- **EndpointName**: Identifier for the endpoint (PascalCase recommended)
- **METHOD**: HTTP verb (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD`, `OPTIONS`)
- **Path**: URL path with optional path parameters
- **body(Type)**: Optional request body type
- **parameters(Type)**: Optional query parameters type
- **returns(Type)**: Optional response type

#### Path Parameters

Embed typed parameters in the URL path:

```api
GET /users/{id UserID}
GET /users/{id UserID}/profile
GET /users/role/{role UserRole}
DELETE /users/{userId UserID}/notifications/{notificationId NotificationID}
```

**Path Parameter Rules:**
- Syntax: `{parameterName Type}`
- Can use primitive types, type aliases, or enums
- Multiple path parameters are supported

#### Request Bodies

Specify request body types (typically for POST, PUT, PATCH):

```api
CreateUser POST /users body(CreateUserRequest) returns(User)
UpdateUser PUT /users/{id UserID} body(UpdateUserRequest) returns(User)
```

**Body Rules:**
- Typically references a struct type
- GET endpoints typically don't have bodies
- Body is optional for all HTTP methods

#### Query Parameters

Specify query parameter types:

```api
ListUsers GET /users parameters(UserListParams) returns(UserCollection)
```

Where `UserListParams` might be:
```api
type UserListParams {
  page int32
  pageSize int32
  sortBy string
  sortOrder string
}
```

**Parameter Type Constraints:**
For query parameters, the struct can contain:
- Primitives: `string`, `int32`, `int64`, `float32`, `float64`, `bool`
- Enums (serialize to underlying primitive)
- Arrays: `repeated(Type)` of any type
- Maps: `map(KeyType: ValueType)` with any key/value types
- Any other defined types

#### Return Types

Specify response types:

```api
GetUser GET /users/{id UserID} returns(DetailedUser)
ListUsers GET /users returns(UserCollection)
GetUserNotifications GET /users/{userId UserID}/notifications returns(repeated(Notification))
```

**Return Type Rules:**
- Can be any defined type
- Can be `repeated(Type)` for arrays
- Can be omitted if endpoint returns no data

### Route Groups

Endpoints sharing a common path prefix can be nested in a `group` block. The group's path and any path parameters it declares are inherited by every endpoint inside it. Endpoints inside a group may omit their own path entirely (the endpoint then resolves to the group prefix). Groups can nest arbitrarily.

```api
server UserService {
  group /users/{id UserID} {
    GetUser    GET                                returns(DetailedUser)
    UpdateUser PUT          body(UpdateUserRequest) returns(User)
    PatchProfile PATCH /profile body(UserProfileUpdate) returns(UserProfile)

    group /notifications {
      ListUserNotifications GET                       returns(repeated(Notification))
      MarkRead              PUT /{nid NotificationID}
    }
  }

  CreateUser POST /users body(CreateUserRequest) returns(User)
}
```

Reusing a path-parameter name across nesting levels is an error.

### Multiple Servers

Multiple servers can be defined in a single file:

```api
server UserService {
  GetUser GET /users/{id UserID} returns(DetailedUser)
  CreateUser POST /users body(CreateUserRequest) returns(User)
}

server NotificationService {
  GetUserNotifications GET /users/{userId UserID}/notifications returns(repeated(Notification))
}
```

### API Versioning

Gonzo has no dedicated `version` keyword — versioning is composed from two existing features:

1. **Path versioning** uses a `group` for the version prefix. Endpoints inside the group inherit `/vN` automatically.
2. **Type versioning** uses namespaced imports. Each version lives in its own `.api` file; the root file imports them with an alias so `User` from v1 and v2 become distinct generated symbols (`V1User`, `V2User`) without manual renaming.

```api
// v1.api
type User { Name string }

// v2.api
type User { Name string; Email string }

// api.api
import "v1.api" as "v1"
import "v2.api" as "v2"

server API {
  group /v1 {
    GetUser GET /users/{id int64} returns(v1.User)
  }
  group /v2 {
    GetUser GET /users/{id int64} returns(v2.User)
  }
}
```

This covers URL-based versioning and type evolution. Header- or content-negotiation-based versioning (`Accept: application/vnd.api.v2+json`) is not currently supported and would need middleware in the consuming app.

### Decorators

Endpoints, `group` declarations, and struct fields can be annotated with `@decorator` lines stacked above the declaration. Decorators are open-vocabulary metadata: the parser collects them verbatim, and individual generators decide which names to honor. Unknown names are ignored, so templates can evolve independently.

```api
server UserService {
  @deprecated
  @auth("bearer")
  GetUser GET /users/{id UserID} returns(User)
}
```

**Syntax:**

- `@name` — no-arg form
- `@name(arg, arg, ...)` — positional args (string, number, bool literals)
- `@name(key: value, ...)` — named args, must come after any positional args

Decorators stack arbitrarily and may be placed above a `group` to cascade onto every endpoint inside (including endpoints inside nested groups). When the same decorator name appears on both a group and a nested endpoint, the endpoint's value wins (last-wins). Useful for `@auth`-style annotations that apply to whole sections of an API:

```api
server AdminAPI {
  @auth("bearer")
  group /admin {
    DeleteUser DELETE /users/{id UserID}
    PurgeAll   DELETE /purge

    @auth("none")
    Heartbeat GET /heartbeat   // exempt
  }

  Public GET /health           // unaffected
}
```

**Reaching arbitrary decorators from middleware:** every decorator on an endpoint — known or not — is emitted into `RouteInfo.Decorators`. Middleware can dispatch on names the generator does not consume itself:

```go
func CacheMiddleware(next handle.Handler) handle.Handler {
    return func(req handle.Request) handle.Response {
        if d := req.RouteInfo().Find("cache"); d != nil {
            if maxAge, ok := d.Kwargs["maxAge"]; ok {
                req.ResponseHeader().Set("Cache-Control", "max-age="+maxAge.Value)
            }
        }
        return next(req)
    }
}
```

So `@cache(maxAge: 60)` works end-to-end without any code-generator change — just write the middleware that reads it. Argument values are exposed as `(Kind, Value)` pairs where `Kind` is `"string"`, `"number"`, or `"bool"` and `Value` is the raw lexeme; numeric/bool callers parse on demand.

**Built-in decorators:**

#### `@validation(...)` (field-level)

Attach constraint metadata to a struct field. Surfaces in three places:

- **Go server**: the generated `Validate()` method enforces each constraint after the existing required-field checks. Failures return `gerrors.InvalidArgumentError` (HTTP 400).
- **TypeScript client**: the generated `validate{TypeName}` helper applies the same checks before sending a request.
- **OpenAPI**: constraints become `minLength`/`maxLength`/`minimum`/`maximum`/`pattern`/`format` on the field schema.

```api
type CreateUserRequest {
  @validation(minLength: 3, maxLength: 32, pattern: "^[a-z0-9_]+$")
  required Username string

  @validation(format: "email")
  required Email string

  @validation(minLength: 8, maxLength: 128)
  required Password string

  @validation(min: 13, max: 120)
  Age int32
}
```

**Recognized kwargs** (all optional):

- `min`, `max` — numeric bounds (inclusive). Apply to int/float fields.
- `minLength`, `maxLength` — string-length bounds (inclusive).
- `pattern` — RE2 regex; must match the entire value at runtime in Go (uses `regexp.MatchString`) and in TS (uses `RegExp.test`).
- `format` — semantic hint passed through to OpenAPI as `format`. No runtime check today; useful for `"email"`, `"uuid"`, `"url"`, etc.

Constraints on optional fields only run when the field is present (non-nil in Go, defined in TS). Constraints on required fields run after the nil check.

There is no parser-side check that constraints match the field type — putting `min` on a string is silently ignored by the runtime checks but emitted into OpenAPI as `minimum`. Treat that as user discipline for now.

When a field's type renders as an OpenAPI `$ref` (i.e., a named non-primitive type), validation keywords are dropped from the spec to avoid producing an invalid document. The Go/TS runtime checks still fire — only the spec is lossy in that case.

#### `@deprecated` / `@deprecated("message")`

Marks an endpoint as deprecated. The optional message is surfaced wherever the host language has a convention:

- **Go server**: `// Deprecated: <message>` doc comment on the interface method (`gopls` flags callers).
- **TypeScript client**: `/** @deprecated <message> */` JSDoc on the client method.
- **OpenAPI**: `deprecated: true` on the operation; tooling like Swagger UI strikes the route through.

```api
@deprecated("use SearchUsers instead")
GetUsersByRole GET /users/role/{role UserRole} returns(UserCollection)
```

#### `@auth("<scheme>")`

Marks the route as requiring a particular authentication scheme. The scheme name is a contract label — Gonzo does **not** generate any auth verification code. Two things happen:

1. The Go server template populates `RouteInfo.AuthScheme` with the scheme name. The consuming app's middleware reads this field and decides what to enforce (token validation, scope checks, etc.).
2. The OpenAPI generator emits a per-operation `security:` requirement and declares the scheme under `components.securitySchemes`.

Recognized scheme names with default OpenAPI mappings:

- `"bearer"` → `http` `bearer` with `bearerFormat: JWT`
- `"apiKey"` → `apiKey` in header `X-API-Key`
- `"none"` → explicit opt-out (no security emitted; useful when default-on auth middleware should skip a public route)
- Any other name → falls back to `http bearer` in the spec; `RouteInfo.AuthScheme` carries the name verbatim so app middleware can dispatch on it

```api
server PaymentService {
  @auth("bearer")
  Charge POST /charges body(ChargeReq) returns(Charge)

  @auth("none")
  Health GET /health
}
```

In the consuming app:

```go
func RequireAuth(next handle.Handler) handle.Handler {
    return func(req handle.Request) handle.Response {
        info := req.RouteInfo()
        if info.AuthScheme == "" || info.AuthScheme == "none" {
            return next(req) // public
        }
        // app's own token verification keyed off info.AuthScheme
        ...
    }
}
```

## Best Practices

### Naming Conventions

- **Types**: PascalCase (`User`, `UserProfile`, `CreateUserRequest`)
- **Enum Values**: SCREAMING_SNAKE_CASE (`ADMIN`, `MODERATOR`, `USER`)
- **Endpoints**: PascalCase (`GetUser`, `CreateUser`, `ListUsers`)
- **Fields**: camelCase (`firstName`, `isActive`) or PascalCase (`FirstName`, `IsActive`)

### Type Organization

1. **Primitives first**: Define type aliases at the top
2. **Enums**: Define after primitives
3. **Collections**: Define simple collections early
4. **Structs**: Define in dependency order when possible
5. **Servers**: Define at the end

### Parameter Design

- **Path parameters**: Use for resource identifiers (`{id UserID}`)
- **Query parameters**: Use for filtering, pagination, sorting
- **Request bodies**: Use for creating or updating resources

## Language Generation

The API definition generates code for multiple target languages:

- **Go**: Server interfaces, types, and routing
- **TypeScript**: Client classes, interfaces, and types

Each target language has its own configuration and templates for code generation.

## Example Complete API

```api
// Type aliases
type UserID int64
type Email string

// Enums
enum UserRole string {
  ADMIN = "admin"
  USER = "user"
}

// Collections
type UserCollection repeated(User)

// Structs
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

// Server definition
server UserService {
  GetUser GET /users/{id UserID} returns(User)
  CreateUser POST /users body(CreateUserRequest) returns(User)
  ListUsers GET /users parameters(UserListParams) returns(UserCollection)
}
```

This generates type-safe server interfaces and client code with proper parameter extraction, type conversion, and routing.