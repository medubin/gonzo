package data

import (
	"fmt"
	"regexp"

	"github.com/medubin/gonzo/api/generate/utils"
)

var re = regexp.MustCompile(`(?mi)^(?P<t>body|returns)\((?P<v>[a-zA-Z]+)\)$`)

type Data struct {
	Servers   []Server
	Variables []Variable
}

type Server struct {
	Name      string
	Endpoints []Endpoint
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
	Fields []Field
}

type Field struct {
	Name string
	Type string
}

func ConvertUnitToData(units []Unit) (*Data, error) {
	data := Data{}

	for _, unit := range units {
		switch unit.Class {
		case ClassType:
			variable, err := convertUnitToVariable(unit)
			if err != nil {
				return nil, err
			}
			data.Variables = append(data.Variables, variable)
		case ClassServer:
			server, err := convertUnitToServer(unit)
			if err != nil {
				return nil, err
			}
			data.Servers = append(data.Servers, *server)
			data.Variables = append(data.Variables, convertURLsToVariables(*server)...)
		default:
			return nil, fmt.Errorf("type is incorrect: %s", unit.Class)
		}
	}
	return &data, nil
}

func convertUnitToVariable(unit Unit) (Variable, error) {
	variable := Variable{}
	for idx, line := range unit.Lines {
		if idx == 0 {
			variable.Name = line[1]
			variable.Type = line[2]
		} else if line[0] != "}" {
			if len(line) == 2 {
				// 3 items is a simple variable
				variable.Fields = append(variable.Fields, Field{
					Name: line[0],
					Type: line[1],
				})
			} else if len(line) == 3 {
				// should only be an array
				if line[1] != "repeated" {
					return variable, fmt.Errorf("unknown type %s", line[1])
				}

				variable.Fields = append(variable.Fields, Field{
					Name: line[0],
					Type: "[]" + line[2],
				})
			} else if len(line) == 4 {
				// should only be a map
				if line[1] != "map" {
					return variable, fmt.Errorf("unknown type %s", line[1])
				}
				variable.Fields = append(variable.Fields, Field{
					Name: line[0],
					Type: "map[" + line[2] + "]" + line[3],
				})

			}
		}
	}
	return variable, nil
}

func convertUnitToServer(unit Unit) (*Server, error) {
	server := Server{}
	for idx, line := range unit.Lines {
		if idx == 0 {
			server.Name = line[1]
		} else if line[0] != "}" {
			endpoint, err := createEndpoint(line)
			if err != nil {
				return nil, err
			}
			server.Endpoints = append(server.Endpoints, *endpoint)
		}
	}

	return &server, nil
}

func createEndpoint(line []string) (*Endpoint, error) {
	endpoint := Endpoint{
		Verb: line[0],
		Url:  line[1],
		Name: line[2],
	}

	for _, line := range line[3:] {
		match := re.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		parName := match[1]
		parType := match[2]
		if parName == "returns" {
			endpoint.Return = parType
		} else if parName == "body" {
			endpoint.Body = parType
		} else {
			return nil, fmt.Errorf("invalid type: %s", parName)
		}
	}

	return &endpoint, nil
}

func convertURLsToVariables(server Server) []Variable {
	variables := make([]Variable, len(server.Endpoints))
	for i, endpoint := range server.Endpoints {
		variables[i] = convertURLToVariable(endpoint)
	}
	return variables
}

func convertURLToVariable(endpoint Endpoint) Variable {
	matches := utils.GetKeys(endpoint.Url)
	fields := make([]Field, len(matches))
	for i, match := range matches {
		fields[i] = Field{
			Name: match,
			Type: "string",
		}
	}
	return Variable{
		Name:   endpoint.Name + "Url",
		Type:   "{",
		Fields: fields,
	}
}
