package api

import (
	"go/format"

	"github.com/medubin/gonzo/api/generate/data"
	"github.com/medubin/gonzo/api/generate/output"
	"github.com/medubin/gonzo/api/generate/utils"
	"github.com/medubin/gonzo/api/generate/validation"
)

func GenerateData(lines []string) (*data.Data, error) {
	units, err := data.GenerateUnits(lines)
	if err != nil {
		return nil, err
	}

	data, err := data.ConvertUnitToData(*units)
	if err != nil {
		return nil, err
	}

	err = validation.CheckDuplicates(*data)
	if err != nil {
		return nil, err
	}

	err = validation.CheckTypes(*data)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func GenerateTypes(data *data.Data) (string, error) {
	server := output.Types(data)
	formattedOutput, err := format.Source([]byte(server))
	if err != nil {
		return "", err
	}
	return string(formattedOutput), nil
}

func GenerateEndpoints(data *data.Data) (map[string]string, error) {
	endpoints := make(map[string]string)
	for _, e := range data.Servers[0].Endpoints {
		endpoint := output.Endpoint(e)
		formattedEndpoint, err := format.Source([]byte(endpoint))
		if err != nil {
			return nil, err
		}
		println(e.Name)
		endpoints[utils.ToSnakeCase(e.Name)] = string(formattedEndpoint)
	}

	return endpoints, nil
}

func GenerateServer() (string, error) {
	server := output.Server()
	formattedServer, err := format.Source([]byte(server))
	if err != nil {
		return "", err
	}
	return string(formattedServer), nil
}
