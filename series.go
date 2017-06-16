package traces

import (
	"reflect"
	"sort"
)

// Series represents a discrete function f(x)=y with a collection of
// (x, y) pairs.  Each (x, y) pair represents a transition, i.e. if (x₀, y₀)
// and (x₁, y₁) are consecute pairs, then f(x)=y₀ for x₀ ≤ x < x₁.
// All x and y values are int64.
type Series struct {
	points   map[int64]int64
	sorted   []int64
	unsorted []int64
	// Internal consistency is always maintained such that all of the keys
	// in the points map appear exactly once in either sorted or unsorted,
	// and all of the keys in sorted are in order.
}

// NewSeries returns a new, empty Series object.
func NewSeries() *Series {
	return &Series{
		points:   make(map[int64]int64),
		sorted:   make([]int64, 0),
		unsorted: make([]int64, 0),
	}
}

// NewSeriesData returns a new Series object prefilled with the data in the map.
func NewSeriesData(data map[int64]int64) *Series {
	points := make(map[int64]int64)
	unsorted := make([]int64, 0, len(points))
	for key, val := range data {
		points[key] = val
		unsorted = append(unsorted, key)
	}

	return &Series{
		points:   points,
		sorted:   make([]int64, 0),
		unsorted: unsorted,
	}
}

// sort takes any unsorted keys in s.unsorted and merges them into s.sorted.
func (s *Series) sort() {
	if len(s.unsorted) == 0 {
		return
	}
	s.sorted = append(s.sorted, s.unsorted...)
	s.unsorted = make([]int64, 0)
	sort.Slice(s.sorted, func(i, j int) bool { return s.sorted[i] < s.sorted[j] })
}

// find finds and returns the largest index i into sorted such that
// s.sorted[i] <= x.  Returns -1 if x < s.sorted[0] or if s.sorted is
// empty.
func (s *Series) find(x int64) int {

	if len(s.sorted) == 0 || x < s.sorted[0] {
		return -1
	}

	// i, j are bounds such that sorted[i] <= key < sorted[j].
	// We will narrow the bounds until j-i is 1 or until we find
	// the exact key.
	i, j := 0, len(s.sorted)

	for j-i > 1 {
		half := (i + j + 1) / 2
		if x == s.sorted[half] {
			return half
		} else if x > s.sorted[half] {
			i = half
		} else {
			j = half
		}
	}

	return i
}

// Size returns the number of stored points in the series.
func (s *Series) Size() int {
	return len(s.points)
}

// Has returns true if there is a stored point at x.
func (s *Series) Has(x int64) bool {
	_, ok := s.points[x]
	return ok
}

// Set adds the point (x, y) to the series, replacing the existing point
// at x if there is one.
func (s *Series) Set(x, y int64) {
	if _, ok := s.points[x]; !ok {
		s.unsorted = append(s.unsorted, x)
	}
	s.points[x] = y
}

// Get retrieves the value f(x).  If x in not a stored point in the series,
// then f(x) is defined as f(x₀) for the largest x₀ < x.  If there is no such
// x₀, then f(x)=0.
func (s *Series) Get(x int64) int64 {
	if y, ok := s.points[x]; ok {
		return y
	}
	s.sort()
	i := s.find(x)
	if i < 0 {
		return 0
	}
	return s.points[s.sorted[i]]
}

// Remove removes the stored point at x from the series if it exists.
func (s *Series) Remove(x int64) {
	if _, ok := s.points[x]; !ok {
		return
	}
	s.sort()
	i := s.find(x)
	s.sorted = append(s.sorted[:i], s.sorted[i+1:]...)
	delete(s.points, x)
}

// Compact optimizes the series by removing any redundant stored points.
// A redundant point is the second in a pair of consecutive points
// (x₀, y₀) and (x₁, y₁) such that y₀ = y₁.  Removing redundant points
// does not affect the value of the function.
//
// Compact never removes the first point, even if the y value is 0.
func (s *Series) Compact() {
	if len(s.points) < 2 {
		return
	}
	s.sort()
	newSorted := []int64{s.sorted[0]}
	lastY := s.points[s.sorted[0]]
	for i, x := range s.sorted {
		if i > 0 {
			if s.points[x] == lastY {
				delete(s.points, x)
			} else {
				newSorted = append(newSorted, x)
				lastY = s.points[x]
			}
		}
	}
	s.sorted = newSorted
}

// Xs returns an ordered slice of all the x values of stored points.
// Use this method along with Get() to iterate through (x, f(x)) in order.
func (s *Series) Xs() []int64 {
	s.sort()
	xs := make([]int64, len(s.sorted))
	copy(xs, s.sorted)
	return xs
}

// X0 returns the x value of the lowest stored point.  This is equivalent
// to Xs[0].  Returns 0 if there are no stored points.
func (s *Series) X0() int64 {
	s.sort()
	if len(s.sorted) == 0 {
		return 0
	}
	return s.sorted[0]
}

// Floor returns the largest x₀ from the stored points such that x₀ ≤ x.
// If there is no such x₀, then 0 is returned along with the ok = false.
func (s *Series) Floor(x int64) (x0 int64, ok bool) {
	if len(s.points) == 0 {
		return
	}

	s.sort()
	i := s.find(x)
	if i < 0 {
		return
	}

	return s.sorted[i], true
}

// Ceiling returns the smallest x₁ from the stored points such that x₁ ≥ x.
// If there is no such x₁ then 0 is returned along with ok = false.
func (s *Series) Ceiling(x int64) (x1 int64, ok bool) {
	if len(s.points) == 0 {
		return
	}

	// We explicitly check for x being in the stored points so that we don't
	// have to account for the case of x₁ = x below.
	if _, okay := s.points[x]; okay {
		return x, true
	}

	s.sort()
	i := s.find(x) + 1 // This finds the index of the smallest x₁ > x.
	if i >= len(s.sorted) {
		return
	}

	return s.sorted[i], true
}

// Copy returns a new Series which is a copy of s.
func (s *Series) Copy() *Series {
	// Sort first to avoid having to sort twice later (once on s and once
	// on the copy).  As a side effect, we now don't have to copy s.unsorted.
	s.sort()

	points := make(map[int64]int64)
	for x, y := range s.points {
		points[x] = y
	}

	sorted := make([]int64, len(s.sorted))
	copy(sorted, s.sorted)

	return &Series{
		points:   points,
		sorted:   sorted,
		unsorted: make([]int64, 0),
	}
}

// Equals returns true if s and s0 have the same set of stored points.
// Equals does *not* ignore redundant points, and it generally advisable
// to compact both series before checking equality.
func (s *Series) Equals(s0 *Series) bool {
	return reflect.DeepEqual(s.points, s0.points)
}
