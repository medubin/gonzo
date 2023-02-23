package utils

import "regexp"

func GetKeys(path string) []string {
	keysRegex := regexp.MustCompile(`{([a-zA-Z0-9_]+):?.*?}`)

	// return keysRegex.FindStringSubmatch(path)
	allMatches := keysRegex.FindAllStringSubmatch(path, -1)

	matches := make([]string, len(allMatches))

	for i, match := range allMatches {
		matches[i] = match[1] //ugh
	}

	return matches
}
