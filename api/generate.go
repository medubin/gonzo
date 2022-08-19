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
	return os.WriteFile("../server/"+name+".go", []byte(output), 0644)
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
		for _, i := range subItems {
			if i != "" {
				items = append(items, i)
			}
		}

		if err := s.Err(); err != nil {
			return nil, err
		}
	}
	return items, nil
}

func GenerateOutput(items Items) string {
	data := Data{}

	for items.ValidPosition() {
		item := items.Item()

		if item == "}" {
			data.FinishVariable()

			// Lines that start with "type" should always have 3 items
		} else if item == "type" {
			name := items.NextItem()
			typeName := items.NextItem()
			if typeName == "{" {
				data.InStruct = true
				data.AddStruct(name)
			} else {
				data.AddVariable(name, typeName)
			}
			// fields in structs should have 2 items
		} else if data.InStruct {
			typeName := items.NextItem()

			data.AddStructField(item, typeName)
			// Server setup contains 2 items
		} else if item == "Server" {
			data.InServer = true
			items.Next()
			// Server can have a variable amount of items
			// The first 3 are HTTP verb (POST, GET, PUT, PATCH, and DELETE), url, and function name
			// followed by key value pairs in the form of name(type)
			// current types: body, returns
			// example: body(TestBody)
		} else if data.InServer {
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
			data.AddEndpoint(e)

		} else {
			panic("AHHH")
		}
		items.Next()
	}
	return Output(data)
}
