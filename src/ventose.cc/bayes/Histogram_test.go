package bayes

import (
	"math"
	"math/big"
	"testing"
)

func SetupTestData(t *testing.T) (h *Hist, items []rune) {
	items = []rune{'ú', 'ŗ', 'Џ', '҉', '҉', '', 'ŗ', 'Џ', '҉', '҉', '', '', '', '', '', 'a', 'b', 'c', 'd', 'e', 'Џ', '', 'Џ', '҉', '҉', '', '', 'Џ', '', '҉', '', '', '', '', '', ''}
	h, ok := NewHistogram(items)
	if !ok {
		t.Fatal("failed to setup Test data")
	}
	return
}

func TestFloatingAddition(t *testing.T) {
	f1 := float64(5)
	f2 := float64(12.1)
	f5 := float64(-12.1)
	f3 := f1 + f2
	f4 := f3 + f5
	if f4 != f1 {
		t.Fatal("flaotings don't add up")
	}
}

func TestHist_MaxFreq(t *testing.T) {
	h, _ := SetupTestData(t)

	_, freq := h.MaxFreq()
	if freq == 0 {
		t.Fatal("Failed to find multi ocuring items")
	}
}

func TestHist_Total(t *testing.T) {
	h, items := SetupTestData(t)
	total := h.Total()
	if total != float64(len(items)) {
		t.Fatal("Histogram has wrong total")
	}
}

func TestHist_Incr(t *testing.T) {
	h, _ := SetupTestData(t)
	value, freq := h.MaxFreq()
	factor := float64(12.1)
	h.Incr(factor)
	_, nfreq := h.Get(value)
	challenge := freq + factor
	if nfreq != challenge {
		t.Fatal("Incrising Histogram gives the wrong value")
	}

	factor = float64(-12.1)
	h.Incr(factor)
	_, nfreq = h.Get(value)
	bigFloat := big.NewFloat(freq)
	addValue := big.NewFloat(factor)
	bigFloat.Add(bigFloat, addValue)
	challenge, _ = bigFloat.Float64()
	if nfreq != challenge {
		t.Fatal("Incrising Histogram with negative gives the wrong value")
	}
}

func TestHist_Exponate(t *testing.T) {
	h, _ := SetupTestData(t)
	value, freq := h.MaxFreq()
	h.Exponate()
	_, expfreq := h.Get(value)
	expchallenge := math.Exp(freq - freq)
	if expfreq != expchallenge {
		t.Fatal("Exponating Histogram gives the wrong value")
	}
}

func TestHist_Scale(t *testing.T) {
	h, _ := SetupTestData(t)
	_, freq := h.MaxFreq()
	h.Scale(math.Phi)
	maxitem, sfreq := h.MaxFreq()
	if freq != (sfreq / math.Phi) {
		t.Fatal("scaling of Histogram give the wrong result")
	}
	h, _ = SetupTestData(t)
	h.Scale(math.E * -1)
	_, sfreq = h.Get(maxitem)
	if freq != (sfreq / (-1 * math.E)) {
		t.Fatal("scaling of Histogram give the wrong result")
	}
}

func TestNewHistogram(t *testing.T) {
	floats := [6]float64{0, 0.0, 1, 2, 0.1111, 444.43543637}
	hf, ok := NewHistogram(floats)
	if !ok {
		t.Fatal("Failed to create Float histogram")
	}
	c, ok := hf.Freq(float64(0))

	if !ok {
		t.Fatal("failed to get Frequensy for given float value")
	}

	if c != 2 {
		t.Fatal("got wrong Frequency for given float value")
	}

	runes := [3]rune{'ä', 'û', 'û'}
	hf, ok = NewHistogram(runes)
	if !ok {
		t.Fatal("Failed to create rune histogram")
	}

	c, ok = hf.Freq('û')

	if !ok {
		t.Fatal("failed to get Frequensy for given rune value")
	}

	if c != 2 {
		t.Fatal("got wrong Frequency for given rune value")
	}
}

func TestHist_Substract(t *testing.T) {
	h, items := SetupTestData(t)

	size := len(items)
	subset := items[(size - 2):size]
	if len(subset) != 2 {
		t.Fatal("failed to create proper Subset")
	}
	_, orgfreq := h.Get(subset[0])
	subh, ok := NewHistogram(subset)
	h.Substract(subh)
	_, newfreq := h.Get(subset[0])
	if orgfreq <= newfreq {
		t.Fatal("failed to substract Histogram")
	}

	subh_freq, ok := subh.Freq(subset[0])
	if !ok {
		t.Fatal("Failed to get item from subset")
	}

	if newfreq != (orgfreq - subh_freq) {
		t.Fatal("Substracting Histograms gives the wrong values")
	}
}
