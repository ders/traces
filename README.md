# traces

Package traces is a library for unevenly-spaced event series analysis.

It is based on a similar but more extensive [python package](https://github.com/datascopeanalytics/traces)
developed by [Datascope Analytics](https://datascopeanalytics.com).

## Usage

    import "github.com/ders/traces"

Create a new series.

    s := NewInt64Series()

Add points.

    s.Set(0, 10)
    s.Set(65000, 20)

Or create and add in one operation.

    s0 := NewInt64Series(map[int64]int64{0: 10, 65000: 20})

[![GoDoc](https://godoc.org/github.com/ders/traces?status.svg)](https://godoc.org/github.com/ders/traces)
