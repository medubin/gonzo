package output

import (
	"fmt"

	"github.com/medubin/gonzo/api/data"
)

func Endpoint(e *data.Endpoint) string {
	endpoint := generateEndpoint(e)

	return fmt.Sprintf(`package server
	import (
		"context"
		"errors"

		"github.com/medubin/gonzo/utils/cookies"
		"github.com/medubin/gonzo/utils/url"
	)
	
func (s *ServerImpl) %s {
	return nil, errors.New("Not Implemented")
}
	`, endpoint)
}
