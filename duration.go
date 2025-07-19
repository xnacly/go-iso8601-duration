package goiso8601duration

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// This parser uses the examplary notion of allowing two numbers before any
// designator, again, see: ISO8601 4.4.3.2 Format with designators
const maxNumCount = 2

// From is a FSM, see https://en.wikipedia.org/wiki/Finite-state_machine
//
// For instance, the state transitions for 'PT9M' are:
//
//	| State   | character |
//	| ------- | --------- |
//	| start   | 'P'       |
//	| P       | 'T'       |
//	| T       | '9'       |
//	| Number  | 'M'       |
//	| M       | EOF       |
//	| Fin     |           |
type state = uint8

// In representations of duration,
// the following designators are used as part of the expression,
// see the doc comment of the function
//
// [Y] [M] [W] [D] [H] [M] [S]

const (
	stateStart state = iota
	// start of duration: is used as duration designator, preceding the component which represents the duration;
	stateP

	// seen [W]
	stateWDesignator

	// seen n
	stateNumber
	// seen [Y], [M], [D]
	stateDesignator

	// start of Time: is used as time designator to indicate: the start of the representation of the number of hours, minutes or seconds in expressions of duration
	stateT
	// seen n
	stateTNumber
	// seen [H], [M], [S]
	stateTDesignator

	stateFin
)

var stateToName = map[state]string{
	stateStart:       "Start",
	stateP:           "P",
	stateWDesignator: "WDesignator",
	stateNumber:      "Number",
	stateT:           "T",
	stateTNumber:     "TNumber",
	stateTDesignator: "TDesignator",
	stateFin:         "Fin",
}

type ISO8601Duration struct {
	year, month, week, day, hour, minute, second float64
}

func numBufferToNumber(buf [maxNumCount]rune) int64 {
	var i int
	for _, n := range buf {
		i = (i * 10) + int(n-'0')
	}
	return int64(i)
}

// P[nn]Y[nn]M[nn]DT[nn]H[nn]M[nn]S, P[nn]W, P<date>T<time>, as seen in
// https://en.wikipedia.org/wiki/ISO_8601#Durations
//
// - P is the duration designator (for period) placed at the start of the duration representation.
//   - Y is the year designator that follows the value for the number of calendar years.
//   - M is the month designator that follows the value for the number of calendar months.
//   - W is the week designator that follows the value for the number of weeks.
//   - D is the day designator that follows the value for the number of calendar days.
//
// - T is the time designator that precedes the time components of the representation.
//   - H is the hour designator that follows the value for the number of hours.
//   - M is the minute designator that follows the value for the number of minutes.
//   - S is the second designator that follows the value for the number of seconds.
func From(s string) (ISO8601Duration, error) {
	var duration ISO8601Duration

	if len(s) == 0 {
		return duration, wrapErr(UnexpectedEof, 0)
	}

	curState := stateStart
	var col uint8

	r := strings.NewReader(s)

	for {
		b, size, err := r.ReadRune()
		fmt.Printf("| rune=%c | col=%d | state=%s |\n", b, col, stateToName[curState])
		if err != nil {
			if err != io.EOF {
				return duration, errors.Join(UnexpectedReaderError, err)
			} else {
				curState = stateFin
			}
		}
		if size > 1 {
			return duration, wrapErr(UnexpectedNonAsciiRune, col)
		}
		col++

		// TODO: other states
		switch curState {
		case stateStart:
			if b != 'P' {
				return duration, wrapErr(MissingPDesignatorAtStart, col)
			}
			curState = stateP
		case stateFin:
			return duration, nil
		}
	}
}

func (i ISO8601Duration) Apply(t time.Time) time.Time {
	newT := t.AddDate(int(i.year), int(i.month), int(i.day))
	d := time.Duration(
		(i.hour * float64(time.Hour)) +
			(i.minute * float64(time.Minute)) +
			(i.second * float64(time.Second)),
	)
	return newT.Add(d)
}

func (i ISO8601Duration) Time() time.Time {
	return time.Now()
}

func (i ISO8601Duration) Duration() time.Duration {
	var ns int64
	return time.Duration(ns)
}

func (i ISO8601Duration) String() string {
	b := strings.Builder{}
	b.WriteRune('P')

	// If the number of years, months, days, hours, minutes or seconds in any of these expressions equals
	// zero, the number and the corresponding designator may be absent; however, at least one number
	// and its designator shall be present
	if i.year == 0 && i.month == 0 && i.week == 0 && i.day == 0 && i.hour == 0 && i.minute == 0 && i.second == 0 {
		b.WriteString("0D")
		return b.String()
	}

	if i.week > 0 {
		b.WriteString(strconv.FormatFloat(i.week, 'g', -1, 64))
		b.WriteRune('W')
		return b.String()
	}

	if i.year > 0 {
		b.WriteString(strconv.FormatFloat(i.year, 'g', -1, 64))
		b.WriteRune('Y')
	}
	if i.month > 0 {
		b.WriteString(strconv.FormatFloat(i.month, 'g', -1, 64))
		b.WriteRune('M')
	}
	if i.day > 0 {
		b.WriteString(strconv.FormatFloat(i.day, 'g', -1, 64))
		b.WriteRune('D')
	}

	// The designator [T] shall be absent if all of the time components are absent.
	if i.hour > 0 || i.minute > 0 || i.second > 0 {
		b.WriteRune('T')

		if i.hour > 0 {
			b.WriteString(strconv.FormatFloat(i.hour, 'g', -1, 64))
			b.WriteRune('H')
		}

		if i.minute > 0 {
			b.WriteString(strconv.FormatFloat(i.minute, 'g', -1, 64))
			b.WriteRune('M')
		}

		if i.second > 0 {
			b.WriteString(strconv.FormatFloat(i.second, 'g', -1, 64))
			b.WriteRune('S')
		}
	}

	return b.String()
}
