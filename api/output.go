package api

import (
	"fmt"
	"strings"

	"github.com/medubin/gonzo/utils/url"
)

type Output struct {
	InStruct        bool
	InServer        bool
	variables       []Variable
	currentVariable *Variable
	endpoints       []Endpoint
}

type Endpoint struct {
	Verb   string
	Url    string
	Name   string
	Body   string
	Return string
}

type Variable struct {
	Name   string
	Type   string
	Fields []string
}

func (o *Output) AddStruct(name string) {
	if o.currentVariable != nil {
		panic("Current Variable not empty")
	}
	o.currentVariable = &Variable{
		Name: name,
		Type: "struct {",
	}
}

func (o *Output) AddAlias(name string, typeName string) {
	if o.currentVariable != nil {
		panic("Current Variable not empty")
	}
	o.currentVariable = &Variable{
		Name: name,
		Type: typeName,
	}
}

func (o *Output) AddStructField(name string, typeName string) {
	if o.currentVariable == nil {
		panic(fmt.Sprintf("Current Variable empty. Tried adding %s %s", name, typeName))
	}
	o.currentVariable.Fields = append(o.currentVariable.Fields, name, typeName)
}

func (o *Output) AddEndpoint(e Endpoint) {
	o.endpoints = append(o.endpoints, e)

	matches := url.GetKeys(e.Url)
	fields := make([]string, len(matches) * 2)
	for i, match := range matches {
		fields[i * 2] = match
		fields[i * 2 + 1] = "string"
	}

	o.variables = append(o.variables, Variable{
		Name: e.Name + "Url",
		Type: "struct {",
		Fields: fields,
	})
}

func (o *Output) FinishVariable() {
	o.InStruct = false
	o.InServer = false
	if o.currentVariable != nil {
		o.variables = append(o.variables, *o.currentVariable)
		o.currentVariable = nil
	}
}

func (o *Output) Header() string {
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

func (o *Output) String() string {
	if o.currentVariable != nil {
		panic("Ended run with current variable still full")
	}

	variables := ""
	for _, v := range o.variables {
		variables += v.String()
	}

	server := "type Server interface {\n"
	serverInit := "func StartServer(s Server, r *router.Router) {"
	for _, e := range o.endpoints {
		server += e.String()
		serverInit += fmt.Sprintf("r.Route(\"%s\", \"%s\", handle.Handle(s.%s))\n", e.Verb, e.Url, e.Name)
	}
	server += "}"
	serverInit += "}"

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", o.Header(), variables, server, serverInit)
}

func (v *Variable) String() string {
	variable := fmt.Sprintf("type %s %s\n", v.Name, v.Type)
	for i, f := range v.Fields {
		variable += f
		if i%2 != 0 {
			variable += "\n"
		} else {
			variable += " "
		}
	}
	if v.Type == "struct {" {
		variable += "}\n\n"
	} else {
		variable += "\n\n"
	}
	return variable

}

func (e *Endpoint) String() string {
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
