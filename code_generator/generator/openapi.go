package generator

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// numericLess compares two enum value strings as floats, falling back to
// string comparison if either fails to parse.
func numericLess(a, b string) bool {
	af, aerr := strconv.ParseFloat(a, 64)
	bf, berr := strconv.ParseFloat(b, 64)
	if aerr == nil && berr == nil {
		return af < bf
	}
	return a < b
}

// RenderOpenAPI produces an OpenAPI 3.1 document for the given API definition.
// The output is YAML and intended to be written verbatim into openapi.yaml.
//
// Scope (v1):
//   - paths: every endpoint with method, path/query params, JSON or multipart
//     request body, and a 200 JSON response (or 204 No Content if the endpoint
//     declares no return type).
//   - components.schemas: every defined type and enum, plus a shared GonzoError
//     schema that backs the default error response.
//   - HTTP methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS.
//
// Out of scope: tags/grouping, security schemes, multi-status responses,
// example values, header declarations, OpenAPI extensions.
func RenderOpenAPI(api *APIDefinition, title string) (string, error) {
	if api == nil {
		return "", fmt.Errorf("openapi: nil APIDefinition")
	}
	r := &openapiRenderer{
		api:        api,
		typeKinds:  indexTypes(api),
		multiparts: indexMultipartTypes(api),
		enums:      indexEnums(api),
	}
	return r.render(title)
}

type openapiRenderer struct {
	api        *APIDefinition
	typeKinds  map[string]string // name -> Kind ("alias", "struct", "repeated", "map")
	multiparts map[string]bool   // struct names that contain a file field
	enums      map[string]*EnumDef
}

func indexTypes(api *APIDefinition) map[string]string {
	out := make(map[string]string, len(api.Types))
	for _, t := range api.Types {
		out[t.Name] = t.Kind
	}
	return out
}

func indexEnums(api *APIDefinition) map[string]*EnumDef {
	out := make(map[string]*EnumDef, len(api.Enums))
	for i := range api.Enums {
		out[api.Enums[i].Name] = &api.Enums[i]
	}
	return out
}

// indexMultipartTypes returns the set of struct type names that contain a
// `file` field (which makes the struct a multipart request body).
func indexMultipartTypes(api *APIDefinition) map[string]bool {
	out := make(map[string]bool)
	for _, t := range api.Types {
		if t.Kind != "struct" {
			continue
		}
		for _, f := range t.Fields {
			if f.Type != nil && f.Type.Kind == "reference" && f.Type.Name == "file" {
				out[t.Name] = true
				break
			}
		}
	}
	return out
}

func (r *openapiRenderer) render(title string) (string, error) {
	var b strings.Builder

	// Resolve fields from the `info { ... }` block if present, with the
	// caller-supplied title as a last-resort fallback.
	resolved := r.api.Info
	if resolved == nil {
		resolved = &InfoDef{}
	}
	if resolved.Title == "" {
		resolved.Title = title
	}
	if resolved.Title == "" {
		resolved.Title = "API"
	}
	version := resolved.Version
	if version == "" {
		version = "0.0.0"
	}

	b.WriteString("openapi: 3.1.0\n")
	b.WriteString("info:\n")
	b.WriteString(fmt.Sprintf("  title: %s\n", yamlQuote(resolved.Title)))
	b.WriteString(fmt.Sprintf("  version: %s\n", yamlQuote(version)))
	if resolved.Description != "" {
		b.WriteString(fmt.Sprintf("  description: %s\n", yamlQuote(resolved.Description)))
	}
	// Contact is treated as an email when it contains '@', otherwise as a
	// free-form name. OpenAPI's contact object accepts either; this keeps
	// the language surface to a single string while still emitting a valid
	// spec.
	if resolved.Contact != "" {
		b.WriteString("  contact:\n")
		if strings.Contains(resolved.Contact, "@") {
			b.WriteString(fmt.Sprintf("    email: %s\n", yamlQuote(resolved.Contact)))
		} else {
			b.WriteString(fmt.Sprintf("    name: %s\n", yamlQuote(resolved.Contact)))
		}
	}
	if resolved.License != "" {
		b.WriteString("  license:\n")
		b.WriteString(fmt.Sprintf("    name: %s\n", yamlQuote(resolved.License)))
	}

	r.renderTags(&b)
	r.renderPaths(&b)
	r.renderComponents(&b)

	return b.String(), nil
}

