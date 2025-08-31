package generator

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/iancoleman/strcase"
	"gopkg.in/yaml.v2"
)

// LanguageSettings holds language-specific settings
type LanguageSettings struct {
	PackageComment  string   `yaml:"package_comment"`
	TypesImports    []string `yaml:"types_imports"`
	EndpointImports []string `yaml:"endpoint_imports"`
}

// LanguageConfig defines language-specific configuration
type LanguageConfig struct {
	Language     string            `yaml:"language"`
	FileExt      string            `yaml:"file_ext"`
	Primitives   map[string]string `yaml:"primitives"`
	TypePatterns map[string]string `yaml:"type_patterns"`
	Templates    map[string]string `yaml:"templates"`
	Imports      []string          `yaml:"default_imports"`
	Settings     LanguageSettings  `yaml:"settings"`
}

// TemplateData represents the data passed to templates
type TemplateData struct {
	PackageName string
	Language    string
	Imports     []string
	Types       []TemplateType
	Enums       []TemplateEnum
	Servers     []TemplateServer
	Settings    LanguageSettings
}

// TemplateComment represents a comment with its type for templates
type TemplateComment struct {
	Content string
	Type    string // "single" or "multi"
}

type TemplateType struct {
	Name       string
	Kind       string
	Target     string // for aliases
	Fields     []TemplateField
	Comments   []TemplateComment
	MappedType string // The complete mapped type string
}

type TemplateField struct {
	Name     string
	Type     string
	Required bool
	Comments []TemplateComment
	JSONTag  string
}

type TemplateEnum struct {
	Name     string
	BaseType string
	Values   []TemplateEnumValue
	Comments []TemplateComment
}

type TemplateEnumValue struct {
	Key   string
	Value string
}

type TemplateServer struct {
	Name      string
	Endpoints []TemplateEndpoint
	Comments  []TemplateComment
}

type TemplateEndpoint struct {
	Name       string
	Method     string
	Path       string
	PathParams []TemplatePathParam
	Parameters string
	BodyType   string
	ReturnType string
	URLType    string
	Comments   []TemplateComment
	HasBody    bool
	HasReturn  bool
	HasParams  bool
}

type TemplatePathParam struct {
	Name   string
	Type   string
}

// TemplateGenerator generates code using templates
type TemplateGenerator struct {
	config    LanguageConfig
	templates map[string]*template.Template
	funcMap   template.FuncMap
}

