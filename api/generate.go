package api

import (
	"bufio"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`(?mi)^(?P<t>returns|body)\((?P<v>[a-zA-Z]+)\)$`)

type Items struct {
	items    []string
	position int
}

func (i *Items) ValidPosition() bool {
	return i.position < len(i.items)
}

func (i *Items) Item() string {
	return i.items[i.position]
}

func (i *Items) Next() {
	i.position++
}

func (i *Items) NextItem() string {
	i.Next()
	return i.Item()
}

func (i *Items) PeekItem() string {
	return i.items[i.position+1]
}

func InitItems(items []string) Items {
	return Items{
		items:    items,
		position: 0,
	}
}

type Output struct {
	variables       []string
	currentVariable string
}

func (o *Output) Add(line string) {
	o.currentVariable += line
}

func (o *Output) FinishVariable() {
	o.variables = append(o.variables, o.currentVariable)
	o.currentVariable = ""
}

func (o *Output) String() string {
	return strings.Join(o.variables, "\n")
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
	inStruct := false
	inServer := false

	// Server function
	// var s bytes.Buffer

	for items.ValidPosition() {
		item := items.Item()

		if item == "}" {
			inStruct = false
			inServer = false
			output.Add(fmt.Sprintf("}\n"))
			output.FinishVariable()

			// Lines that start with "type" should always have 3 items
		} else if item == "type" {
			name := items.NextItem()
			typeName := items.NextItem()
			if typeName == "{" {
				inStruct = true
				typeName = "struct {"
				output.Add(fmt.Sprintf("%s %s %s\n", item, name, typeName))
			} else {
				output.Add(fmt.Sprintf("%s %s %s\n", item, name, typeName))
				output.FinishVariable()
			}
			// fields in structs should have 2 items
		} else if inStruct {
			typeName := items.NextItem()
			output.Add(fmt.Sprintf("%s %s\n", item, typeName))
			// Server setup contains 2 items
		} else if item == "Server" {
			output.Add("type Server interface {\n")
			inServer = true
			items.Next()
			// Server can have a variable amount of items
			// The first 3 are HTTP verb (POST, GET, PUT, PATCH, and DELETE), url, and function name
			// followed by key value pairs in the form of name(type)
			// current types: body, header, returns
			// example: body(TestBody)
			// TODO allow for more then body and return
		} else if inServer {
			//  verb := item
			items.Next()
			//  url := items[i]
			name := items.NextItem()

			returns := ""
			parameters := []string{}
			for items.ValidPosition() {
				match := re.FindStringSubmatch(items.PeekItem())
				if match == nil {
					break
				}
				parName := match[1]
				parType := match[2]
				if parName == "returns" {
					returns = parType
				} else {
					parameters = append(parameters, fmt.Sprintf("%s %s", parName, parType))
				}
				items.Next()
			}
			output.Add(fmt.Sprintf("%s(%s) %s\n", name, strings.Join(parameters, ", "), returns))

		} else {
			panic("AHHH")
		}
		items.Next()
	}
	return output.String()
}
