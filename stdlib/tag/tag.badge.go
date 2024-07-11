package tag

import math "math"

//
//  TODO: draw each dot as a ring of radius 1.3 (sacred geo) in additive grey scale then colorize!
//  USe shaders:  the "badge" is a list of ring centers, radius, and amplitude -- use Linefy or a custom shader to draw the rings
//

// Canonical ASCII digit in a tag.ID visual encoding (3 bits aka base 8)
//type CanonicAsciiDigit byte

type AsciiDigit byte // base 8: ascii rune
type OctalDigit byte // base 8: 3 bits

var CanonicAsciiAlphabet = [8]AsciiDigit{
	'_', '.', ':', '*',
	'~', 'o', 'O', '0',
	
	// '.', 'o', '8', '@',
    // '~', 'x', 'X', '*',
}

type OctalEncoding [64]OctalDigit

const AsciiTemplate_old = "" +
	"     : * : . N N N N N N    \n" +
	"    ~ : : . N N N N N N N   \n" +
	"   O O   . N N N N N N N N  \n" +
	"    0 O o N N N N N N N N   \n" +
	"     O o N N N N N N N N    \n" +
	"      o N N N N N N N N     \n" +
	"       N N N N N N N N      \n" +
	"        N N N N N N N       \n" +
	"         N N N N N N        \n" +
	"          N N N N N         \n" +
	"           N N N N          \n" +
	"            N N N           \n" +
	"                            \n"

const AsciiTemplate = "" + // TODO: mirror along Y
	"        * N N N\n" +
	"       * N N N \n" +
	"      : N N N N\n" +
	"     : N N N N \n" +
	"    . N N N N N\n" +
	"   . N N N N N \n" +
	"  . N N N N N N\n" +
	" ~ N N N N N N \n" +
	"  o N N N N N N\n" +
	"   o N N N N N \n" +
	"    o N N N N N\n" +
	"     O N N N N \n" +
	"      O N N N N\n" +
	"       0 N N N \n" +
	"        0 N N N\n"

type Dot struct {
	X, Y      float32 // 0..1
	Amplitude float32 // -1..1
	CharIndex int
	IndexR    int
}

// // function oddr_offset_to_pixel(hex):
//     var x = size * sqrt(3) * (hex.col + 0.5 * (hex.row&1))
//     var y = size * 3/2 * hex.row
//     return Point(x, y)

type Badge struct {
	Dots     []Dot  // location of each dot in template
	Template []byte // ascii template with newlines
}

func (badge *Badge) RegenFromTemplate(template string) {
	if len(badge.Dots) <= 64 {
		badge.Dots = make([]Dot, 0, 64)
	} else {
		badge.Dots = badge.Dots[:0]
	}
	templateLen := len(template)
	if len(badge.Template) <= templateLen {
		badge.Template = make([]byte, 0, 2*templateLen)
	} else {
		badge.Template = badge.Template[:0]
	}

	root3 := float32(math.Sqrt(3))

	row, col := 0, 0
	for ci := 0; ci < templateLen; ci++ {
		c := template[ci]
		switch c {
		case 'N':
			badge.Dots = append(badge.Dots, Dot{
				X:         root3 * (float32(col) + 0.5*float32(row&1)),
				Y:         1.5 * float32(row),
				CharIndex: ci,
			})
		case '\n':
			// rowLen := ci - rowStart - 1
			// for i := 0; i < rowLen; i++ {
			// 	badge.Template = append(badge.Template, ' ')
			// }
			// dots := len(badge.Dots)
			// for mi := (row & 1); mi < rowLen; mi++ { // mirror the current row
			// 	c_mirror := template[ci-mi]
			// 	badge.Template = append(badge.Template, c_mirror)
			// 	if c_mirror == 'N' {
			// 		dots--
			// 		badge.Dots[dots].IndexR = len(badge.Template)
			// 	}
			// }
			// for di := rowStartDot; di < len(badge.Dots); di++ {
			// 	badge.Dots[di].IndexR = 2*rowLen - badge.Dots[di].IndexL
			// }
			// 	dot.IndexR = len(badge.Dots)
			// 	badge.Dots = append(badge.Dots, Dot{
			// 		X:         -dot.X,
			// 		Y:         dot.Y,
			// 		CharIndex: ci,
			// 		Sort:   col + row,
			// 	})
			// }

			// // mirror the current row
			// for di := dotL; di < len(badge.Dots); di++ {
			// 	dot := &badge.Dots[di]
			// 	dot.IndexR = len(badge.Dots)
			// 	badge.Dots = append(badge.Dots, Dot{
			// 		X:         -dot.X,
			// 		Y:         dot.Y,
			// 		CharIndex: ci,
			// 		Sort:   col + row,
			// 	})
			// }
			row++
			col = 0

		}
		badge.Template = append(badge.Template, c)
	}

	// sort.Slice(badge.Dots, func(i, j int) bool {
	// 	ix, iy := badge.Dots[i].X, badge.Dots[i].Y
	// 	jx, jy := badge.Dots[j].X, badge.Dots[j].Y
	// 	wi := ix + iy
	// 	return badge.Dots[i].CharIndex < badge.Dots[j].CharIndex
	// }
}

var gBadge6424 Badge

func init() {
	gBadge6424.RegenFromTemplate(AsciiTemplate)
}
