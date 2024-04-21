package amp

import (
	"fmt"
	"strings"

	"github.com/alecthomas/participle/v2"
	"github.com/alecthomas/participle/v2/lexer"
)

type TagSpecExpr struct {
	PinLevel   int    `( @PosInt ":" )?`
	SeriesSpec string `( "[" (@Ident)? "]" )?`
	ElemType   string ` @Ident `
	AttrName   string `( ":" @Ident )?`
	AsCanonic  string
}

type HaloTag struct {
}

func (attrID TagSpecID) PinLevel() int { // TODO: remove PinLevel!?  Just stuff it in the HaloTag spec and MD5 that.
	return int(attrID[0] >> 61)
}

const (
	PinLevelBits = 3
	PinLevelMax  = (1 << PinLevelBits) - 1

	pinLevelMask  = uint64(PinLevelMax) << 61
	pinLevelShift = 64 - PinLevelBits
)

func (attrID *TagSpecID) ApplyPinLevel(pinLevel int) {
	attrID[0] &^= pinLevelMask
	attrID[0] |= uint64(pinLevel) << pinLevelShift
}

// Generates the TagSpecID for the given raw attr element type name (vs from an attr spec expression).
func FormBaseTagSpecID(canonicExpr string) TagSpecID {
	attrID := TagSpecID(StringToTagID(canonicExpr))
	attrID[0] &^= pinLevelMask
	return attrID
}

func (attrID *TagSpecID) IsNil() bool {
	return attrID[0] == 0 && attrID[1] == 0
}

var attrLexer = lexer.MustSimple([]lexer.SimpleRule{
	{Name: "PosInt", Pattern: `(?:\d*)?\d+`},
	{Name: "Ident", Pattern: `[a-zA-Z][-._\w]*`},
	{Name: "Punct", Pattern: `[[:/]|]`},
	{Name: "Whitespace", Pattern: `[ \t\n\r]+`},
	//{"Comment", `(?:#|//)[^\n]*\n?`},
	//{"Number", `(?:\d*\.)?\d+`},
	//{Name: "Punct", Pattern: `[[!@#$%^&*()+_={}\|:;"'<,>.?/]|]`},
})

var tagSpecParser = participle.MustBuild[TagSpecExpr](
	participle.Lexer(attrLexer),
	participle.Elide("Whitespace"),
	//, participle.UseLookahead(2))
)

func ParseAttrDef(tagDefExpr string) (expr *TagSpecExpr, err error) {
	expr, err = tagSpecParser.ParseString("", tagDefExpr)
	if err != nil {
		return
	}

	b := strings.Builder{}
	if expr.PinLevel > 0 {
		b.WriteString(fmt.Sprintf("%d:", expr.PinLevel))
	}
	if expr.SeriesSpec != "" {
		b.WriteByte('[')
		b.WriteString(expr.SeriesSpec)
		b.WriteByte(']')
	}
	b.WriteString(expr.ElemType)

	expr.AsCanonic = b.String()
	return
}

func FormTagSpec(canonicAttrSpecExpr string) (TagSpec, error) {
	expr, err := ParseAttrDef(canonicAttrSpecExpr)
	if err != nil {
		return TagSpec{}, err
	}

	elemTypeID := FormBaseTagSpecID(expr.ElemType)

	attrID := FormBaseTagSpecID(expr.AsCanonic)
	attrID.ApplyPinLevel(expr.PinLevel)

	seriesID := StringToTagID(expr.SeriesSpec)

	if expr.PinLevel == 0 && expr.SeriesSpec == "" {
		if elemTypeID != attrID {
			panic("FormTagSpec: elemType should match attrID")
		}
	}

	spec := TagSpec{
		AsCanonic:      canonicAttrSpecExpr,
		AttrSpecIDx0:   int64(attrID[0]),
		AttrSpecIDx1:   attrID[1],
		AttrSpecIDx2:   attrID[2],
		ElemSpecIDx0:   int64(elemTypeID[0]),
		ElemSpecIDx1:   elemTypeID[1],
		ElemSpecIDx2:   elemTypeID[2],
		SeriesSpecIDx0: int64(seriesID[0]),
		SeriesSpecIDx1: seriesID[1],
		SeriesSpecIDx2: seriesID[2],
	}
	return spec, nil
}

func MustFormAttrSpec(tagSpecExpr string) TagID {
	spec, err := FormTagSpec(tagSpecExpr)
	if err != nil {
		panic(err)
	}
	return spec.SpecID()
}

func FormTagSpecID(tagSpecExpr string) (TagID, error) {
	spec, err := FormTagSpec(tagSpecExpr)
	if err != nil {
		return TagID{}, err
	}
	return spec.SpecID(), nil
}
