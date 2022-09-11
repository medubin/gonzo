package fileio

import (
	"errors"
	"os"
)

func WriteToFile(name string, output string) error {
	return os.WriteFile("../server/"+name+".go", []byte(output), 0644)
}

func SafeWriteToFile(name string, output string) error {
	if fileExists(name) {
		return nil
	}
	return os.WriteFile("../server/"+name+".go", []byte(output), 0644)
}

func WriteEndpoints(endpoints map[string]string) error {
	for name, endpoint := range endpoints {
		err := SafeWriteToFile(name, endpoint)
		if err != nil {
			return err
		}
	}
	return nil
}

func fileExists(name string) bool {
	_, err := os.Stat("../server/" + name + ".go")
	return !errors.Is(err, os.ErrNotExist)
}
