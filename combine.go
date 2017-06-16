package traces

// Combine returns a new Series that is the result of appling function f to
// all the points in all the series in the list.
//
// Helper functions are provided for adding, subtracting, any-ing and all-ing.
// For example, to add the series s0, s1 and s2:
//
//   sum := traces.Combine(traces.Sum, s0, s1, s2)
//
func Combine(f func(...int64) int64, list ...*Series) *Series {

	points := make(map[int64]int64)
	unsorted := make([]int64, 0)

	valueSet := make([]int64, len(list))

	for _, s := range list {
		for key := range s.points {
			if _, ok := points[key]; !ok {
				for i, t0 := range list {
					valueSet[i] = t0.Get(key)
				}
				points[key] = f(valueSet...)
				unsorted = append(unsorted, key)
			}
		}
	}

	return &Series{
		points:   points,
		sorted:   make([]int64, 0),
		unsorted: unsorted,
	}
}

// Sum adds all the vals.  Use with the Combine function.
func Sum(vals ...int64) int64 {
	var sum int64
	for _, val := range vals {
		sum += val
	}
	return sum
}

// Diff subtracts from the first value all the remaining values.  Use with
// the Combine function.
func Diff(vals ...int64) int64 {
	var diff int64
	for i, val := range vals {
		if i == 0 {
			diff = val
		} else {
			diff -= val
		}
	}
	return diff
}

// Any returns one if any of the values are nonzero and zero otherwise.
// Use with the Combine function.
func Any(vals ...int64) int64 {
	var any int64
	for _, val := range vals {
		if val != 0 {
			any = 1
		}
	}
	return any
}

// All returns one if all of the values are nonzero and zero otherwise.
// Use with the Combine function.
func All(vals ...int64) int64 {
	var all int64 = 1
	for _, val := range vals {
		if val == 0 {
			all = 0
		}
	}
	return all
}
