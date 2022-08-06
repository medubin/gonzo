package api

import (
	"bufio"
	"bytes"
	"fmt"
	"go/format"
	"os"
	"regexp"
	"strings"
)

var re = regexp.MustCompile(`(?mi)^(?P<t>return|body)\((?P<v>[a-zA-Z]+)\)$`)

func GenerateAPI(name string) (string, error) {
	items, err := ParseFile(name)
	if err != nil {
		return "", err
	}
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

func GenerateOutput(items []string) string {
	indentation := ""
	inStruct := false
	inServer := false

	// types
	var t bytes.Buffer

	// Server function
	// var s bytes.Buffer

	i := 0
	for i < len(items) {
		item := items[i]

		if item == "}" {
			indentation = indentation[2:]
			inStruct = false
			inServer = false
			t.WriteString(fmt.Sprintf("%s}\n\n", indentation))

			// Lines that start with "type" should always have 3 items
		} else if item == "type" {
			i++
			name := items[i]
			i++
			typeName := items[i]
			if typeName == "{" {
				indentation += "  "
				inStruct = true
				typeName = "struct {"
			} else {
				typeName = typeName + "\n"
			}
			t.WriteString(fmt.Sprintf("%s %s %s\n", item, name, typeName))
			// fields in structs should have 2 items
		} else if inStruct {
			i++
			typeName := items[i]
			t.WriteString(fmt.Sprintf("%s%s %s\n", indentation, item, typeName))
			// Server setup contains 2 items
		} else if item == "Server" {
			t.WriteString("type Server interface { \n")
			inServer = true
			indentation += "  "
			i++
			// Server can have a variable amount of items
			// The first 3 are HTTP verb (POST, GET, PUT, PATCH, and DELETE), url, and function name
			// followed by key value pairs in the form of name(type)
			// current types: body, header, returns
			// example: body(TestBody)
			// TODO allow for more then body and return
		} else if inServer {
			//  verb := item
			i++
			//  url := items[i]
			i++
			// name := items[i]
			// for i+1 < len(items) && strings.Contains(items[i+1], "(") && strings.HasSuffix(items[i+1], ")") {
			// if err := s.Err(); err != nil {
			// returnValue:=
			for i < len(items) {
				match := re.FindStringSubmatch(items[i+1]);
				if match == nil {
					break
				}
				parName := match[1]
				parType := match[2]

				println(parName)
				println(parType)
				i++
			}

			// }
			// subItems := strings.Split(items[i+1], "(")
			// for i, match := range re.FindStringSubmatch(items[i+1]) {
			// 	fmt.Println(match, "found at index", i)
			// }
			// i++
			// }

		} else {
			panic("AHHH")
		}
		i++
	}
	return t.String()
}

// // func generateOutput(items []string) {
// // 	indentation := ""
// // 	inType := false
// // 	nameGiven := false

// // 	var b bytes.Buffer
// // 	for _, item := range items {
// // 		if item == "{" {
// // 			toWrite := item
// // 			if inType {
// // 				toWrite = "struct " + item
// // 			}
// // 			b.WriteString(fmt.Sprintf("%s%s\n", indentation, toWrite))
// // 			indentation += "  "
// // 		} else if item == "}" {
// // 			b.WriteString(fmt.Sprintf("%s%s\n\n", indentation, item))
// // 			indentation = indentation[2:]
// // 			inType = false
// // 		} else if item == "type" {
// // 			inType = true
// // 			b.WriteString(fmt.Sprintf("%s%s ", indentation, item))
// // 		} else if inType && nameGiven {
// // 			b.WriteString(fmt.Sprintf("%s%s\n", indentation, item))
// // 			nameGiven = false
// // 		} else if inType {
// // 			nameGiven = true
// // 			b.WriteString(fmt.Sprintf("%s%s ", indentation, item))
// // 		} else {
// // 			b.WriteString(fmt.Sprintf("%s%s\n", indentation, item))
// // 		}

// // 	}
// // 	println(b.String())
// // }

// func readFile(s *bufio.Scanner) {
// 	identation := 0
// 	var b bytes.Buffer
// 	for s.Scan() {
// 		line := strings.TrimSpace(s.Text())
// 		if len(line) == 0 {
// 			continue
// 		}

// 		if strings.HasSuffix(line, "{") {
// 			identation++
// 		}

// 		if strings.HasPrefix(line, "type") {
// 			name, varType := parseType(line)
// 			b.WriteString(fmt.Sprintf("type %s %s\n", name, varType))
// 		} else if line == "}" {
// 			b.WriteString("}\n")
// 		} else {
// 			b.WriteString(fmt.Sprintf("  %s\n", line))
// 		}

// 		if strings.HasSuffix(line, "}") {
// 			identation--
// 			b.WriteString("\n")
// 		}
// 	}
// 	println(b.String())
// }

// func parseType(line string) (string, string) {
// 	items := strings.Split(line, " ")

// 	name, varType := items[1], items[2]

// 	if varType == "{" {
// 		varType = "struct {"
// 	}

// 	return name, varType

// }
