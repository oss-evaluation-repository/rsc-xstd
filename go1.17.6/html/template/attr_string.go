// Code generated by "stringer -type attr"; DO NOT EDIT.

package template

import "rsc.io/xstd/go1.17.6/strconv"

const _attr_name = "attrNoneattrScriptattrScriptTypeattrStyleattrURLattrSrcset"

var _attr_index = [...]uint8{0, 8, 18, 32, 41, 48, 58}

func (i attr) String() string {
	if i >= attr(len(_attr_index)-1) {
		return "attr(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _attr_name[_attr_index[i]:_attr_index[i+1]]
}
