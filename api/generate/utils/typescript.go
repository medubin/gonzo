package utils

import (
	"encoding/json"

	"github.com/gzuidhof/tygo/tygo"
)

func ConvertToTypescript(path string) error {
	println("hi")
	config := &tygo.Config{

		Packages: []*tygo.PackageConfig{
			{

				Path: path,
			},
		},
	}
	gen := tygo.New(config)
	x, _ := json.Marshal(config)
	println(string(x))
	return gen.Generate()
}
