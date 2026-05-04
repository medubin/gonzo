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
	Decorators   []Decorator // every @decorator on the endpoint, in source order. Middleware reads these to act on annotations the generator does not consume itself (e.g. @cache, @rateLimit, @tag).
}

// Decorator is a single @name(...) annotation surfaced to middleware. Its
// shape mirrors what the parser produces; the generator emits it verbatim
// into RouteInfo so consuming apps can dispatch on arbitrary decorator names
// without requiring a code-generator change.
type Decorator struct {
	Name   string
	Args   []DecoratorArg
	Kwargs map[string]DecoratorArg
}

// DecoratorArg is a literal scalar argument. Kind is one of "string",
// "number", "bool". Value is the raw lexeme — for strings, the
// already-unescaped text; for numbers and bools, the source representation.
type DecoratorArg struct {
	Kind  string
	Value string
}

// Find returns the first decorator with the given name, or nil if absent.
func (r *RouteInfo) Find(name string) *Decorator {
	for i := range r.Decorators {
		if r.Decorators[i].Name == name {
			return &r.Decorators[i]
		}
	}
	return nil
}