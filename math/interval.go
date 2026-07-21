// Package math provides discrete intervals over unsigned integer types.
//
// An Interval has a lower bound, an upper bound, and an IntervalType that controls
// whether each endpoint is inclusive or exclusive. BuildInterval and
// BuildIntervalFromString return result.Result values; check HasError before calling Data.
//
// BuildInterval rejects intervals when the lower bound exceeds the upper bound, when
// equal bounds would form an empty interval with any exclusive endpoint, when
// [0,max] inclusive would overflow Length, or when the interval type is undefined.
// BuildIntervalFromString validates bracket syntax, converts each bound with convertFunc,
// and delegates ordering and interval-type rules to BuildInterval.
//
// String and BuildIntervalFromString use comma-separated bounds (for example "[1,5]").
// DistanceFromBeginning returns value minus the lower bound when the value is in the
// interval, or 0 otherwise.
//
// Create intervals with BuildInterval or BuildIntervalFromString rather than
// constructing Interval values directly.
package math

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	. "github.com/Jordanzuo/stdx/result"
)

// IntervalType selects which endpoints of an interval are inclusive or exclusive.
type IntervalType uint8

// Interval boundary combinations.
const (
	InclusiveInclusive IntervalType = 1 // [a, b]
	InclusiveExclusive IntervalType = 2 // [a, b)
	ExclusiveInclusive IntervalType = 3 // (a, b]
	ExclusiveExclusive IntervalType = 4 // (a, b)
)

// Bracket and delimiter literals used when parsing or formatting interval strings.
const (
	InclusiveOpenBracket  = "["
	InclusiveCloseBracket = "]"
	ExclusiveOpenBracket  = "("
	ExclusiveCloseBracket = ")"
	IntervalDelimiter     = ","
)

var (
	// intervalFormatRegex matches any supported bracket form with comma-separated bounds.
	intervalFormatRegex     = regexp.MustCompile(fmt.Sprintf("^[%s%s][[:digit:]]+%s[[:digit:]]+[%s%s]$", InclusiveOpenBracket, ExclusiveOpenBracket, IntervalDelimiter, InclusiveCloseBracket, ExclusiveCloseBracket))
	inclusiveInclusiveRegex = regexp.MustCompile(fmt.Sprintf("^[%s][[:digit:]]+%s[[:digit:]]+[%s]$", InclusiveOpenBracket, IntervalDelimiter, InclusiveCloseBracket))
	inclusiveExclusiveRegex = regexp.MustCompile(fmt.Sprintf("^[%s][[:digit:]]+%s[[:digit:]]+[%s]$", InclusiveOpenBracket, IntervalDelimiter, ExclusiveCloseBracket))
	exclusiveInclusiveRegex = regexp.MustCompile(fmt.Sprintf("^[%s][[:digit:]]+%s[[:digit:]]+[%s]$", ExclusiveOpenBracket, IntervalDelimiter, InclusiveCloseBracket))
	exclusiveExclusiveRegex = regexp.MustCompile(fmt.Sprintf("^[%s][[:digit:]]+%s[[:digit:]]+[%s]$", ExclusiveOpenBracket, IntervalDelimiter, ExclusiveCloseBracket))
)

// Interval is a bounded range of unsigned integers with configurable endpoint inclusion.
//
// Lower must not exceed Upper. Create instances with BuildInterval or
// BuildIntervalFromString rather than constructing Interval values directly.
type Interval[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr] struct {
	Lower T            // Lower bound.
	Upper T            // Upper bound.
	Type  IntervalType // Endpoint inclusion.
}

// Contains reports whether value lies within the interval, respecting endpoint inclusion.
func (this Interval[T]) Contains(value T) bool {
	switch this.Type {
	case InclusiveInclusive:
		return this.Lower <= value && value <= this.Upper
	case InclusiveExclusive:
		return this.Lower <= value && value < this.Upper
	case ExclusiveInclusive:
		return this.Lower < value && value <= this.Upper
	case ExclusiveExclusive:
		return this.Lower < value && value < this.Upper
	default:
		panic("unreachable")
	}
}

// Length returns the number of integer values in the interval:
// [a,b] is upper-lower+1, half-open intervals are upper-lower, and (a,b) is upper-lower-1.
func (this Interval[T]) Length() T {
	switch this.Type {
	case InclusiveInclusive:
		return this.Upper - this.Lower + 1
	case InclusiveExclusive:
		return this.Upper - this.Lower
	case ExclusiveInclusive:
		return this.Upper - this.Lower
	case ExclusiveExclusive:
		return this.Upper - this.Lower - 1
	default:
		panic("unreachable")
	}
}

