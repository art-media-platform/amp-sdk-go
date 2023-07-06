package parse

import (
	"fmt"
	"testing"
)

func TestExpr(t *testing.T) {
	var tsts = []string{
		"elem-type.org",
		"[series.org]elem",
		"elem:name",
		"elem-type.org:name",
		"[series]elem:name",
		"[series]elem-type:name.ext",
	}

	for _, tst := range tsts {
		spec, err := AttrSpecParser.ParseString("", tst)
		if err != nil {
			fmt.Printf("%-30s %v\n", tst, err)
		} else {
			fmt.Printf("%-30s %-15v %-15v %-15v\n", tst, spec.SeriesType, spec.ElemType, spec.AttrName)
		}
	}

}
