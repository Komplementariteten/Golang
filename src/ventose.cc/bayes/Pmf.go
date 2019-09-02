package bayes

import (
	"errors"
)

type Pmf struct {
	SortedHist
}

func NewPmf(values interface{}) (p *Pmf, ok bool) {
	h, ok := NewSortedHist(values)
	if !ok {
		return
	}
	p = &Pmf{}
	p.h = h.h
	p.k = h.k
	p.f = h.f
	p.size = h.size
	return
}

func (p *Pmf) Prob(x interface{}, def float64) float64 {
	prop, ok := p.h.Freq(x)
	if !ok {
		return def
	}
	return prop
}

func (p *Pmf) Normalize(fraction float64) error {
	total := p.h.Total()
	if total == 0.0 {
		return errors.New("total probability is zero")
	}
	// bigRatTotal := new(big.Rat).SetFrac(big.NewInt(1), big.NewInt(int64(total)))
	p.Scale(fraction)
	p.Frac(1)
	return nil
}

func (p *Pmf) Fraction(a, b int64) {
	/* rat := new(big.Rat).SetFrac(big.NewInt(a), big.NewInt(b))
	for k, v := p.h.Items {

	}*/
}
