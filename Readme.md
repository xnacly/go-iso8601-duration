# Go ISO8601 Duration

A small libary for parsing
[ISO8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) compliant duration
format into Go time compatible representation.

```shell
go get github.com/xnacly/go-iso8601-duration
```

## Features

- ISO8601 duration parsing formats: `P[nn]Y[nn]M[nn]DT[nn]H[nn]M[nn]S` or `P[nn]W` via `goiso8601duration.From`
- Interop with [`time.Time`](https://pkg.go.dev/time#Time) and [`time.Duration`](https://pkg.go.dev/time#Duration) via
  `ISO8601Duration.Apply(time.Time) time.Time` and `ISO8601Duration.Duration() time.Duration`
- Serializing `ISO8601Duration` to the correct format via `ISO8601Duration.String`
- High quality errors with position context.

## Non features

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

func main() {
	rawDuration := "PT1H30M12S"
	duration, err := goiso8601duration.From(rawDuration)
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
```
