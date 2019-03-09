package bayes

import (
	"math"
	"reflect"

	"ventose.cc/tools"
)

type Hist struct {
	Items map[interface{}]float64
}

func NewHistogram(values interface{}) (h *Hist, ok bool) {
	var rv reflect.Value
	if rv, ok = tools.IsSlice(values); !ok {
		return
	}
	h = &Hist{}
	h.Items = make(map[interface{}]float64)
	len := rv.Len()
	for i := 0; i < len; i++ {
		item := rv.Index(i).Interface()
		h.set(item)
	}
	return
}

func (h *Hist) set(item interface{}) {
	if _, ok := h.Items[item]; ok {
		h.Items[item]++
	} else {
		h.Items[item] = 1
	}
}

func (h *Hist) Freq(value interface{}) (freq float64, ok bool) {
	if _, ok = h.Items[value]; ok {
		freq = h.Items[value]
	}
	return
}

func (h *Hist) Get(item interface{}) (i interface{}, freq float64) {
	i = nil
	freq = float64(0)
	if v, ok := h.Items[item]; ok {
		i = item
		freq = v
	}
	return
}

func (h *Hist) Scale(factor float64) {
	for k, v := range h.Items {
		h.Items[k] = v * factor
	}
}

func (h *Hist) Incr(factor float64) {
	for k, v := range h.Items {
		//bigFloat := big.NewFloat(v)
		//addValue := big.NewFloat(factor)
		//bigFloat.Add(bigFloat, addValue)
		//result, _ := bigFloat.Float64()
		h.Items[k] = v + factor
	}
}

func (h *Hist) Remove(item interface{}) {
	delete(h.Items, item)
}

func (h *Hist) Exponate() {
	_, max := h.MaxFreq()
	for k, v := range h.Items {
		h.Items[k] = math.Exp(v - max)
	}
}

func (h *Hist) MaxFreq() (item interface{}, count float64) {
	count = 0
	for i, f := range h.Items {
		if count == 0 {
			count = f
		}
		if f > count {
			count = f
			item = i
		}
	}
	return
}

func (h *Hist) Substract(other *Hist) {
	for k, v := range other.Items {
		if _, ok := h.Items[k]; ok {
			h.Items[k] -= v
		}
	}
}

func (h *Hist) Total() float64 {
	t := float64(0)
	for _, v := range h.Items {
		t += v
	}
	return t
}
