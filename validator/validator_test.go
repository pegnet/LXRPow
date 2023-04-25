package validator

import (
	"fmt"
	"math/rand"
	"sort"
	"testing"
)

func Test_Difficulty(t *testing.T) {
	var s float64
	currentDiff := float64(20)

	var window []float64 // Used to compute the median

	w := 100 * 60 / 10 // 4 hours, 10 minute blocks
	target := float64(600)
	min := target * 100
	max := float64(0)

	for i := 1; i <= 45000; i++ {
		r := rand.ExpFloat64() * currentDiff
		s += r
		window = append(window, r)
		if r < min {
			min = r
		}
		if r > max {
			max = r
		}

		if i%w == 0 {
			a := s / float64(w)
			sort.Float64s(window)
			median := window[w/2]
			percentDelta := (target - a) / target
			nx := currentDiff * (1 + percentDelta)
			f := float64(10)
			if nx > f*currentDiff {
				nx = f * currentDiff
			}
			if nx < currentDiff/f {
				nx = currentDiff / f
			}

			fmt.Printf("Index avg = %8.3f, median = %8.3f min = %8.3f max = %8.3f "+
			"x = %8.3f window = %8d target = %8.1f \n", a, median, min, max, nx, w, target)

			currentDiff = nx
			s = 0
			min = target * 100
			max = 0
			window = window[:0]
		}

	}
}
