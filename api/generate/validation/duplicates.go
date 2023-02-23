package validation

import (
	"fmt"

	"github.com/medubin/gonzo/api/generate/data"
)

func CheckDuplicates(data data.Data) error {
	err := checkDuplicateVariables(data)
	if err != nil {
		return err
	}
	err = checkDuplicateServer(data)
	if err != nil {
		return err
	}
	return nil
}

func checkDuplicateVariables(data data.Data) error {
	variableCheck := make(map[string]bool)

	for _, item := range data.Variables {
		name := item.Name
		if variableCheck[name] {
			return fmt.Errorf("duplicate variables with the name %s", name)
		}
		variableCheck[name] = true

		fieldCheck := make(map[string]bool)
		for _, f := range item.Fields {
			if fieldCheck[f.Name] {
				return fmt.Errorf("duplicate field with the name %s on type %s", f.Name, name)
			}
			fieldCheck[f.Name] = true
		}

	}
	return nil
}

func checkDuplicateServer(data data.Data) error {
	serverCheck := make(map[string]bool)
	for _, server := range data.Servers {
		if serverCheck[server.Name] {
			return fmt.Errorf("duplicate server with the name %s", server.Name)
		}
		serverCheck[server.Name] = true

		endpointCheck := make(map[string]bool)
		routeCheck := make(map[string]bool)

		for _, e := range server.Endpoints {
			if endpointCheck[e.Name] {
				return fmt.Errorf("duplicate endpoint with the name %s", e.Name)
			}
			endpointCheck[e.Name] = true

			if routeCheck[e.Url] {
				return fmt.Errorf("duplicate endpoint with the url %s", e.Url)
			}
			routeCheck[e.Url] = true
		}
	}
	return nil
}
