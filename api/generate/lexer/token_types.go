package lex

// keywords - maps string keywords to their corresponding Type.
var keywords = map[string]Type{
	"type":   TYPE,
	"server": SERVER,
	"repeated": REPEATED,
	"map": MAP,
	"returns": RETURNS,
	"body": BODY,
	"true": BOOLEAN,
	"false": BOOLEAN,
}

func LookupKeywords(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

type Type string

const (
	/*
		Special Types
	*/
	EOF     = "EOF"
	ILLEGAL = "ILLEGAL"
	NEWLINE = "NEWLINE"

	/*
		Identifiers & Literals
	*/
	IDENT   = "IDENT"
	STRING  = "STRING"
	INTEGER = "INTEGER"
	DOUBLE  = "DOUBLE"
	BOOLEAN = "BOOLEAN"

	/*
		Symbols
	*/
	COMMA  = "COMMA"
	LCB    = "LCB"
	RCB    = "RCB"
	LP     = "LP"
	RP     = "RP"
	ASSIGN = "ASSIGN"
	SIGN   = "SIGN"
	COLON  = "COLON"
	HASH   = "HASH"
	QM     = "QM"
	DOT    = "DOT"
	LSB    = "LSB"
	RSB    = "RSB"
	EXCL   = "EXCL"
	PLUS   = "PLUS"
	MINUS  = "MINUS"
	TIMES  = "TIMES"
	DIVIDE = "DIVIDE"
	MOD    = "MOD"
	POW    = "POW"
	GT     = "GT"
	LT     = "LT"
	FS     = "FS"

	/*
		Keywords
	*/
	TYPE   = "TYPE"
	SERVER = "SERVER"
	REPEATED = "REPEATED"
	MAP = "MAP"
	RETURNS = "RETURNS"
	BODY = "BODY"


	/*
		Comments
	*/
	SINGLE_LINE_COMMENT = "SINGLE_LINE_COMMENT"
	MULTI_LINE_COMMENT  = "MULTI_LINE_COMMENT"

	/*
		Whitespace
	*/
	SPACE = "SPACE"
	TAB   = "TAB"
)
