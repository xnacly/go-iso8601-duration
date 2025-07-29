# Go ISO8601 Duration

A small libary for parsing
[ISO8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) compliant duration
format into Go time compatible representation.

```shell
go get github.com/xnacly/go-iso8601-duration
```

## Features

- ISO8601 duration parsing formats: `P[nn]Y[nn]M[nn]DT[nn]H[nn]M[nn]S` or
  `P[nn]W` via `goiso8601duration.From`
- Interop with [`time.Time`](https://pkg.go.dev/time#Time) and
  [`time.Duration`](https://pkg.go.dev/time#Duration) via:
  - `goiso8601duration.Duration.Apply(time.Time) time.Time`
  - `goiso8601duration.Duration.Duration() time.Duration`
  - `goiso8601duration.FromDuration(time.Duration) goiso8601duration.Duration`
- Serializing `goiso8601duration.Duration` to the correct format
  `goiso8601duration.Duration.String`
- JSON serialisation support via
  [`json.Unmarshaller`](https://pkg.go.dev/encoding/json#Unmarshaler) and
  [`json.Marshaller`](https://pkg.go.dev/encoding/json#Marshaler)
- High quality errors with position context.

## Non features

> I am open to implement these, if someone needs them

- Negative durations `(-P1Y)` -> because I don't have access to the newest spec, so I have no idea what it specifies
- Fractional seconds `(PT1.5S)` -> technically this can be supported for the last / smallest unit, but I dont think its necessary
- Arbitrary digit lengths -> spec says two digits is fine, so I'll keep it like that

## Example

```go
package main

import (
    "fmt"
    "time"

    "github.com/xnacly/go-iso8601-duration"
)

func Must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}

func main() {
	rawDuration := "PT1H30M12S"
	duration := Must(goiso8601duration.From(rawDuration))

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
		In goiso8601duration.Duration `json:"in"`
	}

	asJson := Must(
		json.Marshal(
			arrival{
				In: goiso8601duration.FromDuration(
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
```
