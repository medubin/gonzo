package generatev2

const (
	Default State = iota
	Type
	Struct
)

type Node struct {
	Parent   *Node
	Children []*Node
	Token    string
	State    State
}

type Parser struct {
	tokens []string
	state  State
}

func (P *Parser) Parse() {

}
