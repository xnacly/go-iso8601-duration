package goiso8601duration

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func TestReadme(t *testing.T) {
	rawDuration := "PT1H30M12S"
	duration := Must(From(rawDuration))

	// 1h30m12s PT1H30M12S
	fmt.Println(duration.Duration().String(), duration.String())

	// 01:00:00 02:30:12
	fmt.Println(
		time.
			Unix(0, 0).
			Format(time.TimeOnly),
		duration.
			Apply(time.Unix(0, 0)).
			Format(time.TimeOnly),
	)

	type arrival struct {
		In Duration `json:"in"`
	}

	asJson := Must(
		json.Marshal(
			arrival{
				In: FromDuration(
					12*time.Minute + 43*time.Second,
				),
			},
		),
	)

	// {"in":"PT12M43S"}
	fmt.Println(string(asJson))

	var a arrival
	json.Unmarshal(asJson, &a)

	// {In:PT12M43S}
	fmt.Printf("%+v\n", a)
}

var testcases = []struct {
	str string
	dur Duration
}{
	{"P0D", Duration{}},
	{"PT15H", Duration{hour: 15}},
	{"P1W", Duration{week: 1}},
	{"P15W", Duration{week: 15}},
	{"P15Y", Duration{year: 15}},
	{"P15Y3M", Duration{year: 15, month: 3}},
	{"P15Y3M41D", Duration{year: 15, month: 3, day: 41}},
	{"PT15M", Duration{minute: 15}},
	{"PT15M10S", Duration{minute: 15, second: 10}},
	{
		"P3Y6M4DT12H30M5S",
		Duration{
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

func TestJSONMarshalUnmarshal(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.str, func(t *testing.T) {
			data, err := json.Marshal(tc.dur)
			assert.NoError(t, err)

			expectedJSON := `"` + tc.str + `"`
			assert.Equal(t, expectedJSON, string(data))

			var unmarshaled Duration
			err = json.Unmarshal(data, &unmarshaled)
			assert.NoError(t, err)
			assert.Equal(t, tc.dur, unmarshaled)
		})
	}
}

func TestDurationRoundtrip(t *testing.T) {
	for _, tc := range testcases {
		t.Run(tc.str, func(t *testing.T) {
			asTimeDuration := tc.dur.Duration()
			asDur := FromDuration(asTimeDuration)
			assert.Equal(t, asTimeDuration, asDur.Duration())
		})
	}
}
