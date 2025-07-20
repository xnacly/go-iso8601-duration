package goiso8601duration

import (
	"io"
	"strconv"
	"strings"
	"time"
	"unicode"
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
const defaultDesignators = "YMWD"
const timeDesignators = "MHS"

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

type ISO8601Duration struct {
	year, month, week, day, hour, minute, second float64
}

func numBufferToNumber(buf [maxNumCount]rune) int64 {
	var i int
	for _, n := range buf {
		if n == 0 { // empty number (zero byte) in buffer, stop
			break
		}
		i = (i * 10) + int(n-'0')
	}

	return int64(i)
}

// P[nn]Y[nn]M[nn]DT[nn]H[nn]M[nn]S or P[nn]W, as seen in
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
	var curNumCount uint8
	var numBuf [maxNumCount]rune

	r := strings.NewReader(s)

	for {
		b, size, err := r.ReadRune()

		// This is for debugging purposes
		// var stateToName = map[state]string{
		// 	stateStart:       "Start",
		// 	stateP:           "P",
		// 	stateWDesignator: "WDesignator",
		// 	stateNumber:      "Number",
		// 	stateT:           "T",
		// 	stateTNumber:     "TNumber",
		// 	stateTDesignator: "TDesignator",
		// 	stateFin:         "Fin",
		// }
		// fmt.Printf("| rune=%c | col=%d | state=%s | buf=%v\n", b, col, stateToName[curState], numBuf)

		if err != nil {
			if err != io.EOF {
				return duration, wrapErr(UnexpectedReaderError, col)
			} else if curState == stateP {
				// being in stateP at the end (io.EOF) means we havent
				// encountered anything after the P, so there were no numbers
				// or states
				return duration, wrapErr(UnexpectedEof, col)
			} else if curState == stateNumber || curState == stateTNumber {
				// if we are in the state of Number or TNumber we had a number
				// but no designator at the end
				return duration, wrapErr(MissingDesignator, col)
			} else {
				curState = stateFin
			}
		}
		if size > 1 {
			return duration, wrapErr(UnexpectedNonAsciiRune, col)
		}
		col++

		switch curState {
		case stateStart:
			if b != 'P' {
				return duration, wrapErr(MissingPDesignatorAtStart, col)
			}
			curState = stateP
		case stateP, stateDesignator:
			if b == 'T' {
				curState = stateT
			} else if unicode.IsDigit(b) {
				if curNumCount > maxNumCount {
					return duration, wrapErr(TooManyNumbersForDesignator, col)
				}
				numBuf[curNumCount] = b
				curNumCount++
				curState = stateNumber
			} else {
				return duration, wrapErr(MissingNumber, col)
			}
		case stateNumber:
			if unicode.IsDigit(b) {
				if curNumCount+1 > maxNumCount {
					return duration, wrapErr(TooManyNumbersForDesignator, col)
				}
				numBuf[curNumCount] = b
				curNumCount++
				curState = stateNumber
			} else if strings.ContainsRune(defaultDesignators, b) {
				if curNumCount == 0 {
					return duration, wrapErr(MissingNumber, col)
				}
				num := numBufferToNumber(numBuf)
				switch b {
				case 'Y':
					if duration.year != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.year = float64(num)
				case 'M':
					if duration.month != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.month = float64(num)
				case 'W':
					if r.Len() != 0 {
						return duration, wrapErr(NoDesignatorsAfterWeeksAllowed, col)
					}
					duration.week = float64(num)
				case 'D':
					if duration.day != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.day = float64(num)
				}
				curNumCount = 0
				numBuf = [maxNumCount]rune{}
				curState = stateDesignator
			} else {
				return duration, wrapErr(UnknownDesignator, col)
			}
		case stateT, stateTDesignator:
			if unicode.IsDigit(b) {
				if curNumCount > maxNumCount {
					return duration, wrapErr(TooManyNumbersForDesignator, col)
				}
				numBuf[curNumCount] = b
				curNumCount++
				curState = stateTNumber
			} else {
				return duration, wrapErr(MissingNumber, col)
			}
		case stateTNumber:
			if unicode.IsDigit(b) {
				if curNumCount+1 > maxNumCount {
					return duration, wrapErr(TooManyNumbersForDesignator, col)
				}
				numBuf[curNumCount] = b
				curNumCount++
				curState = stateTNumber
			} else if strings.ContainsRune(timeDesignators, b) {
				if curNumCount == 0 {
					return duration, wrapErr(MissingNumber, col)
				}
				num := numBufferToNumber(numBuf)
				switch b {
				case 'H':
					if duration.hour != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.hour = float64(num)
				case 'M':
					if duration.minute != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.minute = float64(num)
				case 'S':
					if duration.second != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.second = float64(num)
				}
				curNumCount = 0
				numBuf = [maxNumCount]rune{}
				curState = stateTDesignator
			} else {
				return duration, wrapErr(UnknownDesignator, col)
			}
		case stateFin:
			return duration, nil
		}
	}
}

func (i ISO8601Duration) Apply(t time.Time) time.Time {
	newT := t.AddDate(int(i.year), int(i.month), int(i.day+i.week*7))
	d := time.Duration(
		(i.hour * float64(time.Hour)) +
			(i.minute * float64(time.Minute)) +
			(i.second * float64(time.Second)),
	)
	return newT.Add(d)
}

func (i ISO8601Duration) Duration() time.Duration {
	const (
		nsPerSecond  = int64(time.Second)
		nsPerMinute  = int64(time.Minute)
		nsPerHour    = int64(time.Hour)
		nsPerDay     = int64(24 * time.Hour)
		nsPerWeek    = int64(7 * 24 * time.Hour)
		daysPerYear  = 365.2425
		daysPerMonth = 30.436875
	)

	var ns int64

	ns += int64(i.year * daysPerYear * float64(nsPerDay))
	ns += int64(i.month * daysPerMonth * float64(nsPerDay))
	ns += int64(i.week * float64(nsPerWeek))
	ns += int64(i.day * float64(nsPerDay))
	ns += int64(i.hour * float64(nsPerHour))
	ns += int64(i.minute * float64(nsPerMinute))
	ns += int64(i.second * float64(nsPerSecond))

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
