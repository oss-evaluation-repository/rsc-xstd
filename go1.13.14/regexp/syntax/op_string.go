// Code generated by "stringer -type Op -trimprefix Op"; DO NOT EDIT.

package syntax

import "rsc.io/xstd/go1.13.14/strconv"

const (
	_Op_name_0 = "NoMatchEmptyMatchLiteralCharClassAnyCharNotNLAnyCharBeginLineEndLineBeginTextEndTextWordBoundaryNoWordBoundaryCaptureStarPlusQuestRepeatConcatAlternate"
	_Op_name_1 = "opPseudo"
)

var (
	_Op_index_0 = [...]uint8{0, 7, 17, 24, 33, 45, 52, 61, 68, 77, 84, 96, 110, 117, 121, 125, 130, 136, 142, 151}
)

func (i Op) String() string {
	switch {
	case 1 <= i && i <= 19:
		i -= 1
		return _Op_name_0[_Op_index_0[i]:_Op_index_0[i+1]]
	case i == 128:
		return _Op_name_1
	default:
		return "Op(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
