package api

import (
	"fmt"
	"strings"
)

func Output(d Data) string {
	header := outputHeader()
	variables := outputVariables(d.Variables)
	structs := outputStructs(d.Structs)
	server := outputServer(d.Endpoints)
	serverStart := outputServerStart(d.Endpoints)

	return strings.Join([]string{
		header, variables, structs, server, serverStart,
	}, "\n\n")
}

func outputHeader() string {
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
func outputStructs(structs []*Struct) string {
	output := ""
	for _, s := range structs {
		output += fmt.Sprintf("type %s %s\n", s.Name, s.Type)
		for i, f := range s.Fields {
			output += f
			if i%2 != 0 {
				output += "\n"
			} else {
				output += " "
			}
		}
		output += "}\n\n"
	}
	return output
}

func outputVariables(vs []*Variable) string {
	output := ""
	for _, v := range vs {
		output += fmt.Sprintf("type %s %s\n\n\n", v.Name, v.Type)
	}
	return output
}

func outputServer(endpoints []*Endpoint) string {
	server := "type Server interface {\n"
	for _, e := range endpoints {
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
		server += fmt.Sprintf("%s(%s) (%s)\n", e.Name, strings.Join(parameters, ", "), strings.Join(returns, ", "))
	}

	return server + "}"
}

func outputServerStart(endpoints []*Endpoint) string {
	serverInit := "func StartServer(s Server, r *router.Router) {"
	for _, e := range endpoints {
		serverInit += fmt.Sprintf("r.Route(\"%s\", \"%s\", handle.Handle(s.%s))\n", e.Verb, e.Url, e.Name)
	}

	return serverInit + "}"
}
