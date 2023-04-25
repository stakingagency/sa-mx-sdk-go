package utils

import (
	"strings"
	"unicode"
)

func ToLowerFirstChar(s string) string {
	r := []rune(s)
	r[0] = unicode.ToLower(r[0])
	return string(r)
}

func ToUpperFirstChar(s string) string {
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func IsList(s string) bool {
	return strings.HasPrefix(s, "List<")
}

func IsSimpleVariadic(s string) bool {
	n := strings.Count(s, "<")

	return strings.HasPrefix(s, "variadic<") && n == 1
}

func IsMultiVariadic(s string) bool {
	return strings.HasPrefix(s, "variadic<multi<") && !IsSimpleVariadic(s)
}

func IsTuple(s string) bool {
	return strings.HasPrefix(s, "tuple<")
}
