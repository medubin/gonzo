package url

import (
	"regexp"
	"strings"
)

func ConvertPathToRegex(path string) *regexp.Regexp {
	path = strings.TrimRight(path, "/")
	// First convert all parameters that are not given an explicit match type. Give it \w+
	unspecifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+)}`)
	path = unspecifiedConversion.ReplaceAllString(path, `(?P<$1>\w+)`)

	specifiedConversion := regexp.MustCompile(`{([a-zA-Z0-9]+):(.+)}`)
	path = specifiedConversion.ReplaceAllString(path, `(?P<$1>$2)`)

	// NOTE: ^ means start of string and $ means end. Without these,
	//   we'll still match if the path has content before or after
	//   the expression (/foo/bar/baz would match the "/bar" route).
	return  regexp.MustCompile("^" + path + "/?$")
}

func GetKeys(path string) []string {
	keysRegex := regexp.MustCompile(`{([a-zA-Z0-9_]+):?.*?}`)

	// keysRegex := regexp.MustCompile(`{([a-zA-Z0-9_]+)}`)

	// return keysRegex.FindStringSubmatch(path)
	allMatches :=  keysRegex.FindAllStringSubmatch(path, -1)
	
	matches := make([]string, len(allMatches))

	for i, match := range allMatches {
		matches[i] = match[1] //ugh
	}

	return matches
}

type URL struct{
	params map[string]string
}

func (u *URL) Get(param string) string{
	return u.params[param]
}
