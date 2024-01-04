package simplegolangserverparser

import (
	"fmt"
	"strings"

	lex "github.com/medubin/gonzo/api/generate/lexer"
	"github.com/medubin/gonzo/api/generate/utils"
)

type server struct {
	Name            string
	Endpoints       []endpoint
	currentEndpoint endpoint
	isInURLVariable bool
}

type endpoint struct {
	Name         string
	Body         string
	Returns      string
	Verb         string
	URL          string
	URLVariables []string
}

func (s *server) ParseServer(token *lex.Token) {
	switch token.Type {
	case lex.FS:
		s.currentEndpoint.URL += token.Chars
	case lex.NEWLINE:
		if s.currentEndpoint.Name != "" {
			s.Endpoints = append(s.Endpoints, s.currentEndpoint)
			s.currentEndpoint = endpoint{}
		}
	case lex.LT:
		s.currentEndpoint.URL += "{"
		s.isInURLVariable = true
	case lex.GT:
		s.currentEndpoint.URL += "}"
		s.isInURLVariable = false
	case lex.IDENT:
		// first ident should be server name
		if s.Name == "" {
			s.Name = token.Chars
		} else if s.currentEndpoint.Name == "" {
			s.currentEndpoint.Name = token.Chars
		} else if s.currentEndpoint.Verb == "" {
			s.currentEndpoint.Verb = token.Chars
		} else {
			s.currentEndpoint.URL += token.Chars
			if s.isInURLVariable {
				s.currentEndpoint.URLVariables = append(s.currentEndpoint.URLVariables, token.Chars)
			}
		}
	}
}

func (s *server) ParseBody(token *lex.Token) {
	switch token.Type {
	case lex.IDENT:
		s.currentEndpoint.Body = token.Chars
	}
}

func (s *server) ParseReturns(token *lex.Token) {
	switch token.Type {
	case lex.IDENT:
		s.currentEndpoint.Returns = token.Chars
	}
}

func (s *server) OutputTypes() string {
	text := ""
	for _, e := range s.Endpoints {
		text += fmt.Sprintf("type %sUrl struct {\n", e.Name)
		for _, v := range e.URLVariables {
			text += fmt.Sprintf("  %s *string\n", v)
		}
		text += "}\n\n"
	}

	text += fmt.Sprintf("type %s interface {\n", s.Name)

	for _, e := range s.Endpoints {
		bodyType := ""
		if e.Body == "" {
			bodyType = "interface{}"
		} else {
			bodyType = e.Body
		}
		returnType := ""
		if e.Returns == "" {
			returnType = "interface{}"
		} else {
			returnType = e.Returns
		}

		text += fmt.Sprintf("  %s(ctx context.Context, body *%s, cookie cookies.Cookies, url url.URL[%s]) (*%s, error)\n", e.Name, bodyType, e.Name+"Url", returnType)
	}
	text += "}\n\n"

	text += fmt.Sprintf("func Start%s(s %s, r *router.Router) {\n", s.Name, s.Name)
	for _, e := range s.Endpoints {
		text += fmt.Sprintf("  r.Route(\"%s\", \"%s\", handle.Handle(s.%s))\n", e.Verb, e.URL, e.Name)
	}
	text += "}\n\n"

	return text
}

func (s *server) OutputEndpoints() map[string]string {
	endpoints := make(map[string]string)
	for _, e := range s.Endpoints {
		endpointHeader := generateEndpoint(e)
		endpoints[utils.ToSnakeCase(e.Name)] = fmt.Sprintf(`package server

import (
	"context"

  "github.com/medubin/gonzo/api/src/cookies"
  "github.com/medubin/gonzo/api/src/gerrors"
  "github.com/medubin/gonzo/api/src/url"
)

// %s %s
func (s *ServerImpl) %s {
	return nil, gerrors.UnimplementedError("%s")
}
`, e.Verb, e.URL, endpointHeader, e.Name)
	}

	return endpoints
}

func generateEndpoint(e endpoint) string {
	parameters := []string{"ctx context.Context"}

	if e.Body != "" {
		parameters = append(parameters, fmt.Sprintf("body *%s", e.Body))
	} else {
		parameters = append(parameters, fmt.Sprintf("body *%s", "interface{}"))
	}

	parameters = append(parameters, "cookie cookies.Cookies")
	parameters = append(parameters, fmt.Sprintf("url url.URL[%sUrl]", e.Name))
	returns := []string{}
	if e.Returns != "" {
		returns = append(returns, "*"+e.Returns)
	}

	// Can always return an error
	returns = append(returns, "error")

	return fmt.Sprintf("%s(%s) (%s)", e.Name, strings.Join(parameters, ", "), strings.Join(returns, ", "))
}
