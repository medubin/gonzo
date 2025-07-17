package fileio

import (
	"errors"
	"os"
)

func WriteToFile(directory string, name string, output string, isTypescript bool) error {
	_ = os.Mkdir(directory, os.ModePerm)

	var extension = ".go"
	if isTypescript {
		extension = ".ts"
	}

	file, err := os.Create(directory + "/" + name + extension)
	if err != nil {
		return err
	}
	_, err = file.WriteString(output)
	if err != nil {
		return err
	}

	return nil
}

func SafeWriteToFile(directory string, name string, output string) error {
	_ = os.Mkdir(directory, os.ModePerm)

	if fileExists(directory, name) {
		return nil
	}
	return os.WriteFile(directory+"/"+name+".go", []byte(output), 0644)
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
