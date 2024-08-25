package generatev2

import (
	cr "github.com/medubin/gonzo/api/generate/character_reader"
	lex "github.com/medubin/gonzo/api/generate/lexer"
	parser "github.com/medubin/gonzo/api/generate/parser/simple_golang_server"
)

func Generate(file []byte, stack string) (string, map[string]string, error) {

	// Struct around the []byes reads the characters and provides positional information
	c := cr.NewCharacterReader(file)

	l := lex.NewLexer(c)
	l.Lex()

	isClient := false
	if stack == "client" {
		isClient = true
	}

	p := parser.NewParser(l)
	types, endpoints, err := p.Parse(parser.ParseOptions{
		SkipServer: isClient,
	})

	if err != nil {
		return "", nil, err
	}

	return types, endpoints, nil
}
