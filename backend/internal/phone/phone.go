package phone

import "strings"

func Normalize(raw string) string {
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	s := digits.String()
	if len(s) == 10 {
		return "1" + s
	}
	if len(s) >= 11 && s[0] == '1' {
		return s[:11]
	}
	if len(s) > 11 {
		return "1" + s[len(s)-10:]
	}
	if len(s) > 0 {
		return "1" + s
	}
	return s
}

func IsComplete(raw string) bool {
	n := Normalize(raw)
	return len(n) == 11 && n[0] == '1'
}

func Validate(raw string) bool {
	n := Normalize(raw)
	return len(n) == 11
}