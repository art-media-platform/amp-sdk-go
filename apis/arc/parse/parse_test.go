package parse

import (
	"fmt"
	"testing"
)

func TestExpr(t *testing.T) {
	var tsts = []string{
		"elem-type.org",
		"[UTC16]elem",
		//"elem:name",
		//"elem-type.org:name",
		"[SurfaceName.UID]elem-type",
		"[Locale.Name]elem-type",
	}

	for _, tst := range tsts {
		spec, err := attrSpecParser.ParseString("", tst)
		if err != nil {
			fmt.Printf("%-30s %v\n", tst, err)
		} else {
			fmt.Printf("%-30s %-15v %-15v\n", tst, spec.SeriesSpec, spec.ElemType)
		}
	}

}