// renderTags emits a top-level `tags:` block listing each server declaration,
// so explorers like Swagger UI / Redoc render a section per server. Operations
// reference these via per-operation `tags:` (see renderOperation).
func (r *openapiRenderer) renderTags(b *strings.Builder) {
	if len(r.api.Servers) == 0 {
		return
	}
	b.WriteString("tags:\n")
	for i := range r.api.Servers {
		b.WriteString("  - name: " + yamlQuote(r.api.Servers[i].Name) + "\n")
	}
}

func (r *openapiRenderer) renderPaths(b *strings.Builder) {
	// Group endpoints by path so methods on the same path nest under one key.
	type endpointEntry struct {
		server *ServerDef
		ep     *EndpointDef
	}
	byPath := make(map[string][]endpointEntry)
	var pathOrder []string
	for i := range r.api.Servers {
		server := &r.api.Servers[i]
		for j := range server.Endpoints {
			ep := &server.Endpoints[j]
			if _, exists := byPath[ep.Path]; !exists {
				pathOrder = append(pathOrder, ep.Path)
			}
			byPath[ep.Path] = append(byPath[ep.Path], endpointEntry{server, ep})
		}
	}

	if len(pathOrder) == 0 {
		b.WriteString("paths: {}\n")
		return
	}

	b.WriteString("paths:\n")
	for _, path := range pathOrder {
		b.WriteString(fmt.Sprintf("  %s:\n", yamlQuote(path)))
		for _, entry := range byPath[path] {
			r.renderOperation(b, entry.ep, entry.server.Name)
		}
	}
}

func (r *openapiRenderer) renderOperation(b *strings.Builder, ep *EndpointDef, serverName string) {
	method := strings.ToLower(ep.Method)
	indent := "    "
	b.WriteString(fmt.Sprintf("%s%s:\n", indent, method))
	if serverName != "" {
		b.WriteString(indent + "  tags:\n")
		b.WriteString(indent + "    - " + yamlQuote(serverName) + "\n")
	}
	b.WriteString(fmt.Sprintf("%s  operationId: %s\n", indent, ep.Name))
	if hasDecorator(ep, "deprecated") {
		b.WriteString(indent + "  deprecated: true\n")
	}

	r.renderParameters(b, ep, indent+"  ")
	r.renderRequestBody(b, ep, indent+"  ")
	r.renderResponses(b, ep, indent+"  ")
	r.renderSecurity(b, ep, indent+"  ")
}

// renderSecurity emits the per-operation `security:` requirement when an
// endpoint carries an @auth decorator. `@auth("none")` is treated as an
// explicit opt-out (no requirement emitted) so that a default-on auth
// middleware can be overridden in-source. Unknown scheme names are passed
// through verbatim — `renderComponents` will declare a generic scheme for them
// so the document still validates.
func (r *openapiRenderer) renderSecurity(b *strings.Builder, ep *EndpointDef, indent string) {
	scheme := authSchemeFor(ep)
	if scheme == "" || scheme == "none" {
		return
	}
	b.WriteString(indent + "security:\n")
	b.WriteString(indent + "  - " + securitySchemeID(scheme) + ": []\n")
}

// hasDecorator reports whether ep carries a decorator with the given name.
func hasDecorator(ep *EndpointDef, name string) bool {
	for _, d := range ep.Decorators {
		if d.Name == name {
			return true
		}
	}
	return false
}

// authSchemeFor returns the @auth scheme name on ep, or "". Last @auth wins.
func authSchemeFor(ep *EndpointDef) string {
	scheme := ""
	for _, d := range ep.Decorators {
		if d.Name == "auth" && len(d.Args) >= 1 && d.Args[0].Kind == "string" {
			scheme = d.Args[0].Value
		}
	}
	return scheme
}

