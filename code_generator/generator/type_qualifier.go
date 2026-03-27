package generator

import "regexp"

// upperIdentRegexp matches exported Go identifiers (uppercase-leading, word boundary).
var upperIdentRegexp = regexp.MustCompile(`\b([A-Z][A-Za-z0-9]*)\b`)

// qualifyTypeString prepends alias. to all exported identifiers in typeStr.
// Used in sub-package templates to qualify types from the parent package.
// If alias is empty, returns typeStr unchanged.
func qualifyTypeString(typeStr, alias string) string {
	if alias == "" {
		return typeStr
	}
	return upperIdentRegexp.ReplaceAllString(typeStr, alias+".$1")
}
