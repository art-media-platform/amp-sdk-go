package parse

import (
	"fmt"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type AttrSpecExpr struct {
	SeriesSpec string `( "[" @Ident "]" )?`
	ElemType   string ` @Ident `
	AsCanonic  string
	//AttrName   string `( ":" @Ident )?`
}

var attrLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "Ident", Pattern: `[a-zA-Z][-._\w]*`},
	{Name: "Punct", Pattern: `[[:/]|]`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
	//{"Comment", `(?:#|//)[^\n]*\n?`},
	//{"Number", `(?:\d*\.)?\d+`},
	//{Name: "Punct", Pattern: `[[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
})

var attrSpecParser = participle.MustBuild[AttrSpecExpr](
	participle.Lexer(attrLexer),
	participle.Elide("Whitespace"),
	//, participle.UseLookahead(2))
)

func ParseAttrSpecExpr(attrExpr string) (expr *AttrSpecExpr, err error) {
	expr, err = attrSpecParser.ParseString("", attrExpr)
	if err != nil {
		return
	}

	if expr.SeriesSpec == "" {
		expr.AsCanonic = expr.ElemType
	} else {
		expr.AsCanonic = fmt.Sprintf("[%s]%s", expr.SeriesSpec, expr.ElemType)
	}
	return
}
