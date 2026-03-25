package fileio

import (
	"os"
)

func ParseFile(name string) ([]byte, error) {
	file, err := os.ReadFile(name)
	if err != nil {
		return nil, err
	}

	return file, nil
}
