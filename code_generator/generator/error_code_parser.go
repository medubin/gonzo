package generator

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strings"
)

// TemplateErrorCode represents a parsed error code for use in code generation templates.
type TemplateErrorCode struct {
	ClassName  string // e.g., "NotFoundError"
	Code       string // e.g., "not_found"
	StatusCode int    // e.g., 404
}

// parseErrorCodesFromFile reads a Go source file and extracts error code definitions.
// It finds all exported functions that return GonzoError and extracts their code string
// and HTTP status code by parsing the newError(...) call inside each function.
func parseErrorCodesFromFile(filePath string) ([]TemplateErrorCode, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, filePath, src, 0)
	if err != nil {
		return nil, err
	}

	// Build map: constant name → string value, e.g. "NotFound" → "not_found"
	codeValues := make(map[string]string)
	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.CONST {
			continue
		}
		for _, spec := range genDecl.Specs {
			valueSpec, ok := spec.(*ast.ValueSpec)
			if !ok || valueSpec.Type == nil {
				continue
			}
			ident, ok := valueSpec.Type.(*ast.Ident)
			if !ok || ident.Name != "ErrorCode" {
				continue
			}
			for i, name := range valueSpec.Names {
				if i < len(valueSpec.Values) {
					if lit, ok := valueSpec.Values[i].(*ast.BasicLit); ok {
						codeValues[name.Name] = strings.Trim(lit.Value, `"`)
					}
				}
			}
		}
	}

	// Find exported functions returning GonzoError and extract their code/status.
	var errorCodes []TemplateErrorCode
	for _, decl := range f.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok || funcDecl.Name == nil || !funcDecl.Name.IsExported() {
			continue
		}
		if !funcReturnsType(funcDecl, "GonzoError") {
			continue
		}

		codeName, statusCode := extractNewErrorCall(funcDecl)
		if codeName == "" {
			continue
		}
		codeValue, ok := codeValues[codeName]
		if !ok {
			continue
		}

		errorCodes = append(errorCodes, TemplateErrorCode{
			ClassName:  funcDecl.Name.Name,
			Code:       codeValue,
			StatusCode: statusCode,
		})
	}

	return errorCodes, nil
}

func funcReturnsType(funcDecl *ast.FuncDecl, typeName string) bool {
	if funcDecl.Type.Results == nil {
		return false
	}
	for _, field := range funcDecl.Type.Results.List {
		if ident, ok := field.Type.(*ast.Ident); ok && ident.Name == typeName {
			return true
		}
	}
	return false
}

// extractNewErrorCall walks the function body looking for a newError(code, msg, http.StatusXxx)
// call and returns the code constant name and the resolved HTTP status code integer.
func extractNewErrorCall(funcDecl *ast.FuncDecl) (codeName string, statusCode int) {
	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		callExpr, ok := n.(*ast.CallExpr)
		if !ok {
			return true
		}
		ident, ok := callExpr.Fun.(*ast.Ident)
		if !ok || ident.Name != "newError" {
			return true
		}
		if len(callExpr.Args) < 3 {
			return true
		}
		if codeIdent, ok := callExpr.Args[0].(*ast.Ident); ok {
			codeName = codeIdent.Name
		}
		if selExpr, ok := callExpr.Args[2].(*ast.SelectorExpr); ok {
			statusCode = httpStatusNameToCode(selExpr.Sel.Name)
		}
		return false
	})
	return
}

var httpStatusNames = map[string]int{
	"StatusContinue":           100,
	"StatusSwitchingProtocols": 101,
	"StatusOK":                 200,
	"StatusCreated":            201,
	"StatusAccepted":           202,
	"StatusNoContent":          204,
	"StatusMovedPermanently":   301,
	"StatusFound":              302,
	"StatusNotModified":        304,
	"StatusBadRequest":         400,
	"StatusUnauthorized":       401,
	"StatusForbidden":          403,
	"StatusNotFound":           404,
	"StatusMethodNotAllowed":   405,
	"StatusConflict":           409,
	"StatusGone":               410,
	"StatusUnprocessableEntity": 422,
	"StatusTooManyRequests":    429,
	"StatusInternalServerError": 500,
	"StatusNotImplemented":     501,
	"StatusBadGateway":         502,
	"StatusServiceUnavailable": 503,
}

func httpStatusNameToCode(name string) int {
	if code, ok := httpStatusNames[name]; ok {
		return code
	}
	return 500
}
