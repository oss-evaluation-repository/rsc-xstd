// Code generated by "stringer -type urlPart"; DO NOT EDIT.

package template

import "rsc.io/xstd/go1.17.8/strconv"

const _urlPart_name = "urlPartNoneurlPartPreQueryurlPartQueryOrFragurlPartUnknown"

var _urlPart_index = [...]uint8{0, 11, 26, 44, 58}

func (i urlPart) String() string {
	if i >= urlPart(len(_urlPart_index)-1) {
		return "urlPart(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _urlPart_name[_urlPart_index[i]:_urlPart_index[i+1]]
}
