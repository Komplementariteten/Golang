package bayes

import (
	"log"
	"math"
	"math/big"
	"reflect"

	"ventose.cc/tools"
)

type Hist struct {
	precision uint
	Items     map[interface{}]*big.Float
	size      int
}

func NewHistogram(values interface{}) (h *Hist, ok bool) {
	var rv reflect.Value
	if rv, ok = tools.IsSlice(values); !ok {
		return
	}
	h = &Hist{}
	h.Items = make(map[interface{}]*big.Float)
	h.precision = 10
	h.size = 0
	for i := 0; i < rv.Len(); i++ {
		item := rv.Index(i).Interface()
		h.set(item)
	}
	return
}

func (h *Hist) set(item interface{}) {
	if _, ok := h.Items[item]; ok {
		h.Items[item].Add(h.Items[item], big.NewFloat(1))
	} else {
		h.Items[item] = big.NewFloat(1).SetPrec(h.precision)
		h.size++
	}
}

func (h *Hist) Freq(value interface{}) (freq float64, ok bool) {
	if _, ok = h.Items[value]; ok {
		freq, _ = h.Items[value].SetPrec(h.precision).Float64()
	}
	return
}

func (h *Hist) Freqs() []float64 {
	result := make([]float64, h.size)
	i := 0
	for k := range h.Items {
		result[i], _ = h.Freq(k)
		i++
	}
	return result
}

func (h *Hist) Get(item interface{}) (i interface{}, res float64) {
	if v, ok := h.Items[item]; ok {
		i = item
		res, _ = v.Float64()
	}
	return
}

func (h *Hist) GetBig(item interface{}) (interface{}, *big.Float) {
	if v, ok := h.Items[item]; ok {
		return item, v
	}
	return nil, nil
}

func (h *Hist) Scale(factor float64) {
	bigFactor := big.NewFloat(factor)
	for k, v := range h.Items {
		s := v.Mul(v, bigFactor)
		pre := s.Prec()
		log.Println(pre)
		h.Items[k] = s
	}
}

func (h *Hist) ScaleBig(factor *big.Float) {
	for k, v := range h.Items {
		s := v.Mul(v, factor)
		h.Items[k] = s
	}
}

func (h *Hist) Incr(factor float64) {
	bigFactor := big.NewFloat(factor)
	for k, v := range h.Items {
		h.Items[k] = v.SetPrec(h.precision).Add(v, bigFactor)
	}
}

func (h *Hist) Remove(item interface{}) {
	delete(h.Items, item)
}

func (h *Hist) Exponate() {
	_, maxBig := h.MaxFreq()
	for k, v := range h.Items {
		fl, _ := v.Float64()
		flexp := math.Exp(fl)
		h.Items[k] = big.NewFloat(flexp - maxBig).SetPrec(h.precision)
	}
}

func (h *Hist) MaxFreq() (item interface{}, res float64) {
	var max *big.Float
	for i, f := range h.Items {
		if max == nil {
			max = f
		}
		if f.Cmp(max) >= 1 {
			max.Set(f)
			item = i
			res, _ = max.Float64()
		}
	}
	return
}

func (h *Hist) Substract(other *Hist) {
	wh := other
	for k, v := range wh.Items {
		fvalue, _ := v.Float64()
		fvalue = fvalue * -1
		sub := big.NewFloat(fvalue)
		if i, ok := h.Items[k]; ok {
			h.Items[k] = i.Add(h.Items[k], sub)
		}
	}
}

func (h *Hist) Total() float64 {
	t := big.NewFloat(0).SetPrec(h.precision)
	for _, v := range h.Items {
		t.Add(t, v)
	}
	ret, _ := t.SetPrec(h.precision).Float64()
	return ret
}

// Divides all items of the Histogram in this manner factor / item
func (h *Hist) Fraction(factor float64) {
	bigFactor := big.NewFloat(factor)
	for k, v := range h.Items {
		h.Items[k] = v.Quo(bigFactor, v).SetPrec(h.precision)
	}
}
