// Code generated by "stringer -type element"; DO NOT EDIT.

package template

import "rsc.io/xstd/go1.16.8/strconv"

const _element_name = "elementNoneelementScriptelementStyleelementTextareaelementTitle"

var _element_index = [...]uint8{0, 11, 24, 36, 51, 63}

func (i element) String() string {
	if i >= element(len(_element_index)-1) {
		return "element(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _element_name[_element_index[i]:_element_index[i+1]]
}
