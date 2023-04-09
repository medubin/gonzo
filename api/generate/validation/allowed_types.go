package validation

import (
	"errors"

	"github.com/medubin/gonzo/api/generate/data"
)

var allowedTypes = map[string]map[string]string{
	"double": {"go": "float64"},
	"float":  {"go": "float32"},
	"int32":  {"go": "int32"},
	"int64":  {"go": "int64"},
	"uint32": {"go": "uint32"},
	"uint64": {"go": "uint64"},
	"bool":   {"go": "bool"},
	"string": {"go": "string"},
	"bytes":  {"go": "[]byte"},
	"{":      {"go": "struct{"},
}

func CheckTypes(data data.Data) error {
	// First we grab all of the names of the new variables
	createdVariables := make(map[string]bool)
	for _, item := range data.Variables {
		createdVariables[item.Name] = true
	}

	for _, v := range data.Variables {
		err := checkType(v.Type, createdVariables)
		if err != nil {
			return err
		}

		for _, f := range v.Fields {
			err := checkType(f.Type, createdVariables)
			if err != nil {
				return err
			}

			if f.MapValue != "" {
				err := checkType(f.MapValue, createdVariables)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func checkType(t string, createdVariables map[string]bool) error {
	_, ok := allowedTypes[t]
	_, ok2 := createdVariables[t]
	if !ok && !ok2 {
		return errors.New("unknown type: " + t)
	}
	return nil
}
