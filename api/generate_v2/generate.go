package generatev2

import "os"

func Generate(name string) error {
	file, err := os.ReadFile(name)
	if err != nil {
		return err
	}

	c := CharacterReader{}
	c.Initialize(string(file))

	l := Lexer{}
	l.Initialize(c)

	for !l.IsEOF() {
		println(l.Consume())
	}

	return nil

}
