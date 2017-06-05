package traces

import (
	"reflect"
	"strconv"
	"testing"
)

type testCase struct {
	Points map[int64]int64
	Floors []int64
}

// No points
var case0 = testCase{
	Points: map[int64]int64{},
	Floors: []int64{},
}

// One point
var case1 = testCase{
	Points: map[int64]int64{1: 25},
	Floors: []int64{-10, -1, 0, 1, 2, 3, 24, 25, 26},
}

// Two points
var case2 = testCase{
	Points: map[int64]int64{32: -7, -5: 20},
	Floors: []int64{-6, -5, -4, 0, 31, 32, 33},
}

// Three points
var case3 = testCase{
	Points: map[int64]int64{100: 10, 101: 0, 102: -50},
	Floors: []int64{99, 100, 101, 102, 130},
}

// Many points
var casen = testCase{
	Points: map[int64]int64{-100: 12345678, 0: 1, 1: 5, 3: 77, 5: 0, 8: 1},
	Floors: []int64{0, 1, 2, 3, 4, 5},
}

// Redundant points
var caser = testCase{
	Points: map[int64]int64{0: 0, 2: 10, 4: 10, 5: 9, 10: 8, 20: 8, 22: 8, 30: 0},
	Floors: []int64{-1, 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 20, 22, 30, 100},
}

// Redundant points compacted
var caserc = testCase{
	Points: map[int64]int64{0: 0, 2: 10, 5: 9, 10: 8, 30: 0},
	Floors: []int64{},
}

// Function assert is a simple helper method to generate an error
// when two values are not deeply equal.
func assert(t *testing.T, message string, expected, got interface{}) {
	if !reflect.DeepEqual(expected, got) {
		t.Errorf("%s - expected %v, got %v", message, expected, got)
		// uncomment the next two lines to test the tests
		//} else {
		//	t.Errorf("%s - got %v as expected", message, got)
	}
}

// The most important thing this set of tests can do is to ensure that
// Int64Series objects are always internally consistent.
//
// Function assertConsistent checks that the keys (x values) in s.points
// match s.sorted and s.unsorted exactly.  It also checks that the sorted
// points are indeed sorted.
//
// Tests should call assertConsistent all the time.
func assertConsistent(t *testing.T, s *Int64Series) {
	// Check that points and sorted+unsorted are 1-to-1.  This also ensures
	// that there are no dupes in sorted+unsorted.
	matcher := make(map[int64]struct{})
	for x := range s.points {
		matcher[x] = struct{}{}
	}
	for _, x := range s.sorted {
		if _, ok := matcher[x]; ok {
			delete(matcher, x)
		} else {
			t.Errorf("Inconsistent series: x-value %d from sorted not in points.", x)
		}
	}
	for _, x := range s.unsorted {
		if _, ok := matcher[x]; ok {
			delete(matcher, x)
		} else {
			t.Errorf("Inconsistent series: x-value %d from unsorted not in points.", x)
		}
	}
	for x := range matcher {
		t.Errorf("Inconsistent series: x-value %d in neither sorted nor unsorted", x)
	}
	// Check that sorted is in order.
	for i, x := range s.sorted {
		if i > 0 && s.sorted[i-1] >= x {
			t.Errorf("Inconsistent series: sorted is not in order (%d ≮ %d)",
				s.sorted[i-1], x)
		}
	}
}

func TestNewInt64Series(t *testing.T) {
	s := NewInt64Series()
	assertConsistent(t, s)
	assert(t, "Size of a new series", 0, s.Size())
}

