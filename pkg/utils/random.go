package utils

import (
	"math/rand/v2"
	"strings"
)

type CharSet string

const (
	Digits     CharSet = "0123456789"
	AlphaUpper CharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	AlphaLower CharSet = "abcdefghijklmnopqrstuvwxyz"
	AlphaAll   CharSet = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
)

// RandomCode generates a random string of the given length from the given character sets.
// Uses math/rand/v2 top-level functions which are concurrency-safe.
func RandomCode(length int, charset ...CharSet) string {
	var chars string
	for _, c := range charset {
		chars += string(c)
	}
	code := make([]byte, length)
	for i := range code {
		code[i] = chars[rand.IntN(len(chars))]
	}
	return string(code)
}

func RandomUserCode() string {
	code := RandomCode(16, AlphaUpper, Digits)
	var result strings.Builder
	for i := 0; i < len(code); i += 4 {
		if i > 0 {
			result.WriteString("-")
		}
		result.WriteString(code[i : i+4])
	}
	return result.String()
}
