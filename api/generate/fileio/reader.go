package fileio

import (
	"bufio"
	"os"
	"strings"
)


func ParseFile(name string) ([]string, error) {
	file, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	return convertFileToItems(scanner)
}


func convertFileToItems(s *bufio.Scanner) ([]string, error) {
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