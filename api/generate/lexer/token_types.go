package lex

// keywords - maps string keywords to their corresponding Type.
var keywords = map[string]Type{
	"type":     TYPE,
	"server":   SERVER,
	"repeated": REPEATED,
	"map":      MAP,
	"returns":  RETURNS,
	"body":     BODY,
	"true":     BOOLEAN,
	"false":    BOOLEAN,
	"POST":     POST,
	"GET":      GET,
	"DELETE":   DELETE,
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
	EOF     Type = "EOF"
	ILLEGAL Type = "ILLEGAL"
	NEWLINE Type = "NEWLINE"

	/*
		Identifiers & Literals
	*/
	IDENT   Type = "IDENT"
	STRING  Type = "STRING"
	INTEGER Type = "INTEGER"
	DOUBLE  Type = "DOUBLE"
	BOOLEAN Type = "BOOLEAN"

	/*
		Symbols
	*/
	COMMA  Type = "COMMA"
	LCB    Type = "LCB"
	RCB    Type = "RCB"
	LP     Type = "LP"
	RP     Type = "RP"
	ASSIGN Type = "ASSIGN"
	SIGN   Type = "SIGN"
	COLON  Type = "COLON"
	HASH   Type = "HASH"
	QM     Type = "QM"
	DOT    Type = "DOT"
	LSB    Type = "LSB"
	RSB    Type = "RSB"
	EXCL   Type = "EXCL"
	PLUS   Type = "PLUS"
	MINUS  Type = "MINUS"
	TIMES  Type = "TIMES"
	DIVIDE Type = "DIVIDE"
	MOD    Type = "MOD"
	POW    Type = "POW"
	GT     Type = "GT"
	LT     Type = "LT"
	FS     Type = "FS"

	/*
		Keywords
	*/
	TYPE     Type = "TYPE"
	SERVER   Type = "SERVER"
	REPEATED Type = "REPEATED"
	MAP      Type = "MAP"
	RETURNS  Type = "RETURNS"
	BODY     Type = "BODY"
	POST     Type = "POST"
	GET      Type = "GET"
	DELETE   Type = "DELETE"
	URL      Type = "URL"

	/*
		Comments
	*/
	SINGLE_LINE_COMMENT Type = "SINGLE_LINE_COMMENT"
	MULTI_LINE_COMMENT  Type = "MULTI_LINE_COMMENT"

	/*
		Whitespace
	*/
	SPACE Type = "SPACE"
	TAB   Type = "TAB"
)
