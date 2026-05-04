package types

// RouteInfo provides complete metadata about a route for middleware and tooling
type RouteInfo struct {
	Method       string // "GET", "POST", etc.
	Path         string // "/users/{id}"
	Endpoint     string // "CreateUser"
	Server       string // "UserService" 
	BodyType     string // "CreateUserRequest"
	ReturnType   string // "User"
	RequiresBody bool   // Whether this endpoint requires a request body
	IsMultipart  bool   // Whether this endpoint expects multipart/form-data
	AuthScheme   string // "" if unannotated; otherwise the @auth scheme name ("bearer", "apiKey", "none", ...). Middleware in the consuming app reads this and decides what to enforce.
}