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

*Note: The code generator does not currently enforce these constraints, but using non-comparable key types will result in invalid Go code that fails to compile.*

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
- **METHOD**: HTTP verb (`GET`, `POST`, `PUT`, `PATCH`, `DELETE`)
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