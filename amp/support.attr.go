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


func (attrID AttrID) PinLevel() int {   // TODO: remove PinLevel!?  Just stuff it in the HaloTag spec and MD5 that.
	return int(attrID[0] >> 61)
}

const (
	PinLevelBits = 3
	PinLevelMax  = (1 << PinLevelBits) - 1

	pinLevelMask  = uint64(PinLevelMax) << 61
	pinLevelShift = 64 - PinLevelBits
)

func (attrID *AttrID) ApplyPinLevel(pinLevel int) {
	attrID[0] &^= pinLevelMask
	attrID[0] |= uint64(pinLevel) << pinLevelShift
}

// Generates the AttrID for the given raw attr element type name (vs from an attr spec expression).
func FormBaseAttrID(canonicExpr string) AttrID {
	attrID := AttrID(StringToUID(canonicExpr))
	attrID[0] &^= pinLevelMask
	return attrID
}

func (attrID *AttrID) IsNil() bool {
	return attrID[0] == 0 && attrID[1] == 0
}

func (spec *TagSpec) AttrID() AttrID {
	return [2]uint64{
		spec.AttrIDx0,
		spec.AttrIDx1,
	}
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

var attrSpecParser = participle.MustBuild[TagSpecExpr](
	participle.Lexer(attrLexer),
	participle.Elide("Whitespace"),
	//, participle.UseLookahead(2))
)

func ParseAttrDef(tagDefExpr string) (expr *TagSpecExpr, err error) {
	expr, err = attrSpecParser.ParseString("", tagDefExpr)
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

func FormTagSpec(tagSpecExpr string) (TagSpec, error) {
	expr, err := ParseAttrDef(tagSpecExpr)
	if err != nil {
		return TagSpec{}, err
	}

	elemTypeID := FormBaseAttrID(expr.ElemType)

	attrID := FormBaseAttrID(expr.AsCanonic)
	attrID.ApplyPinLevel(expr.PinLevel)

	seriesID := StringToUID(expr.SeriesSpec)

	if expr.PinLevel == 0 && expr.SeriesSpec == "" {
		if elemTypeID != attrID {
			panic("FormTagSpec: elemType should match attrID")
		}
	}

	spec := TagSpec{
		AsCanonic:      expr.AsCanonic,
		AttrIDx0:       attrID[0],
		AttrIDx1:       attrID[1],
		ElemTypeIDx0:   elemTypeID[0],
		ElemTypeIDx1:   elemTypeID[1],
		SeriesSpecIDx0: seriesID[0],
		SeriesSpecIDx1: seriesID[1],
	}

	return spec, nil
}

func MustFormAttrID(tagSpecExpr string) AttrID {
	spec, err := FormTagSpec(tagSpecExpr)
	if err != nil {
		panic(err)
	}
	return spec.AttrID()
}

func FormAttrID(tagSpecExpr string) (AttrID, error) {
	spec, err := FormTagSpec(tagSpecExpr)
	if err != nil {
		return AttrID{}, err
	}
	return spec.AttrID(), nil
}
