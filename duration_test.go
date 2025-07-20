package goiso8601duration

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestReadme(t *testing.T) {
	rawDuration := "PT1H30M12S"
	duration, err := From(rawDuration)
	if err != nil {
		panic(duration)
	}

	fmt.Println(duration.Duration().String(), duration.String())
	fmt.Println(
		time.
			Unix(0, 0).
			Format(time.TimeOnly),
		duration.
			Apply(time.Unix(0, 0)).
			Format(time.TimeOnly),
	)
}

var testcases = []struct {
	str string
	dur ISO8601Duration
}{
	{"P0D", ISO8601Duration{}},
	{"PT15H", ISO8601Duration{hour: 15}},
	{"P1W", ISO8601Duration{week: 1}},
	{"P15W", ISO8601Duration{week: 15}},
	{"P15Y", ISO8601Duration{year: 15}},
	{"P15Y3M", ISO8601Duration{year: 15, month: 3}},
	{"P15Y3M41D", ISO8601Duration{year: 15, month: 3, day: 41}},
	{"PT15M", ISO8601Duration{minute: 15}},
	{"PT15M10S", ISO8601Duration{minute: 15, second: 10}},
	{
		"P3Y6M4DT12H30M5S",
		ISO8601Duration{
			year:   3,
			month:  6,
			day:    4,
			hour:   12,
			minute: 30,
			second: 5,
		},
	},
}

func TestDurationStringer(t *testing.T) {
	for _, i := range testcases {
		t.Run(i.str, func(t *testing.T) {
			stringified := i.dur.String()
			assert.Equal(t, i.str, stringified)
		})
	}
}

func TestDuration(t *testing.T) {
	for _, i := range testcases {
		t.Run(i.str, func(t *testing.T) {
			parsed, err := From(i.str)
			assert.NoError(t, err)
			assert.Equal(t, i.dur, parsed)
		})
	}
}

func TestBiliteral(t *testing.T) {
	for _, i := range testcases {
		t.Run(i.str, func(t *testing.T) {
			parsed, err := From(i.str)
			assert.NoError(t, err)
			assert.Equal(t, i.dur, parsed)
			stringified := parsed.String()
			assert.Equal(t, i.str, stringified)
		})
	}
}

// TestDurationErr makes sure all expected edgecases are implemented correctly
func TestDurationErr(t *testing.T) {
	cases := []string{
		"",        // UnexpectedEof
		"P",       // UnexpectedEof
		"è",       // UnexpectedNonAsciiRune
		"P1",      // MissingDesignator
		"P1A",     // UnknownDesignator
		"P12D12D", // DuplicateDesignator
		"P1YD",    // MissingNumber
		"P111Y",   // TooManyNumbersForDesignator
		"Z",       // MissingPDesignatorAtStart
		"P15W2D",  // NoDesignatorsAfterWeeksAllowed
	}

	for _, i := range cases {
		t.Run(i, func(t *testing.T) {
			_, err := From(i)
			assert.Error(t, err)
		})
	}
}
