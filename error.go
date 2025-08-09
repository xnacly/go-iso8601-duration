package goiso8601duration

import (
	"errors"
	"fmt"
)

var (
	UnexpectedEof               = errors.New("Unexpected EOF in duration format string")
	UnexpectedReaderError       = errors.New("Failed to retrieve next byte of duration format string")
	UnexpectedNonAsciiRune      = errors.New("Unexpected non ascii component in duration format string")
	MissingDesignator           = errors.New("Missing unit designator")
	UnknownDesignator           = errors.New("Unknown designator, expected YMWD or after a T, HMS")
	DuplicateDesignator         = errors.New("Duplicate designator")
	MissingNumber               = errors.New("Missing number specifier before unit designator")
	TooManyNumbersForDesignator = errors.New("Only 2 numbers before any designator allowed")
	MissingPDesignatorAtStart   = errors.New("Missing [P] designator at the start of the duration format string")
)

type ISO8601DurationError struct {
	Inner  error
	Column uint8
}

func wrapErr(inner error, col uint8) error {
	return ISO8601DurationError{
		Inner:  inner,
		Column: col,
	}
}

func (i ISO8601DurationError) String() string {
	return fmt.Sprint("ISO8601DurationError: ", i.Inner, ", at col: ", i.Column)
}

func (i ISO8601DurationError) Error() string {
	return i.String()
}
