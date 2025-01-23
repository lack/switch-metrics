package restconf

import (
	"fmt"
	"math"
	"strings"
)

type Histogram struct {
	Buckets []int
	Count   []int
}

func (h *Histogram) init() {
	if h.Buckets[len(h.Buckets)-1] != math.MaxInt {
		h.Buckets = append(h.Buckets, math.MaxInt)
	}
	h.Count = make([]int, len(h.Buckets))
}

func (h *Histogram) Add(value int) {
	if len(h.Count) == 0 {
		h.init()
	}
	for i, boundary := range h.Buckets {
		if value < boundary {
			h.Count[i] = h.Count[i] + 1
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

// Render returns a two-string tuple; the header row, and the data row
func (h *Histogram) Render() (string, string) {
	header := strings.Builder{}
	counts := strings.Builder{}
	for i, v := range h.Count {
		name := h.BucketName(i)
		header.WriteString(name)
		header.WriteRune(' ')
		counts.WriteString(fmt.Sprintf("%*d ", len(name), v))
	}
	return header.String(), counts.String()
}
