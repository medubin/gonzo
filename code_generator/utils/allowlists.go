package utils

func GetLanguageStackConfig(language string, stack string) string {
	switch stack {
	case "server":
		switch language {
		case "go":
			return "code_generator/generator/languages/go/server/config.yaml"
		}
	case "client":
		switch language {
		case "typescript":
			return "code_generator/generator/languages/typescript/client/config.yaml"
		}
	case "spec":
		switch language {
		case "openapi":
			return "code_generator/generator/languages/openapi/spec/config.yaml"
		}
	}
	return ""
}
