// Code generated by "stringer -type jsCtx"; DO NOT EDIT.

package template

import "rsc.io/xstd/go1.12.7/strconv"

const _jsCtx_name = "jsCtxRegexpjsCtxDivOpjsCtxUnknown"

var _jsCtx_index = [...]uint8{0, 11, 21, 33}

func (i jsCtx) String() string {
	if i >= jsCtx(len(_jsCtx_index)-1) {
		return "jsCtx(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _jsCtx_name[_jsCtx_index[i]:_jsCtx_index[i+1]]
}
