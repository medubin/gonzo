package api

import (
	"go/format"

	"github.com/medubin/gonzo/api/generate/data"
	"github.com/medubin/gonzo/api/generate/output"

	"github.com/medubin/gonzo/api/generate/utils"
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

	return data, nil

	// parse file into lines
	// split line sections into functional units
	// make sure there are no duplicates
	// process each functional unit
	// write to file
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