// securitySchemeID maps an @auth scheme name to the OpenAPI components key
// used to reference it. Stable, lowercase, suffixed with "Auth" for
// readability ("bearer" → "bearerAuth", "apiKey" → "apiKeyAuth", "foo" →
// "fooAuth").
func securitySchemeID(scheme string) string {
	return scheme + "Auth"
}

// collectAuthSchemes walks every endpoint and returns the deduplicated set of
// non-"none" @auth scheme names actually used in the document, sorted for
// stable output.
func (r *openapiRenderer) collectAuthSchemes() []string {
	seen := make(map[string]bool)
	for i := range r.api.Servers {
		for j := range r.api.Servers[i].Endpoints {
			s := authSchemeFor(&r.api.Servers[i].Endpoints[j])
			if s == "" || s == "none" {
				continue
			}
			seen[s] = true
		}
	}
	out := make([]string, 0, len(seen))
	for s := range seen {
		out = append(out, s)
	}
	sort.Strings(out)
	return out
}

func (r *openapiRenderer) renderParameters(b *strings.Builder, ep *EndpointDef, indent string) {
	type param struct {
		name     string
		in       string
		required bool
		schema   *TypeExpr
	}
	var params []param
	for _, p := range ep.PathParams {
		params = append(params, param{
			name:     p.Name,
			in:       "path",
			required: true,
			schema:   &TypeExpr{Kind: "reference", Name: p.Type},
		})
	}
	if ep.Parameters != nil {
		// Resolve the parameters TypeExpr to the underlying struct fields.
		fields := r.resolveStructFields(ep.Parameters)
		for _, f := range fields {
			params = append(params, param{
				name:     f.Name,
				in:       "query",
				required: f.Required,
				schema:   f.Type,
			})
		}
	}
	if len(params) == 0 {
		return
	}
	b.WriteString(indent + "parameters:\n")
	for _, p := range params {
		b.WriteString(indent + "  - name: " + yamlQuote(p.name) + "\n")
		b.WriteString(indent + "    in: " + p.in + "\n")
		if p.required {
			b.WriteString(indent + "    required: true\n")
		}
		b.WriteString(indent + "    schema:\n")
		b.WriteString(r.renderSchema(p.schema, len(indent)+6))
	}
}

func (r *openapiRenderer) renderRequestBody(b *strings.Builder, ep *EndpointDef, indent string) {
	if ep.Body == nil {
		return
	}
	b.WriteString(indent + "requestBody:\n")
	b.WriteString(indent + "  required: true\n")
	b.WriteString(indent + "  content:\n")

	contentType := "application/json"
	if ep.Body.Kind == "reference" && r.multiparts[ep.Body.Name] {
		contentType = "multipart/form-data"
	}
	b.WriteString(indent + "    " + contentType + ":\n")
	b.WriteString(indent + "      schema:\n")
	b.WriteString(r.renderSchema(ep.Body, len(indent)+8))
}

func (r *openapiRenderer) renderResponses(b *strings.Builder, ep *EndpointDef, indent string) {
	b.WriteString(indent + "responses:\n")
	if ep.Returns == nil {
		b.WriteString(indent + "  '204':\n")
		b.WriteString(indent + "    description: No Content\n")
	} else {
		b.WriteString(indent + "  '200':\n")
		b.WriteString(indent + "    description: OK\n")
		b.WriteString(indent + "    content:\n")
		b.WriteString(indent + "      application/json:\n")
		b.WriteString(indent + "        schema:\n")
		b.WriteString(r.renderSchema(ep.Returns, len(indent)+10))
	}
	b.WriteString(indent + "  default:\n")
	b.WriteString(indent + "    description: Error\n")
	b.WriteString(indent + "    content:\n")
	b.WriteString(indent + "      application/json:\n")
	b.WriteString(indent + "        schema:\n")
	b.WriteString(indent + "          $ref: '#/components/schemas/GonzoError'\n")
}

