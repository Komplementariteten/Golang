package bayes

import (
	"math/big"
	"reflect"
	"sort"
)

type sortedFloat []float32

func (s sortedFloat) Len() int           { return len(s) }
func (s sortedFloat) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortedFloat) Less(i, j int) bool { return s[i] < s[j] }

type OrderedHist struct {
	Hist
	keys []float32
}

func NewOrderedHistFromFloat(values []float32) (o *OrderedHist, ok bool) {
	sort.Sort(sortedFloat(values))
	o = &OrderedHist{}
	o.keys = values
	o.Items = make(map[interface{}]*big.Float)
	o.size = len(values)
	for _, item := range values {
		o.set(item)
	}
	ok = true
	return
}

func (oh *OrderedHist) Get(v float32) (key float32, value interface{}) {
	if _, ok := oh.Items[v]; ok {
		key = v
		value = oh.Items[v]
	}
	return
}

func (oh *OrderedHist) Keys() []float32 {
	return oh.keys
}

type SortedHist struct {
	h    *Hist
	f    []float64
	k    []reflect.Value
	size int
}

func (s *SortedHist) Len() int { return s.size }

func (s *SortedHist) Swap(i, j int) {
	s.f[i], s.f[j] = s.f[j], s.f[i]
	s.k[i], s.k[j] = s.k[j], s.k[i]
}
func (s *SortedHist) Less(i, j int) bool {
	a := s.f[i]
	b := s.f[j]
	return a < b
}

func (s *SortedHist) Freqs() []float64 {
	return s.f
}

func (s *SortedHist) Keys() (r []interface{}) {
	r = make([]interface{}, len(s.k))
	for ke, ve := range s.k {
		r[ke] = ve
	}
	return
}

func (s *SortedHist) Scale(factor float64) {
	s.h.Scale(factor)
	scaledList := CreateSortedHist(s.h)
	*s = *scaledList
}

func (s *SortedHist) Frac(factor *big.Rat) {
	s.h.Fraction(factor)
	scaledList := CreateSortedHist(s.h)
	*s = *scaledList
}

func CreateSortedHist(h *Hist) (sh *SortedHist) {
	sh = &SortedHist{
		h:    h,
		f:    h.Freqs(),
		k:    reflect.ValueOf(h.Items).MapKeys(),
		size: h.size,
	}
	sort.Sort(sh)
	return
}

func NewSortedHist(values interface{}) (sh *SortedHist, ok bool) {
	h, ok := NewHistogram(values)
	sh = CreateSortedHist(h)
	return
}
