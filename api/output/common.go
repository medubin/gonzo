package output

import (
	"fmt"
	"strings"

	"github.com/medubin/gonzo/api/data"
)

func generateEndpoint(e *data.Endpoint) string {
	parameters := []string{"ctx context.Context"}

	if e.Body != "" {
		parameters = append(parameters, fmt.Sprintf("body *%s", e.Body))
	} else {
		parameters = append(parameters, fmt.Sprintf("body *%s", "interface{}"))
	}

	parameters = append(parameters, "cookie cookies.Cookies")
	parameters = append(parameters, fmt.Sprintf("url url.URL[%sUrl]", e.Name))
	returns := []string{}
	if e.Return != "" {
		returns = append(returns, "*"+e.Return)
	}

	// Can always return an error
	returns = append(returns, "error")

	return fmt.Sprintf("%s(%s) (%s)", e.Name, strings.Join(parameters, ", "), strings.Join(returns, ", "))
}
