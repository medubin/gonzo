package output

import (
	"fmt"

	"github.com/medubin/gonzo/api/generate/data"
)

func Endpoint(e data.Endpoint) string {
	endpoint := generateEndpoint(e)

	return fmt.Sprintf(`package server
	import (
		"context"

		"github.com/medubin/gonzo/api/src/cookies"
		"github.com/medubin/gonzo/api/src/url"
		"github.com/medubin/gonzo/api/src/gerrors"
	)
	
func (s *ServerImpl) %s {
	return nil, gerrors.UnimplementedError("%s")
}
	`, endpoint, e.Name)
}