// NewTemplateGenerator creates a new template generator
func NewTemplateGenerator(configPath string) (*TemplateGenerator, error) {
	config, err := loadLanguageConfig(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	tg := &TemplateGenerator{
		config:    config,
		templates: make(map[string]*template.Template),
		funcMap:   make(template.FuncMap),
	}

	// Add template functions
	tg.setupTemplateFunctions()

	// Load templates
	if err := tg.loadTemplates(); err != nil {
		return nil, fmt.Errorf("failed to load templates: %v", err)
	}

	return tg, nil
}

// setupTemplateFunctions adds helper functions for templates
func (tg *TemplateGenerator) setupTemplateFunctions() {
	tg.funcMap["mapType"] = tg.mapType
	tg.funcMap["mapTypeExpr"] = tg.mapTypeExpr
	tg.funcMap["capitalize"] = tg.capitalize
	tg.funcMap["lower"] = strings.ToLower
	tg.funcMap["upper"] = strings.ToUpper
	tg.funcMap["camelCase"] = strcase.ToLowerCamel
	tg.funcMap["join"] = strings.Join
	tg.funcMap["hasPrefix"] = strings.HasPrefix
	tg.funcMap["quote"] = func(s string) string {
		escaped := strings.ReplaceAll(s, `"`, `\"`)
		return fmt.Sprintf(`"%s"`, escaped)
	}
	tg.funcMap["formatComment"] = tg.formatComment
	tg.funcMap["formatComments"] = tg.formatTemplateComments
	tg.funcMap["indent"] = func(tabCount int, text string) string {
		if text == "" {
			return ""
		}
		indent := strings.Repeat("\t", tabCount)
		lines := strings.Split(text, "\n")
		for i, line := range lines {
			if line != "" {
				lines[i] = indent + line
			}
		}
		return strings.Join(lines, "\n")
	}
	tg.funcMap["replace"] = func(old, new, s string) string {
		return strings.ReplaceAll(s, old, new)
	}
	tg.funcMap["getUsedTypes"] = tg.getUsedTypes
	tg.funcMap["handleUnknownType"] = tg.handleUnknownType
}

// loadLanguageConfig loads language configuration from YAML
func loadLanguageConfig(configPath string) (LanguageConfig, error) {
	var config LanguageConfig
	data, err := os.ReadFile(configPath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(data, &config)
	return config, err
}

// loadTemplates loads all templates from the config
func (tg *TemplateGenerator) loadTemplates() error {
	for name, content := range tg.config.Templates {
		tmpl, err := template.New(name).Funcs(tg.funcMap).Parse(content)
		if err != nil {
			return fmt.Errorf("failed to parse template %s: %v", name, err)
		}
		tg.templates[name] = tmpl
	}
	return nil
}

// Generate generates code from API definition
func (tg *TemplateGenerator) Generate(api *APIDefinition, packageName string) (map[string]string, error) {
	data := tg.prepareTemplateData(api, packageName)

	files := make(map[string]string)

	for templateName, tmpl := range tg.templates {
		// Handle special case for endpoint template - generate one file per endpoint
		if templateName == "endpoint" {
			endpointFiles, err := tg.generateEndpointFiles(tmpl, data)
			if err != nil {
				return nil, fmt.Errorf("failed to generate endpoint files: %v", err)
			}
			// Merge endpoint files into main files map
			for filename, content := range endpointFiles {
				files[filename] = content
			}
		} else {
			// Normal template processing
			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, data); err != nil {
				return nil, fmt.Errorf("failed to execute template %s: %v", templateName, err)
			}

			filename := templateName + tg.config.FileExt
			files[filename] = buf.String()
		}
	}

	return files, nil
}

// generateEndpointFiles generates individual files for each endpoint
func (tg *TemplateGenerator) generateEndpointFiles(tmpl *template.Template, data TemplateData) (map[string]string, error) {
	files := make(map[string]string)

	for _, server := range data.Servers {
		for _, endpoint := range server.Endpoints {
			// Create data for single endpoint
			endpointData := struct {
				TemplateData
				Server   TemplateServer
				Endpoint TemplateEndpoint
			}{
				TemplateData: data,
				Server:       server,
				Endpoint:     endpoint,
			}

			var buf bytes.Buffer
			if err := tmpl.Execute(&buf, endpointData); err != nil {
				return nil, fmt.Errorf("failed to execute endpoint template for %s: %v", endpoint.Name, err)
			}

			// Use endpoint name for filename (snake case)
			filename := fmt.Sprintf("%s%s", strcase.ToSnake(endpoint.Name), tg.config.FileExt)
			files[filename] = buf.String()
		}
	}

	return files, nil
}

// prepareTemplateData converts API definition to template data
func (tg *TemplateGenerator) prepareTemplateData(api *APIDefinition, packageName string) TemplateData {
	data := TemplateData{
		PackageName: packageName,
		Language:    tg.config.Language,
		Imports:     tg.config.Imports,
		Settings:    tg.config.Settings,
	}

	// Convert types
	for _, typeDef := range api.Types {
		data.Types = append(data.Types, tg.convertType(&typeDef))
	}

	// Convert enums
	for _, enumDef := range api.Enums {
		data.Enums = append(data.Enums, tg.convertEnum(&enumDef))
	}

	// Convert servers
	for _, serverDef := range api.Servers {
		data.Servers = append(data.Servers, tg.convertServer(&serverDef))
	}

	return data
}

// convertType converts TypeDef to TemplateType
func (tg *TemplateGenerator) convertType(typeDef *TypeDef) TemplateType {
	tt := TemplateType{
		Name:     typeDef.Name,
		Kind:     typeDef.Kind,
		Target:   tg.mapType(typeDef.Target),
		Comments: tg.extractComments(typeDef.Comments),
	}

	// Convert fields for structs
	for _, field := range typeDef.Fields {
		tt.Fields = append(tt.Fields, tg.convertField(&field))
	}

	// Set MappedType based on kind
	switch typeDef.Kind {
	case "alias":
		tt.MappedType = tg.mapType(typeDef.Target)
	case "repeated":
		if typeDef.ElementType != nil {
			repeatedTypeExpr := &TypeExpr{
				Kind:        "repeated",
				ElementType: typeDef.ElementType,
			}
			tt.MappedType = tg.mapTypeExpr(repeatedTypeExpr)
		}
	case "map":
		if typeDef.KeyType != nil && typeDef.ValueType != nil {
			mapTypeExpr := &TypeExpr{
				Kind:      "map",
				KeyType:   typeDef.KeyType,
				ValueType: typeDef.ValueType,
			}
			tt.MappedType = tg.mapTypeExpr(mapTypeExpr)
		}
	case "struct":
		tt.MappedType = "struct"
	}

	return tt
}

// convertField converts FieldDef to TemplateField
func (tg *TemplateGenerator) convertField(field *FieldDef) TemplateField {
	jsonTag := strings.ToLower(field.Name)
	if !field.Required {
		jsonTag += ",omitempty"
	}

	return TemplateField{
		Name:     field.Name,
		Type:     tg.mapTypeExpr(field.Type),
		Required: field.Required,
		Comments: tg.extractComments(field.Comments),
		JSONTag:  jsonTag,
	}
}

// convertEnum converts EnumDef to TemplateEnum
func (tg *TemplateGenerator) convertEnum(enumDef *EnumDef) TemplateEnum {
	te := TemplateEnum{
		Name:     enumDef.Name,
		BaseType: enumDef.BaseType,
		Comments: tg.extractComments(enumDef.Comments),
	}

	// Sort keys for consistent ordering
	var keys []string
	for key := range enumDef.Values {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	// Add values in sorted order
	for _, key := range keys {
		te.Values = append(te.Values, TemplateEnumValue{
			Key:   key,
			Value: enumDef.Values[key],
		})
	}

	return te
}

// convertServer converts ServerDef to TemplateServer
func (tg *TemplateGenerator) convertServer(serverDef *ServerDef) TemplateServer {
	ts := TemplateServer{
		Name:     serverDef.Name,
		Comments: tg.extractComments(serverDef.Comments),
	}

	for _, endpoint := range serverDef.Endpoints {
		ts.Endpoints = append(ts.Endpoints, tg.convertEndpoint(&endpoint))
	}

	return ts
}

// convertEndpoint converts EndpointDef to TemplateEndpoint
func (tg *TemplateGenerator) convertEndpoint(endpoint *EndpointDef) TemplateEndpoint {
	te := TemplateEndpoint{
		Name:     endpoint.Name,
		Method:   endpoint.Method,
		Path:     endpoint.Path,
		Comments: tg.extractComments(endpoint.Comments),
	}

	// Convert path parameters
	for _, param := range endpoint.PathParams {
		te.PathParams = append(te.PathParams, TemplatePathParam{
			Name:   param.Name,
			Type:   tg.mapType(param.Type),
		})
	}

	// Set body type
	if endpoint.Body != nil {
		te.BodyType = fmt.Sprintf("*%s", tg.mapTypeExpr(endpoint.Body))
		te.HasBody = true
	} else {
		te.BodyType = fmt.Sprintf("*%s", tg.getDefaultType())
	}

	// Set return type
	if endpoint.Returns != nil {
		te.ReturnType = fmt.Sprintf("*%s", tg.mapTypeExpr(endpoint.Returns))
		te.HasReturn = true
	} else {
		te.ReturnType = fmt.Sprintf("*%s", tg.getDefaultType())
	}

	// Set URL type
	if len(endpoint.PathParams) > 0 {
		te.URLType = fmt.Sprintf("%sUrl", endpoint.Name)
	} else {
		te.URLType = tg.getDefaultType()
	}

	// Check for parameters
	te.HasParams = endpoint.Parameters != nil
	if endpoint.Parameters != nil {
		te.Parameters = tg.mapTypeExpr(endpoint.Parameters)
	}

	return te
}

// Helper functions
func (tg *TemplateGenerator) mapType(typeName string) string {
	if mapped, ok := tg.config.Primitives[typeName]; ok {
		return mapped
	}
	return tg.capitalize(typeName)
}

func (tg *TemplateGenerator) mapTypeExpr(typeExpr *TypeExpr) string {
	if typeExpr == nil {
		return tg.getDefaultType()
	}

	switch typeExpr.Kind {
	case "reference":
		return tg.mapType(typeExpr.Name)
	case "repeated":
		if pattern, ok := tg.config.TypePatterns["repeated"]; ok {
			return fmt.Sprintf(pattern, tg.mapTypeExpr(typeExpr.ElementType))
		}
		return tg.getRepeatedTypeFallback(tg.mapTypeExpr(typeExpr.ElementType))
	case "map":
		if pattern, ok := tg.config.TypePatterns["map"]; ok {
			return fmt.Sprintf(pattern, tg.mapTypeExpr(typeExpr.KeyType), tg.mapTypeExpr(typeExpr.ValueType))
		}
		return tg.getMapTypeFallback(tg.mapTypeExpr(typeExpr.KeyType), tg.mapTypeExpr(typeExpr.ValueType))
	default:
		return tg.getDefaultType()
	}
}

// getDefaultType returns the default type for unknown/null types
func (tg *TemplateGenerator) getDefaultType() string {
	if defaultPattern, ok := tg.config.TypePatterns["default"]; ok {
		return defaultPattern
	}
	// Language-agnostic fallback - let templates handle this
	return "UNKNOWN_TYPE"
}

// getRepeatedTypeFallback returns a language-agnostic fallback for repeated types
func (tg *TemplateGenerator) getRepeatedTypeFallback(elementType string) string {
	// Language-agnostic fallback - let templates handle this
	return fmt.Sprintf("REPEATED_OF_%s", elementType)
}

// getMapTypeFallback returns a language-agnostic fallback for map types
func (tg *TemplateGenerator) getMapTypeFallback(keyType, valueType string) string {
	// Language-agnostic fallback - let templates handle this
	return fmt.Sprintf("MAP_OF_%s_TO_%s", keyType, valueType)
}

func (tg *TemplateGenerator) capitalize(s string) string {
	if len(s) == 0 {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func (tg *TemplateGenerator) extractComments(comments []Comment) []TemplateComment {
	var result []TemplateComment
	for _, comment := range comments {
		result = append(result, TemplateComment{
			Content: comment.Content,
			Type:    comment.Type,
		})
	}
	return result
}

func (tg *TemplateGenerator) formatComment(comment string) string {
	return fmt.Sprintf("// %s", comment)
}

func (tg *TemplateGenerator) formatTemplateComments(comments []TemplateComment) string {
	if len(comments) == 0 {
		return ""
	}

	var result []string
	for _, comment := range comments {
		if comment.Type == "multi" {
			// Format as multiline comment
			lines := strings.Split(comment.Content, "\n")
			if len(lines) == 1 {
				// Single line multiline comment
				result = append(result, fmt.Sprintf("/* %s */", comment.Content))
			} else {
				// True multiline comment
				result = append(result, "/*")
				for _, line := range lines {
					trimmed := strings.TrimSpace(line)
					if trimmed != "" {
						result = append(result, fmt.Sprintf(" * %s", trimmed))
					} else {
						result = append(result, " *")
					}
				}
				result = append(result, " */")
			}
		} else {
			// Single line comment
			result = append(result, fmt.Sprintf("// %s", comment.Content))
		}
	}
	return strings.Join(result, "\n")
}

// GenerateFromJSONWithTemplate generates code using templates
// This is the main entry point function you would call
func GenerateFromJSONWithTemplate(jsonData []byte, packageName, configPath string) (map[string]string, error) {
	var apiDef APIDefinition
	if err := json.Unmarshal(jsonData, &apiDef); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %v", err)
	}

	generator, err := NewTemplateGenerator(configPath)
	if err != nil {
		return nil, err
	}

	return generator.Generate(&apiDef, packageName)
}

// getUsedTypes analyzes template data and returns only the types that are actually used
func (tg *TemplateGenerator) getUsedTypes(data TemplateData) []TemplateType {
	usedTypeNames := make(map[string]bool)
	
	// Collect all type names used in servers/endpoints
	for _, server := range data.Servers {
		for _, endpoint := range server.Endpoints {
			// Check return type
			if endpoint.ReturnType != "" {
				tg.extractTypeNames(endpoint.ReturnType, usedTypeNames, data)
			}
			// Check body type
			if endpoint.BodyType != "" {
				tg.extractTypeNames(endpoint.BodyType, usedTypeNames, data)
			}
			// Check parameters
			if endpoint.Parameters != "" {
				tg.extractTypeNames(endpoint.Parameters, usedTypeNames, data)
			}
			// Add path param interface types (e.g., UpdateUserParams) 
			// but don't scan their field types since those are handled in the types file
			if len(endpoint.PathParams) > 0 {
				paramTypeName := endpoint.Name + "Params"
				usedTypeNames[paramTypeName] = true
			}
		}
	}
	
	// Filter types to only include used ones, plus add generated param types
	var usedTypes []TemplateType
	for _, typ := range data.Types {
		if usedTypeNames[typ.Name] {
			usedTypes = append(usedTypes, typ)
		}
	}
	
	// Add dynamically generated parameter interface types
	for _, server := range data.Servers {
		for _, endpoint := range server.Endpoints {
			if len(endpoint.PathParams) > 0 {
				paramTypeName := endpoint.Name + "Params"
				if usedTypeNames[paramTypeName] {
					// Create a synthetic TemplateType for the param interface
					paramType := TemplateType{
						Name: paramTypeName,
						Kind: "interface",
					}
					usedTypes = append(usedTypes, paramType)
				}
			}
		}
	}
	
	return usedTypes
}

// extractTypeNames extracts type names from a type string
func (tg *TemplateGenerator) extractTypeNames(typeStr string, usedNames map[string]bool, data TemplateData) {
	// Remove common decorators
	cleaned := strings.ReplaceAll(typeStr, "*", "")
	cleaned = strings.ReplaceAll(cleaned, "Array<", "")
	cleaned = strings.ReplaceAll(cleaned, "Record<", "")
	cleaned = strings.ReplaceAll(cleaned, ">", "")
	cleaned = strings.ReplaceAll(cleaned, ",", " ")
	
	// Split by common separators and check each part
	parts := strings.Fields(cleaned)
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" || tg.isPrimitive(part) {
			continue
		}
		
		// Check if it's one of our defined types
		for _, typ := range data.Types {
			if typ.Name == part {
				usedNames[part] = true
				// Don't recursively check fields - in TypeScript, importing a type 
				// automatically makes its field types available
				break
			}
		}
	}
}


// isPrimitive checks if a type is a primitive type
func (tg *TemplateGenerator) isPrimitive(typeName string) bool {
	// Check if it's defined in the language's primitive mappings
	_, isPrimitive := tg.config.Primitives[typeName]
	if isPrimitive {
		return true
	}
	
	// Check for default type pattern (which would be language-specific)
	if defaultType, ok := tg.config.TypePatterns["default"]; ok && typeName == defaultType {
		return true
	}
	
	// Common language-agnostic primitives that should be recognized
	commonPrimitives := map[string]bool{
		"void":      true,
		"undefined": true,
		"null":      true,
	}
	return commonPrimitives[typeName]
}

// handleUnknownType processes unknown type markers and converts them to appropriate language types
func (tg *TemplateGenerator) handleUnknownType(typeStr string) string {
	// Convert fallback markers to actual language types
	switch {
	case strings.HasPrefix(typeStr, "REPEATED_OF_"):
		elementType := strings.TrimPrefix(typeStr, "REPEATED_OF_")
		elementType = tg.handleUnknownType(elementType) // recursive for nested unknowns
		return tg.getRepeatedTypeFallback(elementType)
	case strings.HasPrefix(typeStr, "MAP_OF_") && strings.Contains(typeStr, "_TO_"):
		parts := strings.Split(strings.TrimPrefix(typeStr, "MAP_OF_"), "_TO_")
		if len(parts) == 2 {
			keyType := tg.handleUnknownType(parts[0])
			valueType := tg.handleUnknownType(parts[1])
			return tg.getMapTypeFallback(keyType, valueType)
		}
	case typeStr == "UNKNOWN_TYPE":
		return tg.getDefaultType()
	}
	return typeStr
}

// SaveGeneratedFiles saves generated files to disk
func SaveGeneratedFiles(files map[string]string, outputDir string) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return err
	}

	for filename, content := range files {
		filepath := filepath.Join(outputDir, filename)
		if err := os.WriteFile(filepath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %v", filename, err)
		}
	}

	return nil
}
