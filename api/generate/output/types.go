package output

import (
	"fmt"
	"strings"

	"github.com/medubin/gonzo/api/generate/data"
)

func Types(d *data.Data) string {
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

		"github.com/medubin/gonzo/api/utils/cookies"
		"github.com/medubin/gonzo/api/utils/handle"
		"github.com/medubin/gonzo/api/utils/router"
		"github.com/medubin/gonzo/api/utils/url"
	)
	`
}
func outputStructs(structs []*data.Struct) string {
	output := ""
	for _, s := range structs {
		structForm := fmt.Sprintf("type %s %s\n", s.Name, s.Type)
		structFunc := ""
		name := ""
		for _, f := range s.Fields {
			if name == "" {
				name = f
			} else {
				structForm += fmt.Sprintf("%s *%s\n", name, f)
				structFunc += fmt.Sprintf("func (s *%s) Get%s() *%s {\n ", s.Name, name, f)
				structFunc += "if (s == nil) {\n return nil\n}\n"
				structFunc += fmt.Sprintf("return s.%s\n}\n\n", name)
				name = ""
			}
		}
		structForm += "}\n\n"
		output += structForm
		output += structFunc
	}
	return output
}

func outputVariables(vs []*data.Variable) string {
	output := ""
	for _, v := range vs {
		output += fmt.Sprintf("type %s *%s\n\n\n", v.Name, v.Type)
	}
	return output
}

func outputServer(endpoints []*data.Endpoint) string {
	server := "type Server interface {\n"

	endpointList := []string{}
	for _, e := range endpoints {
		endpointList = append(endpointList, generateEndpoint(e))
	}

	server += strings.Join(endpointList, "\n")
	return server + "}"
}

func outputServerStart(endpoints []*data.Endpoint) string {
	serverInit := "func StartServer(s Server, r *router.Router) {"
	for _, e := range endpoints {
		serverInit += fmt.Sprintf("r.Route(\"%s\", \"%s\", handle.Handle(s.%s))\n", e.Verb, e.Url, e.Name)
	}

	return serverInit + "}"
}
