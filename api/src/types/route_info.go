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
}