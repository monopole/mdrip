// Code generated by "stringer -type=EnumArch -linecomment"; DO NOT EDIT.

package internal

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ArchUnknown-0]
	_ = x[ArchAmd64-1]
	_ = x[ArchArm64-2]
}

const _EnumArch_name = "unknownamd64arm64"

var _EnumArch_index = [...]uint8{0, 7, 12, 17}

func (i EnumArch) String() string {
	if i < 0 || i >= EnumArch(len(_EnumArch_index)-1) {
		return "EnumArch(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _EnumArch_name[_EnumArch_index[i]:_EnumArch_index[i+1]]
}