package bayes

type Interpolator struct {
	xs []float64
	ys []float64
}

func NewInterpolator(x []float64, y []float64) *Interpolator {
	return &Interpolator{xs: x, ys: y}
}

func (i *Interpolator) Lookup(x float64) {

}
