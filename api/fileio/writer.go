package fileio

import (
	"errors"
	"os"
)

func WriteToFile(directory string, name string, output string) error {
	return os.WriteFile(directory + "/" + name+".go", []byte(output), 0644)
}

func SafeWriteToFile(directory string, name string, output string) error {
	if fileExists(directory, name) {
		return nil
	}
	return os.WriteFile(directory + "/" + name+".go", []byte(output), 0644)
}

func WriteEndpoints(directory string, endpoints map[string]string) error {
	for name, endpoint := range endpoints {
		err := SafeWriteToFile(directory, name, endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func fileExists(directory string, name string) bool {
	_, err := os.Stat(directory + "/" + name + ".go")
	return !errors.Is(err, os.ErrNotExist)
}
