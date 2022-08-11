package api

import (
	"bufio"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`(?mi)^(?P<t>body|returns)\((?P<v>[a-zA-Z]+)\)$`)

type Items struct {
	items    []string
	position int
}

func (i *Items) ValidPosition() bool {
	return i.position < len(i.items)
}

func (i *Items) Item() string {
	if i.ValidPosition() {
		return i.items[i.position]
	}
	panic("Error: Outside of valid position")
}

func (i *Items) Next() {
	i.position++
}

func (i *Items) NextItem() string {
	i.Next()
	return i.Item()
}

func (i *Items) PeekItem() string {
	if i.position+1 < len(i.items) {
		return i.items[i.position+1]
	}
	panic("Error: Peeked outside of Valid Position")
}

func InitItems(items []string) Items {
	return Items{
		items:    items,
		position: 0,
	}
}

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
	if len(v.Fields) > 0 {
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
	}

	parameters = append(parameters, "cookie router.Cookies")
	returns := []string{}
	if e.Return != "" {
		returns = append(returns, "*" + e.Return)
	}

	// Can always return an error
	returns = append(returns, "error")
	// return fmt.Sprintf("r.Route(\"%s\", \"%s\", router.Handle(server.%s))", e.Verb, e.Url, e.Name)
	return fmt.Sprintf("%s(%s) (%s)\n", e.Name, strings.Join(parameters, ", "), strings.Join(returns, ", "))
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

		"github.com/medubin/gonzo/router"
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
		serverInit += fmt.Sprintf("r.Route(\"%s\", \"%s\", router.Handle(s.%s))\n", e.Verb, e.Url, e.Name)
	}
	server += "}"
	serverInit += "}"

	return fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s", o.Header(), variables, server, serverInit)
}

func GenerateAPI(name string) (string, error) {
	i, err := ParseFile(name + ".api")
	if err != nil {
		return "", err
	}
	items := InitItems(i)
	output := GenerateOutput(items)

	formattedOutput, err := format.Source([]byte(output))
	if err != nil {
		return "", err
	}

	return string(formattedOutput), nil
}

func WriteToFile(name string, output string) error {
	return os.WriteFile("../server/" + name + ".go", []byte(output), 0644)
}

func ParseFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	return itemize(scanner)
}

func itemize(s *bufio.Scanner) ([]string, error) {
	var items []string
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if len(line) == 0 {
			continue
		}
		subItems := strings.Split(line, " ")
		items = append(items, subItems...)

		if err := s.Err(); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func GenerateOutput(items Items) string {
	output := Output{}

	for items.ValidPosition() {
		item := items.Item()

		if item == "}" {
			output.FinishVariable()

			// Lines that start with "type" should always have 3 items
		} else if item == "type" {
			name := items.NextItem()
			typeName := items.NextItem()
			if typeName == "{" {
				output.InStruct = true
				output.AddStruct(name)
			} else {
				output.AddAlias(name, typeName)
				output.FinishVariable()
			}
			// fields in structs should have 2 items
		} else if output.InStruct {
			typeName := items.NextItem()

			output.AddStructField(item, typeName)
			// Server setup contains 2 items
		} else if item == "Server" {
			output.InServer = true
			items.Next()
			// Server can have a variable amount of items
			// The first 3 are HTTP verb (POST, GET, PUT, PATCH, and DELETE), url, and function name
			// followed by key value pairs in the form of name(type)
			// current types: body, setcookie, header, returns, getcookie
			// example: body(TestBody)
			// TODO allow for more then body and return
		} else if output.InServer {
			e := Endpoint{}
			e.Verb = item
			e.Url = items.NextItem()
			e.Name = items.NextItem()

			for items.ValidPosition() {
				match := re.FindStringSubmatch(items.PeekItem())
				if match == nil {
					break
				}
				parName := match[1]
				parType := match[2]
				if parName == "returns" {
					e.Return = parType
				} else if parName == "body" {
					e.Body = parType
				} else {
					panic(fmt.Sprintf("Invalid type: %s", parName))
				}
				items.Next()
			}
			output.AddEndpoint(e)

		} else {
			panic("AHHH")
		}
		items.Next()
	}
	return output.String()
}
