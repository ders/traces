package traces

import (
	"testing"
)

type combineTestCase struct {
	Vals []map[int64]int64
	Sum  map[int64]int64
	Diff map[int64]int64
	Any  map[int64]int64
	All  map[int64]int64
}

var combineTestCases = []combineTestCase{

	combineTestCase{
		Vals: []map[int64]int64{},
		Sum:  map[int64]int64{},
		Diff: map[int64]int64{},
		Any:  map[int64]int64{},
		All:  map[int64]int64{},
	},

	combineTestCase{
		Vals: []map[int64]int64{
			map[int64]int64{-5: 3, 0: 0, 123: 1},
			map[int64]int64{-10: 1, 0: 2, 50: 0},
		},
		Sum:  map[int64]int64{-10: 1, -5: 4, 0: 2, 50: 0, 123: 1},
		Diff: map[int64]int64{-10: -1, -5: 2, 0: -2, 50: 0, 123: 1},
		Any:  map[int64]int64{-10: 1, 50: 0, 123: 1},
		All:  map[int64]int64{-10: 0, -5: 1, 0: 0},
	},
}

func TestCombine(t *testing.T) {
	for _, c := range combineTestCases {
		list := make([]*Series, len(c.Vals))
		for i, val := range c.Vals {
			list[i] = NewSeriesData(val)
		}

		sum := Combine(Sum, list...)
		expectedSum := NewSeriesData(c.Sum)
		expectedSum.Compact()
		assertConsistent(t, sum)
		sum.Compact()
		assertConsistent(t, sum)
		assert(t, "Combine as sum", expectedSum, sum)

		diff := Combine(Diff, list...)
		expectedDiff := NewSeriesData(c.Diff)
		expectedDiff.Compact()
		assertConsistent(t, diff)
		diff.Compact()
		assertConsistent(t, diff)
		assert(t, "Combine as diff", expectedDiff, diff)

		any := Combine(Any, list...)
		expectedAny := NewSeriesData(c.Any)
		expectedAny.Compact()
		assertConsistent(t, any)
		any.Compact()
		assertConsistent(t, any)
		assert(t, "Combine as any", expectedAny, any)

		all := Combine(All, list...)
		expectedAll := NewSeriesData(c.All)
		expectedAll.Compact()
		assertConsistent(t, all)
		all.Compact()
		assertConsistent(t, all)
		assert(t, "Combine as or", expectedAll, all)
	}
}
