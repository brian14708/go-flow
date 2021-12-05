package ident

var validIdentChars []bool

func init() {
	const validIdent = "0123456789_." +
		"abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	max := rune(0)
	for _, c := range validIdent {
		if c > max {
			max = c
		}
	}
	validIdentChars = make([]bool, int(max)+1)
	for _, c := range validIdent {
		validIdentChars[int(c)] = true
	}
}

func Check(s string) bool {
	if len(s) == 0 {
		return false
	}
	validLen := rune(len(validIdentChars))
	for _, c := range s {
		if c >= validLen || !validIdentChars[int(c)] {
			return false
		}
	}
	return true
}