func (r *openapiRenderer) renderComponents(b *strings.Builder) {
	authSchemes := r.collectAuthSchemes()

	if len(r.api.Types) == 0 && len(r.api.Enums) == 0 {
		// Still need GonzoError so default error responses can resolve.
		b.WriteString("components:\n")
		b.WriteString("  schemas:\n")
		b.WriteString(r.gonzoErrorSchema(4))
		r.renderSecuritySchemes(b, authSchemes, 2)
		return
	}

	b.WriteString("components:\n")
	b.WriteString("  schemas:\n")

	// Stable order: alphabetical by name across both types and enums.
	type entry struct {
		name   string
		render func(indent int) string
	}
	var entries []entry
	for i := range r.api.Types {
		td := &r.api.Types[i]
		entries = append(entries, entry{td.Name, func(ind int) string { return r.renderTypeDefSchema(td, ind) }})
	}
	for name, ed := range r.enums {
		ed := ed
		_ = name
		entries = append(entries, entry{ed.Name, func(ind int) string { return r.renderEnumSchema(ed, ind) }})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].name < entries[j].name })

	for _, e := range entries {
		b.WriteString(fmt.Sprintf("    %s:\n", e.name))
		b.WriteString(e.render(6))
	}

	b.WriteString(r.gonzoErrorSchema(4))
	r.renderSecuritySchemes(b, authSchemes, 2)
}

// renderSecuritySchemes writes a `securitySchemes:` block under `components:`
// declaring one entry per used @auth scheme. Defaults are conservative:
//
//   - "bearer" → http bearer with bearerFormat: JWT
//   - "apiKey" → apiKey in header X-API-Key
//   - any other name → http bearer (a permissive fallback so unknown schemes
//     still produce a valid document; a future top-level scheme-declaration
//     syntax will let users override these defaults)
func (r *openapiRenderer) renderSecuritySchemes(b *strings.Builder, schemes []string, baseIndent int) {
	if len(schemes) == 0 {
		return
	}
	pad := strings.Repeat(" ", baseIndent)
	body := strings.Repeat(" ", baseIndent+2)
	b.WriteString(pad + "securitySchemes:\n")
	for _, s := range schemes {
		b.WriteString(body + securitySchemeID(s) + ":\n")
		switch s {
		case "apiKey":
			b.WriteString(body + "  type: apiKey\n")
			b.WriteString(body + "  in: header\n")
			b.WriteString(body + "  name: X-API-Key\n")
		case "bearer":
			b.WriteString(body + "  type: http\n")
			b.WriteString(body + "  scheme: bearer\n")
			b.WriteString(body + "  bearerFormat: JWT\n")
		default:
			b.WriteString(body + "  type: http\n")
			b.WriteString(body + "  scheme: bearer\n")
		}
	}
}

func (r *openapiRenderer) gonzoErrorSchema(baseIndent int) string {
	pad := strings.Repeat(" ", baseIndent)
	body := strings.Repeat(" ", baseIndent+2)
	return pad + "GonzoError:\n" +
		body + "type: object\n" +
		body + "properties:\n" +
		body + "  error:\n" +
		body + "    type: string\n" +
		body + "    description: 'Error envelope: \"<code>: <message>\"'\n" +
		body + "required:\n" +
		body + "  - error\n"
}

func (r *openapiRenderer) renderTypeDefSchema(td *TypeDef, indent int) string {
	pad := strings.Repeat(" ", indent)
	switch td.Kind {
	case "alias":
		// Alias to a primitive renders as the primitive type; alias to a named
		// type renders as a $ref to that schema.
		if mapping, ok := primitiveOpenAPI(td.Target); ok {
			return mapping.toYAML(indent)
		}
		return pad + fmt.Sprintf("$ref: '#/components/schemas/%s'\n", td.Target)
	case "struct":
		return r.renderStructBody(td.Fields, indent)
	case "repeated":
		out := pad + "type: array\n" + pad + "items:\n"
		out += r.renderSchema(td.ElementType, indent+2)
		return out
	case "map":
		out := pad + "type: object\n" + pad + "additionalProperties:\n"
		out += r.renderSchema(td.ValueType, indent+2)
		return out
	}
	return pad + "{}\n"
}

