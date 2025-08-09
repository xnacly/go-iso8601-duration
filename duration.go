package goiso8601duration

import (
	"bytes"
	"encoding/json"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// constants for roundtripping between time.Duration and Duration
const (
	nsPerSecond  = int64(time.Second)
	nsPerMinute  = int64(time.Minute)
	nsPerHour    = int64(time.Hour)
	nsPerDay     = int64(24 * time.Hour)
	nsPerWeek    = int64(7 * 24 * time.Hour)
	daysPerYear  = 365.2425
	daysPerMonth = 30.436875
)

// From is a FSM, see https://en.wikipedia.org/wiki/Finite-state_machine
type state = uint8

// In representations of duration,
// the following designators are used as part of the expression,
// see the doc comment of the function
//
// [Y] [M] [W] [D] [H] [M] [S]
const (
	defaultDesignators = "YMWD"
	timeDesignators    = "MHS"
)

const (
	stateStart state = iota
	// start of duration: is used as duration designator, preceding the component which represents the duration;
	stateP

	// seen n
	stateNumber
	// seen [Y], [W], [M], [D]
	stateDesignator

	// start of Time: is used as time designator to indicate: the start of the representation of the number of hours, minutes or seconds in expressions of duration
	stateT
	// seen n
	stateTNumber
	// seen [H], [M], [S]
	stateTDesignator

	stateFin
)

type Duration struct {
	hasNegativeSign                              bool
	year, month, week, day, hour, minute, second int64
}

func FromDuration(d time.Duration) Duration {
	ns := d.Nanoseconds()
	duration := Duration{}

	years := float64(ns) / (float64(nsPerDay) * daysPerYear)
	duration.year = int64(years) // truncate to integer years
	ns -= int64(float64(duration.year) * daysPerYear * float64(nsPerDay))

	months := float64(ns) / (float64(nsPerDay) * daysPerMonth)
	duration.month = int64(months)
	ns -= int64(float64(duration.month) * daysPerMonth * float64(nsPerDay))

	weeks := ns / nsPerWeek
	duration.week = int64(weeks)
	ns -= weeks * nsPerWeek

	days := ns / nsPerDay
	duration.day = int64(days)
	ns -= days * nsPerDay

	hours := ns / nsPerHour
	duration.hour = int64(hours)
	ns -= hours * nsPerHour

	minutes := ns / nsPerMinute
	duration.minute = int64(minutes)
	ns -= minutes * nsPerMinute

	duration.second = int64(float64(ns) / float64(nsPerSecond))

	return duration
}

func numBufferToNumber(buf bytes.Buffer) (int64, error) {
	var i int64
	for _, n := range buf.Bytes() {
		digit := int64(n - '0')
		if i > (math.MaxInt64-digit)/10 {
			return 0, DesignatorNumberTooLarge
		}
		i = (i * 10) + digit
	}

	return i, nil
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
func From(s string) (Duration, error) {
	var duration Duration

	if len(s) == 0 {
		return duration, wrapErr(UnexpectedEof, 0)
	}

	curState := stateStart
	var col uint8
	numBuf := bytes.Buffer{}

	r := strings.NewReader(s)

	for {
		b, size, err := r.ReadRune()
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
			switch b {
			case '-':
				duration.hasNegativeSign = true
				curState = stateStart
			case '+':
				curState = stateStart
			case 'P':
				curState = stateP
			default:
				return duration, wrapErr(MissingPDesignatorAtStart, col)
			}
		case stateP, stateDesignator:
			if b == 'T' {
				curState = stateT
			} else if unicode.IsDigit(b) {
				numBuf.WriteRune(b)
				curState = stateNumber
			} else {
				return duration, wrapErr(MissingNumber, col)
			}
		case stateNumber:
			if unicode.IsDigit(b) {
				numBuf.WriteRune(b)
				curState = stateNumber
			} else if strings.ContainsRune(defaultDesignators, b) {
				if numBuf.Len() == 0 {
					return duration, wrapErr(MissingNumber, col)
				}
				num, err := numBufferToNumber(numBuf)
				if err != nil {
					return duration, err
				}
				switch b {
				case 'Y':
					if duration.year != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.year = num
				case 'M':
					if duration.month != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.month = num
				case 'W':
					if duration.week != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.week = num
				case 'D':
					if duration.day != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.day = num
				}
				numBuf.Reset()
				curState = stateDesignator
			} else {
				return duration, wrapErr(UnknownDesignator, col)
			}
		case stateT, stateTDesignator:
			if unicode.IsDigit(b) {
				numBuf.WriteRune(b)
				curState = stateTNumber
			} else {
				return duration, wrapErr(MissingNumber, col)
			}
		case stateTNumber:
			if unicode.IsDigit(b) {
				numBuf.WriteRune(b)
				curState = stateTNumber
			} else if strings.ContainsRune(timeDesignators, b) {
				if numBuf.Len() == 0 {
					return duration, wrapErr(MissingNumber, col)
				}
				num, err := numBufferToNumber(numBuf)
				if err != nil {
					return duration, err
				}
				switch b {
				case 'H':
					if duration.hour != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.hour = num
				case 'M':
					if duration.minute != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.minute = num
				case 'S':
					if duration.second != 0 {
						return duration, wrapErr(DuplicateDesignator, col)
					}
					duration.second = num
				}
				numBuf.Reset()
				curState = stateTDesignator
			} else {
				return duration, wrapErr(UnknownDesignator, col)
			}
		case stateFin:
			return duration, nil
		}
	}
}

func (i Duration) Apply(t time.Time) time.Time {
	newT := t.AddDate(int(i.year), int(i.month), int(i.day+i.week*7))
	d := time.Duration(
		(i.hour * int64(time.Hour)) +
			(i.minute * int64(time.Minute)) +
			(i.second * int64(time.Second)),
	)
	if i.hasNegativeSign {
		d = -d
	}
	return newT.Add(d)
}

func (i Duration) Duration() time.Duration {
	var ns int64

	ns += int64(float64(i.year) * daysPerYear * float64(nsPerDay))
	ns += int64(float64(i.month) * daysPerMonth * float64(nsPerDay))
	ns += int64(float64(i.week) * float64(nsPerWeek))
	ns += int64(float64(i.day) * float64(nsPerDay))
	ns += int64(float64(i.hour) * float64(nsPerHour))
	ns += int64(float64(i.minute) * float64(nsPerMinute))
	ns += int64(float64(i.second) * float64(nsPerSecond))

	if i.hasNegativeSign {
		ns = -ns
	}

	return time.Duration(ns)
}

func (i Duration) String() string {
	b := strings.Builder{}
	if i.hasNegativeSign {
		b.WriteByte('-')
	}
	b.WriteByte('P')

	// If the number of years, months, days, hours, minutes or seconds in any of these expressions equals
	// zero, the number and the corresponding designator may be absent; however, at least one number
	// and its designator shall be present
	if i.year == 0 && i.month == 0 && i.week == 0 && i.day == 0 && i.hour == 0 && i.minute == 0 && i.second == 0 {
		b.WriteString("0D")
		return b.String()
	}

	if i.year > 0 {
		b.WriteString(strconv.FormatInt(i.year, 10))
		b.WriteByte('Y')
	}
	if i.month > 0 {
		b.WriteString(strconv.FormatInt(i.month, 10))
		b.WriteByte('M')
	}
	if i.week > 0 {
		b.WriteString(strconv.FormatInt(i.week, 10))
		b.WriteByte('W')
	}
	if i.day > 0 {
		b.WriteString(strconv.FormatInt(i.day, 10))
		b.WriteByte('D')
	}

	// The designator [T] shall be absent if all of the time components are absent.
	if i.hour > 0 || i.minute > 0 || i.second > 0 {
		b.WriteByte('T')

		if i.hour > 0 {
			b.WriteString(strconv.FormatInt(i.hour, 10))
			b.WriteByte('H')
		}

		if i.minute > 0 {
			b.WriteString(strconv.FormatInt(i.minute, 10))
			b.WriteByte('M')
		}

		if i.second > 0 {
			b.WriteString(strconv.FormatInt(i.second, 10))
			b.WriteByte('S')
		}
	}

	return b.String()
}

func (i Duration) MarshalJSON() ([]byte, error) {
	return json.Marshal(i.String())
}

func (i *Duration) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	d, err := From(s)
	if err != nil {
		return err
	}
	*i = d

	return nil
}