// DistanceFromBeginning returns value minus the lower bound when value is in the interval,
// or 0 when value is not in the interval.
func (this Interval[T]) DistanceFromBeginning(value T) T {
	if !this.Contains(value) {
		return 0
	}
	return value - this.Lower
}

// String formats the interval using bracket notation and a comma between bounds.
// Examples: [1,5], [2,6), (3,7], (4,8).
func (this Interval[T]) String() string {
	switch this.Type {
	case InclusiveInclusive:
		return fmt.Sprintf("%s%d%s%d%s", InclusiveOpenBracket, this.Lower, IntervalDelimiter, this.Upper, InclusiveCloseBracket)
	case InclusiveExclusive:
		return fmt.Sprintf("%s%d%s%d%s", InclusiveOpenBracket, this.Lower, IntervalDelimiter, this.Upper, ExclusiveCloseBracket)
	case ExclusiveInclusive:
		return fmt.Sprintf("%s%d%s%d%s", ExclusiveOpenBracket, this.Lower, IntervalDelimiter, this.Upper, InclusiveCloseBracket)
	case ExclusiveExclusive:
		return fmt.Sprintf("%s%d%s%d%s", ExclusiveOpenBracket, this.Lower, IntervalDelimiter, this.Upper, ExclusiveCloseBracket)
	default:
		panic("unreachable")
	}
}

// BuildInterval constructs an interval from bounds and an IntervalType.
//
// It returns an error when lower exceeds upper, when equal bounds would form an empty
// interval with any exclusive endpoint, when [0,max] inclusive would overflow Length,
// or when _type is not a defined IntervalType constant.
func BuildInterval[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](lower, upper T, _type IntervalType) Result[Interval[T]] {
	switch _type {
	case InclusiveInclusive, InclusiveExclusive, ExclusiveInclusive, ExclusiveExclusive:
	default:
		return NewErrorResult[Interval[T]](fmt.Errorf("math: invalid interval type"))
	}

	if lower > upper {
		return NewErrorResult[Interval[T]](fmt.Errorf("math: lower bound exceeds upper bound"))
	}
	if _type == InclusiveInclusive && lower == 0 && upper == ^T(0) {
		return NewErrorResult[Interval[T]](fmt.Errorf("math: interval length overflow"))
	}
	if _type != InclusiveInclusive && lower == upper {
		return NewErrorResult[Interval[T]](fmt.Errorf("math: empty interval"))
	}

	return NewDataResult(Interval[T]{
		Lower: lower,
		Upper: upper,
		Type:  _type,
	})
}

// BuildIntervalFromString parses an interval string and builds an Interval.
//
// Input must match [lower,upper], [lower,upper), (lower,upper], or (lower,upper) with a
// comma between bounds. convertFunc is called once for the lower substring and once for
// the upper substring.
func BuildIntervalFromString[T ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr](input string, convertFunc func(string) Result[T]) Result[Interval[T]] {
	if len(input) == 0 {
		return NewErrorResult[Interval[T]](errors.New("interval string is empty"))
	}
	if !intervalFormatRegex.MatchString(input) {
		return NewErrorResult[Interval[T]](errors.New("invalid interval format"))
	}

	var intervalType IntervalType
	switch {
	case inclusiveInclusiveRegex.MatchString(input):
		intervalType = InclusiveInclusive
	case inclusiveExclusiveRegex.MatchString(input):
		intervalType = InclusiveExclusive
	case exclusiveInclusiveRegex.MatchString(input):
		intervalType = ExclusiveInclusive
	case exclusiveExclusiveRegex.MatchString(input):
		intervalType = ExclusiveExclusive
	default:
		panic("unreachable")
	}

	input = strings.TrimPrefix(input, InclusiveOpenBracket)
	input = strings.TrimPrefix(input, ExclusiveOpenBracket)
	input = strings.TrimSuffix(input, InclusiveCloseBracket)
	input = strings.TrimSuffix(input, ExclusiveCloseBracket)
	inputList := strings.Split(input, IntervalDelimiter)
	lowerResult := convertFunc(inputList[0])
	if lowerResult.HasError() {
		return NewErrorResult[Interval[T]](lowerResult.Error())
	}
	upperResult := convertFunc(inputList[1])
	if upperResult.HasError() {
		return NewErrorResult[Interval[T]](upperResult.Error())
	}

	return BuildInterval(lowerResult.Data(), upperResult.Data(), intervalType)
}