// renderValidationConstraints emits OpenAPI keywords for any @validation
// decorator on the field. Keys are placed at the same indent as the rest of
// the schema body so they sit alongside `type:`, `format:`, etc.
//
// Skipped when the field renders as a `$ref` (a reference to a named
// non-primitive type). OpenAPI 3.1 forbids most validation keywords next to
// `$ref`; honoring constraints there would produce an invalid spec. The
// runtime `Validate()` still enforces them — the spec is just lossy in this
// edge case until we fix it via an `allOf` wrapper.
func renderValidationConstraints(f *FieldDef, indent int) string {
	if rendersAsRef(f.Type) {
		return ""
	}
	pad := strings.Repeat(" ", indent)
	var b strings.Builder
	for _, d := range f.Decorators {
		if d.Name != "validation" {
			continue
		}
		for _, kw := range d.Kwargs {
			switch kw.Name {
			case "min":
				b.WriteString(pad + "minimum: " + kw.Arg.Value + "\n")
			case "max":
				b.WriteString(pad + "maximum: " + kw.Arg.Value + "\n")
			case "minLength":
				b.WriteString(pad + "minLength: " + kw.Arg.Value + "\n")
			case "maxLength":
				b.WriteString(pad + "maxLength: " + kw.Arg.Value + "\n")
			case "pattern":
				b.WriteString(pad + "pattern: " + yamlQuote(kw.Arg.Value) + "\n")
			case "format":
				b.WriteString(pad + "format: " + yamlQuote(kw.Arg.Value) + "\n")
			}
		}
	}
	return b.String()
}

// rendersAsRef reports whether the type expression would emit `$ref: ...` in
// the generated schema. References to primitives render inline as
// `type: string` etc., so they are not refs for this purpose.
func rendersAsRef(expr *TypeExpr) bool {
	if expr == nil || expr.Kind != "reference" {
		return false
	}
	_, isPrimitive := primitiveOpenAPI(expr.Name)
	return !isPrimitive
}

func (r *openapiRenderer) renderStructBody(fields []FieldDef, indent int) string {
	pad := strings.Repeat(" ", indent)
	var b strings.Builder
	b.WriteString(pad + "type: object\n")
	if len(fields) == 0 {
		b.WriteString(pad + "properties: {}\n")
		return b.String()
	}
	b.WriteString(pad + "properties:\n")
	var required []string
	for _, f := range fields {
		b.WriteString(pad + "  " + yamlQuote(f.Name) + ":\n")
		b.WriteString(r.renderSchema(f.Type, indent+4))
		b.WriteString(renderValidationConstraints(&f, indent+4))
		if f.Required {
			required = append(required, f.Name)
		}
	}
	if len(required) > 0 {
		b.WriteString(pad + "required:\n")
		for _, name := range required {
			b.WriteString(pad + "  - " + yamlQuote(name) + "\n")
		}
	}
	return b.String()
}

func (r *openapiRenderer) renderEnumSchema(ed *EnumDef, indent int) string {
	pad := strings.Repeat(" ", indent)
	mapping, ok := primitiveOpenAPI(ed.BaseType)
	if !ok {
		mapping = openapiPrimitive{Type: "string"}
	}
	var b strings.Builder
	b.WriteString(pad + "type: " + mapping.Type + "\n")
	if mapping.Format != "" {
		b.WriteString(pad + "format: " + mapping.Format + "\n")
	}

	// Emit values in deterministic order: numeric ascending for integer-based
	// enums, lexical ascending for everything else. (The parser stores values
	// in a map, so source order is unrecoverable.)
	values := make([]string, 0, len(ed.Values))
	for _, v := range ed.Values {
		values = append(values, v)
	}
	if mapping.Type == "integer" || mapping.Type == "number" {
		sort.Slice(values, func(i, j int) bool { return numericLess(values[i], values[j]) })
	} else {
		sort.Strings(values)
	}

	b.WriteString(pad + "enum:\n")
	for _, v := range values {
		if mapping.Type == "string" {
			b.WriteString(pad + "  - " + yamlQuote(v) + "\n")
		} else {
			b.WriteString(pad + "  - " + v + "\n")
		}
	}
	return b.String()
}