// Test a few basic oprations with all the test cases
func TestBasic(t *testing.T) {
	for _, c := range []testCase{case0, case1, case2, case3, casen, caser} {
		s := NewInt64SeriesData(c.Points)
		assertConsistent(t, s)
		assert(t, "Size of a new series", len(c.Points), s.Size())
		xs := s.Xs() // this forces a sort
		assertConsistent(t, s)
		assert(t, "Size of Xs", len(c.Points), len(xs))
		for i, x := range xs {
			if i > 0 && x <= xs[i-1] {
				t.Errorf("X values out of order: expected %d < %d", xs[i-1], x)
			}
		}
		if len(xs) == 0 {
			assert(t, "x0", int64(0), s.X0())
		} else {
			assert(t, "x0", xs[0], s.X0())
		}

		var expectedFloor int64
		var floorOk bool
		var lastY int64
		for _, x := range c.Floors {
			xString := strconv.FormatInt(x, 10)
			expectedY, has := c.Points[x]
			assert(t, "Has point "+xString, has, s.Has(x))
			if has {
				expectedFloor = x
				floorOk = true
				lastY = expectedY
			} else {
				expectedY = lastY
			}
			y := s.Get(x)
			assert(t, "f("+xString+")", expectedY, y)
			x0, ok := s.Floor(x)
			assert(t, "Floor of "+xString, expectedFloor, x0)
			assert(t, "Floor ok", floorOk, ok)
			assertConsistent(t, s)
		}
	}
}

func TestSetRemoveCopy(t *testing.T) {
	for _, c := range []testCase{case0, case1, case2, case3, casen, caser} {
		s0 := NewInt64SeriesData(c.Points)
		s1 := NewInt64Series()
		for x, y := range c.Points {
			s1.Set(x, y)
		}
		assertConsistent(t, s1)
		if !s1.Equals(s0) {
			t.Errorf("Set points - expected %v, got %v", s0, s1)
		}
		s2 := s0.Copy()
		assertConsistent(t, s0)
		assertConsistent(t, s2)
		if !s2.Equals(s0) {
			t.Errorf("Copy - expected %v, got %v", s0, s2)
		}
		size := s2.Size()
		for x := range c.Points {
			s2.Remove(x)
			size--
			assert(t, "Size", size, s2.Size())
			assertConsistent(t, s2)
		}
		assert(t, "Size", 0, size)

		// this is to ensure our copy is actually distinct
		if !s0.Equals(s1) {
			t.Errorf("Copy - expected %v, got %v", s1, s0)
		}
	}
}

func TestCeiling(t *testing.T) {
	for _, c := range []testCase{case0, case1, case2, case3, casen, caser} {
		s := NewInt64SeriesData(c.Points)
		var expectedCeiling int64
		var ceilingOk bool
		// iterate Floors backwards for ceilings
		for i := len(c.Floors) - 1; i >= 0; i-- {
			x := c.Floors[i]
			if s.Has(x) {
				expectedCeiling = x
				ceilingOk = true
			}
			x0, ok := s.Ceiling(x)
			assert(t, "Ceiling of "+strconv.FormatInt(x, 10), expectedCeiling, x0)
			assert(t, "Ceiling ok", ceilingOk, ok)
			assertConsistent(t, s)
		}
	}
}

func TestCompact(t *testing.T) {
	s0 := NewInt64SeriesData(caser.Points)
	s1 := NewInt64SeriesData(caserc.Points)
	if s0.Equals(s1) {
		t.Errorf("Ccompact - expected %v to differ from %v", s1, s0)
	}
	s0.Compact()
	assertConsistent(t, s0)
	if !s0.Equals(s1) {
		t.Errorf("Ccompact - expected %v, got %v", s1, s0)
	}

	// a little extra beating
	for _, c := range []testCase{case0, case1, case2, case3, casen} {
		sRef := NewInt64SeriesData(c.Points)
		sRef.Compact()
		assertConsistent(t, sRef)

		sBloat := NewInt64SeriesData(c.Points)
		// this adds a bunch of redundant points
		x0 := sRef.X0()
		for _, x := range c.Floors {
			if x > x0 {
				sBloat.Set(x, sRef.Get(x))
			}
		}
		assertConsistent(t, sBloat)
		sBloat.Compact()
		assertConsistent(t, sBloat)
		if !sBloat.Equals(sRef) {
			t.Errorf("Compact (extra) - expected %v, got %v", sRef, sBloat)
		}
	}
}
