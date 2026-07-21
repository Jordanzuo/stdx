// Package base converts uint64 values to and from string representations in custom
// number systems. It supports arbitrary bases defined by a unique character set, plus
// common predefined bases such as binary, octal, hexadecimal, base-26, base-36, and base-62.
//
// Parsing is case-sensitive: every character in the input must match the Base alphabet
// exactly. For example, NewBase16(false) accepts "ff" but rejects "FF"; NewBase16(true)
// accepts "FF" but rejects "ff". Use NewBase62 when both letter cases are valid digits.
//
// Output is not padded with leading zeros. Parse accepts leading zeros ("00ff" and "ff"
// both yield 255), but Transform returns the shortest form ("ff"). Do not use string
// equality on Parse input when comparing encoded values.
//
// NewBase62 treats visually similar characters as distinct digits: 0/O/o and 1/l/I are
// different values. Human readers may confuse them; avoid Base62 for codes that people
// type or read aloud unless the ambiguity is acceptable.
//
// Custom alphabets from New may include Unicode lookalikes (for example Latin "a" and
// Cyrillic "а") as separate digits. Invalid UTF-8 byte sequences are replaced with U+FFFD
// (�) when reading element or source strings.
//
// Base values must be created with New or the NewBase* helpers. A nil *Base or a manually
// constructed Base panics when Transform or Parse is called.
package base

import (
	"fmt"
	"math/bits"

	. "github.com/Jordanzuo/stdx/result"
)

const maxInputLength = 128

// Base represents a number system with a custom set of characters.
// Create instances with New or the NewBase* helpers; do not construct Base manually.
// Methods panic when called on a nil *Base.
type Base struct {
	itemIndexMap map[string]uint64
	indexItemMap map[uint64]string
	base         uint64
}

// New creates a Base from a string of unique characters. The first character represents
// 0, the second represents 1, and so on.
//
// Each Unicode rune in elements becomes one digit. Visually similar characters from
// different scripts are distinct digits. Invalid UTF-8 is replaced with U+FFFD (�).
func New(elements string) Result[*Base] {
	type ResultDataType = *Base

	itemIndexMap := make(map[string]uint64, maxInputLength)
	indexItemMap := make(map[uint64]string, maxInputLength)
	elementCount := 0
	for _, v := range elements {
		elementCount++
		if elementCount > maxInputLength {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: too many elements"))
		}
		index := uint64(elementCount - 1)
		item := string(v)
		if _, exist := itemIndexMap[item]; exist {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: duplicate character %q", item))
		}
		itemIndexMap[item] = index
		indexItemMap[index] = item
	}

	if elementCount == 0 {
		return NewErrorResult[ResultDataType](fmt.Errorf("base: empty elements"))
	}
	if elementCount == 1 {
		return NewErrorResult[ResultDataType](fmt.Errorf("base: single-character base is not supported"))
	}

	return NewDataResult(&Base{
		itemIndexMap: itemIndexMap,
		indexItemMap: indexItemMap,
		base:         uint64(elementCount),
	})
}

// NewBase2 returns a binary (base-2) number system using digits 0 and 1.
func NewBase2() *Base {
	return New("01").Data()
}

// NewBase8 returns an octal (base-8) number system using digits 0 through 7.
func NewBase8() *Base {
	return New("01234567").Data()
}

// NewBase16 returns a hexadecimal (base-16) number system using digits 0-9 and a-f or A-F.
// When uppercase is false, letters are lowercase; when true, letters are uppercase.
// Parse accepts only the selected letter case; mixed or opposite-case input returns an error.
func NewBase16(uppercase bool) *Base {
	elements := "0123456789abcdef"
	if uppercase {
		elements = "0123456789ABCDEF"
	}
	return New(elements).Data()
}

// NewBase26 returns a base-26 number system using letters a-z or A-Z.
// When uppercase is false, letters are lowercase; when true, letters are uppercase.
// Parse accepts only the selected letter case; mixed or opposite-case input returns an error.
func NewBase26(uppercase bool) *Base {
	elements := "abcdefghijklmnopqrstuvwxyz"
	if uppercase {
		elements = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	return New(elements).Data()
}

// NewBase36 returns a base-36 number system using digits 0-9 and letters a-z or A-Z.
// When uppercase is false, letters are lowercase; when true, letters are uppercase.
// Parse accepts only the selected letter case; mixed or opposite-case input returns an error.
func NewBase36(uppercase bool) *Base {
	elements := "0123456789abcdefghijklmnopqrstuvwxyz"
	if uppercase {
		elements = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	}
	return New(elements).Data()
}

// NewBase62 returns a base-62 number system using digits, lowercase letters, and uppercase letters.
// Lowercase and uppercase letters are distinct digits; Parse matches each character exactly.
//
// Several digits look alike to humans but encode different values, including 0/O/o and 1/l/I.
// Prefer another base or a restricted alphabet when users must manually copy encoded strings.
func NewBase62() *Base {
	return New("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ").Data()
}

// Transform converts a decimal uint64 value to its string representation in this base.
// The result has no leading zeros; zero encodes as a single digit (the alphabet's zero symbol).
func (this *Base) Transform(source uint64) string {
	quotient, remainder := uint64(0), source
	encoded := ""

	for {
		quotient, remainder = remainder/this.base, remainder%this.base
		encoded = this.indexItemMap[remainder] + encoded
		if quotient == 0 {
			break
		}
		remainder = quotient
	}

	return encoded
}

// Parse converts a string in this base to its decimal uint64 value.
// Matching is case-sensitive: each rune must appear in the Base alphabet exactly as defined.
// For example, parsing "FF" with NewBase16(false) returns an error, not 255.
//
// Leading zeros are accepted but not preserved: Parse("00ff") and Parse("ff") both return 255,
// while Transform(255) returns "ff". Compare numeric values, not raw strings.
//
// Invalid UTF-8 in source is replaced with U+FFFD (�) before matching the alphabet.
func (this *Base) Parse(source string) Result[uint64] {
	type ResultDataType = uint64

	if source == "" {
		return NewErrorResult[ResultDataType](fmt.Errorf("base: empty source"))
	}
	var value uint64
	runeCount := 0
	for _, v := range source {
		runeCount++
		if runeCount > maxInputLength {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: source too long"))
		}
		item := string(v)
		index, exists := this.itemIndexMap[item]
		if !exists {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: unknown character %q", item))
		}
		hi, lo := bits.Mul64(value, this.base)
		if hi != 0 {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: overflow"))
		}
		var carry uint64
		value, carry = bits.Add64(lo, index, 0)
		if carry != 0 {
			return NewErrorResult[ResultDataType](fmt.Errorf("base: overflow"))
		}
	}

	return NewDataResult(value)
}
