package generatev2

import (
	cr "github.com/medubin/gonzo/api/generate/character_reader"
	lex "github.com/medubin/gonzo/api/generate/lexer"
	golang_parser "github.com/medubin/gonzo/api/generate/parser/simple_golang_server"
	ts_parser "github.com/medubin/gonzo/api/generate/parser/typescript_client"
)

func Generate(file []byte, stack string, language string) (string, map[string]string, error) {

	// Struct around the []byes reads the characters and provides positional information
	c := cr.NewCharacterReader(file)

	l := lex.NewLexer(c)
	l.Lex()

	if stack == "client" && language == "typescript" {
		p := ts_parser.NewParser(l)
		types, err := p.Parse()
		if err != nil {
			return "", nil, err
		}
		return types, nil, nil
	}

	p := golang_parser.NewParser(l)
	types, endpoints, err := p.Parse(golang_parser.ParseOptions{
		SkipServer: stack == "client",
	})

	if err != nil {
		return "", nil, err
	}

	return types, endpoints, nil
}
