// Code generated by "stringer -type=TokenType"; DO NOT EDIT.

package sexp

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[RAWSTRING-66]
	_ = x[QUOTESTRING-81]
	_ = x[NUMBER-78]
	_ = x[RIGHT-41]
	_ = x[LEFT-40]
	_ = x[EOF-9220]
	_ = x[INVALID-65533]
}

const (
	_TokenType_name_0 = "LEFTRIGHT"
	_TokenType_name_1 = "RAWSTRING"
	_TokenType_name_2 = "NUMBER"
	_TokenType_name_3 = "QUOTESTRING"
	_TokenType_name_4 = "EOF"
	_TokenType_name_5 = "INVALID"
)

var (
	_TokenType_index_0 = [...]uint8{0, 4, 9}
)

func (i TokenType) String() string {
	switch {
	case 40 <= i && i <= 41:
		i -= 40
		return _TokenType_name_0[_TokenType_index_0[i]:_TokenType_index_0[i+1]]
	case i == 66:
		return _TokenType_name_1
	case i == 78:
		return _TokenType_name_2
	case i == 81:
		return _TokenType_name_3
	case i == 9220:
		return _TokenType_name_4
	case i == 65533:
		return _TokenType_name_5
	default:
		return "TokenType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
}
