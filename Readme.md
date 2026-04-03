# Go ISO8601 Duration

A small, zero alloc and high performance libary for parsing
[ISO8601](https://en.wikipedia.org/wiki/ISO_8601#Durations) compliant duration
format into Go time compatible representation.

```shell
go get github.com/xnacly/go-iso8601-duration
```

## Features

- ISO8601 duration parsing formats: `[+-]P[n]Y[n]W[n]M[n]DT[n]H[n]M[n]S` via
  `goiso8601duration.From`:
    - aligned with [ECMAScript: Temporal - 7 Temporal.Duration
      Objects](https://tc39.es/proposal-temporal/#sec-temporal-duration-objects)
      and ISO8601
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

## Benchmarks

Reproduce with:

```text
go test -bench=. -benchmem
goos: linux
goarch: amd64
pkg: github.com/xnacly/go-iso8601-duration
cpu: AMD Ryzen 7 3700X 8-Core Processor
BenchmarkDuration/P0D-16                100000000               10.51 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/PT15H-16              80446293                14.58 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P1W-16                100000000               10.33 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P15W-16               93517448                12.75 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P1Y15W-16             70472916                17.19 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P15Y-16               89461000                12.75 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P15Y3M-16             61222558                18.65 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P15Y3M41D-16          46391493                25.45 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/PT15M-16              80630947                14.98 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/PT15M10S-16           55008163                22.14 ns/op            0 B/op          0 allocs/op
BenchmarkDuration/P3Y6M4DT12H30M5S-16   24965916                46.99 ns/op            0 B/op          0 allocs/op
PASS
ok      github.com/xnacly/go-iso8601-duration   12.956s
```
