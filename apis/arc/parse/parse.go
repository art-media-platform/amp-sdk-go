package parse

import (
	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type AttrSpecExpr struct {
	SeriesType string `( "[" @Ident "]" )?`
	ElemType   string ` @Ident `
	AttrName   string `( ":" @Ident )?`
}

var attrLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z][-._\w]*`},
	{Name: "Punct", Pattern: `[[:/]|]`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
	//{"Comment", `(?:#|//)[^\n]*\n?`},
	//{"Number", `(?:\d*\.)?\d+`},
	//{Name: "Punct", Pattern: `[[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
})

var AttrSpecParser = participle.MustBuild[AttrSpecExpr](
	participle.Lexer(attrLexer),
	participle.Elide("Whitespace"),
	//, participle.UseLookahead(2))
)
