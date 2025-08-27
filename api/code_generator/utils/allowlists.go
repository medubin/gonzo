package utils

func IsLanguageStackAllowed(language string, stack string) bool {
	switch stack {
	case "server":
		switch language {
		case "go":
			return true
		}
	case "client":
		switch language {
		case "typescript":
			return true
		}
	}
	return false
}