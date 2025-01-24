package restconf

import (
	"fmt"
	"math"
	"strings"
)

type Histogram struct {
	Buckets []int
	Count   []int
	Mean    []float64
}

func (h *Histogram) init() {
	if h.Buckets[len(h.Buckets)-1] != math.MaxInt {
		h.Buckets = append(h.Buckets, math.MaxInt)
	}
	h.Count = make([]int, len(h.Buckets))
	h.Mean = make([]float64, len(h.Buckets))
}

func (h *Histogram) Add(value int) {
	if len(h.Count) == 0 {
		h.init()
	}
	for i, boundary := range h.Buckets {
		if value < boundary {
			// Increment the count
			h.Count[i] = h.Count[i] + 1
			// Accumulate the average based on the value
			if h.Count[i] > 1 {
				count := float64(h.Count[i])
				h.Mean[i] = (h.Mean[i] * ((count - 1) / count)) + (float64(value) / count)
			} else {
				h.Mean[i] = float64(value)
			}
			return
		}
	}
}

func (h *Histogram) BucketName(i int) string {
	if i == 0 {
		return fmt.Sprintf("<%d", h.Buckets[i])
	}
	if i == len(h.Buckets)-1 {
		return fmt.Sprintf(">=%d", h.Buckets[i-1])
	}
	return fmt.Sprintf("[%d %d)", h.Buckets[i-1], h.Buckets[i])
}

// Render returns a three-string tuple; the header row, counts, and the averages
func (h *Histogram) Render() (string, string, string) {
	header := strings.Builder{}
	counts := strings.Builder{}
	averages := strings.Builder{}
	for i, v := range h.Count {
		name := h.BucketName(i)
		header.WriteString(name)
		header.WriteRune(' ')
		counts.WriteString(fmt.Sprintf("%*d ", len(name), v))
		averages.WriteString(fmt.Sprintf("%*.1f ", len(name), h.Mean[i]))
	}
	return header.String(), counts.String(), averages.String()
}
