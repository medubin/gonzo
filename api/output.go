package api

import (
	"fmt"
	"strings"
)

func Output(d Data) string {
	head := header()

	variables := outputVariables(d.Variables)

	structs := ""
	for _, v := range d.Structs {
		structs += outputStruct(v)
	}

	server := "type Server interface {\n"
	serverInit := "func StartServer(s Server, r *router.Router) {"
	for _, e := range d.Endpoints {
		server += endpoint(e)
		serverInit += fmt.Sprintf("r.Route(\"%s\", \"%s\", handle.Handle(s.%s))\n", e.Verb, e.Url, e.Name)
	}
	server += "}"
	serverInit += "}"

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s\n\n%s", head, variables,  structs, server, serverInit)
}

func header() string {
	return `package server
	import (
		"context"

		"github.com/medubin/gonzo/utils/cookies"
		"github.com/medubin/gonzo/utils/handle"
		"github.com/medubin/gonzo/utils/router"
		"github.com/medubin/gonzo/utils/url"
	)
	`
}
func outputStruct(s *Struct) string {
	variable := fmt.Sprintf("type %s %s\n", s.Name, s.Type)
	for i, f := range s.Fields {
		variable += f
		if i%2 != 0 {
			variable += "\n"
		} else {
			variable += " "
		}
	}
	if s.Type == "struct {" {
		variable += "}\n\n"
	} else {
		variable += "\n\n"
	}
	return variable
}

func outputVariables(vs []Variable) string {
	output := ""
	for _, v := range vs {
		output += fmt.Sprintf("type %s %s\n\n\n", v.Name, v.Type)
	}
	return output
}

func endpoint(e Endpoint) string {
	parameters := []string{"ctx context.Context"}

	if e.Body != "" {
		parameters = append(parameters, fmt.Sprintf("body %s", e.Body))
	} else {
		parameters = append(parameters, fmt.Sprintf("body %s", "interface{}"))
	}

	parameters = append(parameters, "cookie cookies.Cookies")
	parameters = append(parameters, fmt.Sprintf("url url.URL[%sUrl]", e.Name))
	returns := []string{}
	if e.Return != "" {
		returns = append(returns, "*"+e.Return)
	}

	// Can always return an error
	returns = append(returns, "error")
	// return fmt.Sprintf("r.Route(\"%s\", \"%s\", router.Handle(server.%s))", e.Verb, e.Url, e.Name)
	return fmt.Sprintf("%s(%s) (%s)\n", e.Name, strings.Join(parameters, ", "), strings.Join(returns, ", "))
}
