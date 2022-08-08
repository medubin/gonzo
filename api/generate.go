package api

import (
	"bufio"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`(?mi)^(?P<t>body|header|getcookie|returns|setcookie)\((?P<v>[a-zA-Z]+)\)$`)

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
	variables       []string
	currentVariable string
	server          []string
}

type CurrentServer struct {
	Verb      string
	Url       string
	Name      string
	Body      string
	Header    string
	SetCookie string
	Return    string
	GetCookie string
}

func (o *Output) Add(line string) {
	o.currentVariable += line
}

func (o *Output) AddStruct(name string) {
	o.Add(fmt.Sprintf("type %s struct {\n", name))
}

func (o *Output) AddAlias(name string, typeName string) {
	o.Add(fmt.Sprintf("type %s %s\n", name, typeName))
}

func (o *Output) AddStructField(name string, typeName string) {
	o.Add(fmt.Sprintf("%s %s\n", name, typeName))
}

func (o *Output) AddEndpoint(server CurrentServer) {
	parameters := []string{}
	if server.Header != "" {
		parameters = append(parameters, fmt.Sprintf("header %s", server.Header))
	}
	if server.Body != "" {
		parameters = append(parameters, fmt.Sprintf("body %s", server.Body))
	}
	if server.GetCookie != "" {
		parameters = append(parameters, fmt.Sprintf("cookie %s", server.GetCookie))
	}
	returns := []string{}
	if server.Return != "" {
		returns = append(returns, server.Return)
	}

	if server.SetCookie != "" {
		returns = append(returns, server.SetCookie)
	}

	// Can always return an error
	returns = append(returns, "error")

	o.Add(fmt.Sprintf("%s(%s) (%s)\n", server.Name, strings.Join(parameters, ", "), strings.Join(returns, ", ")))
	o.server = append(o.server, fmt.Sprintf("mux.HandleFunc(\"%s\", utils.Handle(server.%s))", server.Url, server.Name))
}

func (o *Output) FinishVariable() {
	o.InStruct = false
	o.InServer = false
	o.variables = append(o.variables, o.currentVariable)
	o.currentVariable = ""
}

func (o *Output) String() string {
	if o.currentVariable != "" {
		panic("Ended run with current variable still full")
	}
	variables := strings.Join(o.variables, "\n\n")
	server := strings.Join(o.server, "\n")

	return fmt.Sprintf("%s\n\n%s", server, variables)
}

func GenerateAPI(name string) (string, error) {
	i, err := ParseFile(name)
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
			output.Add("}")
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
			output.Add("type Server interface {\n")
			output.InServer = true
			items.Next()
			// Server can have a variable amount of items
			// The first 3 are HTTP verb (POST, GET, PUT, PATCH, and DELETE), url, and function name
			// followed by key value pairs in the form of name(type)
			// current types: body, setcookie, header, returns, getcookie
			// example: body(TestBody)
			// TODO allow for more then body and return
		} else if output.InServer {
			server := CurrentServer{}
			server.Verb = item
			server.Url = items.NextItem()
			server.Name = items.NextItem()

			for items.ValidPosition() {
				match := re.FindStringSubmatch(items.PeekItem())
				if match == nil {
					break
				}
				parName := match[1]
				parType := match[2]
				if parName == "returns" {
					server.Return = parType
				} else if parName == "setcookie" {
					server.SetCookie = parType
				} else if parName == "body" {
					server.Body = parType
				} else if parName == "header" {
					server.Header = parName
				} else if parName == "getcookie" {
					server.GetCookie = parType
				} else {
					panic(fmt.Sprintf("Invalid type: %s", parName))
				}
				items.Next()
			}
			output.AddEndpoint(server)

		} else {
			panic("AHHH")
		}
		items.Next()
	}
	return output.String()
}
