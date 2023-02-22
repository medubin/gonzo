package data

import (
	"fmt"
	"strings"
)

type Class string

const (
	ClassType   Class = "type"
	ClassServer Class = "server"
)

// A unit is a single functional item. A type, a struct, a server, and an option are all units.
type Unit struct {
	Name  string
	Class Class
	Lines [][]string
}

// Takes a file split into lines and builds a list of units
func GenerateUnits(lines []string) (*[]Unit, error) {
	var units []Unit
	depth := 0
	currentUnit := Unit{}
	for lineNumber, s := range lines {

		line := strings.TrimSpace(s)

		// Skip empty lines
		if len(line) == 0 {
			continue
		}

		currentUnit.Lines = append(currentUnit.Lines, strings.Split(line, " "))

		// Check if we are going down depth. We panic if we are already at 0 depth
		if line == "}" {
			if depth == 0 {
				return nil, fmt.Errorf("mismatched parenthesis on line %d", lineNumber)
			}
			depth -= 1
		}

		// Add depth if we are opening a parenthesis
		if line[len(line)-1:] == "{" {
			depth += 1
		}

		// if at the end of this we are at depth 0 we are done with this unit
		if depth == 0 {
			units = append(units, currentUnit)
			currentUnit = Unit{}
		}
	}

	if depth != 0 {
		return nil, fmt.Errorf("mismatched parenthesis")
	}

	// Attaching the name and type seperately for readability.
	for i, unit := range units {
		units[i].Class = Class(unit.Lines[0][0])
		units[i].Name = unit.Lines[0][1]
	}

	return &units, nil
}
