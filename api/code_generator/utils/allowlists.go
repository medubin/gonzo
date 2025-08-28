package utils

func GetLanguageStackConfig(language string, stack string) string {
	switch stack {
	case "server":
		switch language {
		case "go":
			return "api/code_generator/generator/languages/go/server/config.yaml"
		}
	case "client":
		switch language {
		case "typescript":
			return "api/code_generator/generator/languages/typescript/client/config.yaml"
		}
	}
	return ""
}
