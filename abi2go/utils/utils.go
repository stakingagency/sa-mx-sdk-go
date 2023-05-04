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

func IsMultiVariadic(s string, multiResult ...bool) bool {
	isMultiVariadic := strings.HasPrefix(s, "variadic<multi<") && !IsSimpleVariadic(s)
	if len(multiResult) > 0 && multiResult[0] {
		isMultiVariadic = isMultiVariadic || strings.HasPrefix(s, "variadic<")
	}

	return isMultiVariadic
}

func IsMulti(s string) bool {
	return strings.HasPrefix(s, "multi<") && !IsSimpleVariadic(s)
}

func IsTuple(s string) bool {
	return strings.HasPrefix(s, "tuple<")
}

func SplitTypes(s string) []string {
	res := make([]string, 0)
	i := 0
	insideTuple := false
	for {
		switch s[i] {
		case ',':
			if !insideTuple {
				res = append(res, string([]byte(s)[:i]))
				s = string([]byte(s)[i+1:])
				i = -1
			}
		case '<':
			insideTuple = true
		case '>':
			insideTuple = false
		}
		i++
		if i == len(s) {
			res = append(res, s)
			break
		}
	}

	return res
}
