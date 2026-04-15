package fileio

import (
	"errors"
	"go/format"
	"os"
	"path/filepath"
)

func WriteToFile(directory string, name string, output string, safe bool) error {
	filename := filepath.Join(directory, name)
	if err := os.MkdirAll(filepath.Dir(filename), os.ModePerm); err != nil {
		return err
	}

	if filepath.Ext(name) == ".go" {
		if formatted, err := format.Source([]byte(output)); err == nil {
			output = string(formatted)
		}
	}

	if safe {
		if fileExists(filename) {
			return nil
		}
		return os.WriteFile(filename, []byte(output), 0644)
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(output)
	return err
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return !errors.Is(err, os.ErrNotExist)
}
