package goiso8601duration

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDurationStringer(t *testing.T) {
	cases := []struct {
		expected string
		c        ISO8601Duration
	}{
		{"P0D", ISO8601Duration{}},
		{"PT15H", ISO8601Duration{hour: 15}},
		{"P1W", ISO8601Duration{week: 1}},
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

	for _, i := range cases {
		stringified := i.c.String()
		assert.Equal(t, i.expected, stringified)
	}

}

func TestDurationErr(t *testing.T) {
	cases := []string{
		"",
		"Z",
	}

	for _, i := range cases {
		_, err := From(i)
		assert.Error(t, err)
	}
}

func TestDuration(t *testing.T) {
	cases := []struct {
		c        string
		expected ISO8601Duration
	}{
		{"PT15M", ISO8601Duration{minute: 15}},
		{"PT15M10S", ISO8601Duration{minute: 15, second: 10}},
		{"P15W", ISO8601Duration{week: 15}},
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

	for _, i := range cases {
		t.Run(i.c, func(t *testing.T) {
			parsed, err := From(i.c)
			assert.NoError(t, err)
			assert.Equal(t, i.expected, parsed)
		})
	}
}
