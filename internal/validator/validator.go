package validator

import (
	"regexp"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsEmail(s string) bool {
	return emailRegex.MatchString(strings.TrimSpace(s))
}

func NotEmpty(s string) bool {
	return strings.TrimSpace(s) != ""
}

func MinLen(s string, n int) bool {
	return len(strings.TrimSpace(s)) >= n
}

func OneOf(s string, valid ...string) bool {
	for _, v := range valid {
		if s == v {
			return true
		}
	}
	return false
}