func (r *openapiRenderer) renderSchema(expr *TypeExpr, indent int) string {
	pad := strings.Repeat(" ", indent)
	if expr == nil {
		return pad + "{}\n"
	}
	switch expr.Kind {
	case "reference":
		if mapping, ok := primitiveOpenAPI(expr.Name); ok {
			return mapping.toYAML(indent)
		}
		return pad + fmt.Sprintf("$ref: '#/components/schemas/%s'\n", expr.Name)
	case "repeated":
		out := pad + "type: array\n" + pad + "items:\n"
		out += r.renderSchema(expr.ElementType, indent+2)
		return out
	case "map":
		out := pad + "type: object\n" + pad + "additionalProperties:\n"
		out += r.renderSchema(expr.ValueType, indent+2)
		return out
	}
	return pad + "{}\n"
}

// resolveStructFields walks an alias chain and returns the underlying struct's
// fields. Returns nil if the expression doesn't resolve to a struct (e.g. a
// repeated/map type used as `parameters(...)`).
func (r *openapiRenderer) resolveStructFields(expr *TypeExpr) []FieldDef {
	if expr == nil || expr.Kind != "reference" {
		return nil
	}
	visited := make(map[string]bool)
	name := expr.Name
	for !visited[name] {
		visited[name] = true
		for i := range r.api.Types {
			td := &r.api.Types[i]
			if td.Name != name {
				continue
			}
			switch td.Kind {
			case "struct":
				return td.Fields
			case "alias":
				name = td.Target
			default:
				return nil
			}
		}
		// name not found among defined types
		return nil
	}
	return nil
}

type openapiPrimitive struct {
	Type   string
	Format string
}

func (p openapiPrimitive) toYAML(indent int) string {
	pad := strings.Repeat(" ", indent)
	out := pad + "type: " + p.Type + "\n"
	if p.Format != "" {
		out += pad + "format: " + p.Format + "\n"
	}
	return out
}

func primitiveOpenAPI(name string) (openapiPrimitive, bool) {
	switch name {
	case "string":
		return openapiPrimitive{Type: "string"}, true
	case "int32":
		return openapiPrimitive{Type: "integer", Format: "int32"}, true
	case "int64":
		return openapiPrimitive{Type: "integer", Format: "int64"}, true
	case "float32":
		return openapiPrimitive{Type: "number", Format: "float"}, true
	case "float64":
		return openapiPrimitive{Type: "number", Format: "double"}, true
	case "bool":
		return openapiPrimitive{Type: "boolean"}, true
	case "file":
		return openapiPrimitive{Type: "string", Format: "binary"}, true
	}
	return openapiPrimitive{}, false
}

// yamlQuote single-quotes any string that could be misinterpreted as a non-string
// scalar by a YAML 1.2 parser. Anything containing a colon, quote, brace,
// bracket, leading dash, or whitespace at the boundaries gets quoted; pure
// identifiers are emitted bare. Single quotes in the input are doubled.
func yamlQuote(s string) string {
	if s == "" {
		return "''"
	}
	if needsYAMLQuote(s) {
		return "'" + strings.ReplaceAll(s, "'", "''") + "'"
	}
	return s
}

func needsYAMLQuote(s string) bool {
	if s != strings.TrimSpace(s) {
		return true
	}
	// Reserved scalar values.
	switch strings.ToLower(s) {
	case "true", "false", "null", "yes", "no", "on", "off", "~":
		return true
	}
	// First-character constraints.
	switch s[0] {
	case '-', '?', ':', ',', '[', ']', '{', '}', '#', '&', '*', '!', '|', '>', '\'', '"', '%', '@', '`':
		return true
	}
	for _, c := range s {
		if c == ':' || c == '#' || c == '\n' || c == '\t' {
			return true
		}
	}
	return false
}
