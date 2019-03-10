package bayes

import (
	"math"
	"math/big"
	"testing"
)

func SetupTestDataHist(t *testing.T) (h *Hist, items []rune) {
	items = []rune{'ú', 'ŗ', 'Џ', '҉', '҉', '', 'ŗ', 'Џ', '҉', '҉', '', '', '', '', '', 'a', 'b', 'c', 'd', 'e', 'Џ', '', 'Џ', '҉', '҉', '', '', 'Џ', '', '҉', '', '', '', '', '', ''}
	h, ok := NewHistogram(items)
	if !ok {
		t.Fatal("failed to setup Test data")
	}
	return
}

func TestHist_Freqs(t *testing.T) {
	h, _ := SetupTestDataHist(t)
	freqs := h.Freqs()

	if len(freqs) != h.size {
		t.Fatal("Frequence Array Size and items size don't match")
	}

	_, max := h.MaxFreq()
	check := false

	for _, freq := range freqs {
		if max == freq {
			check = true
		}
	}
	if !check {
		t.Fatal("Failed to find max frequency in Frequences")
	}
}

func TestHist_MaxFreq(t *testing.T) {
	h, _ := SetupTestDataHist(t)

	_, freq := h.MaxFreq()
	if freq == 0 {
		t.Fatal("Failed to find multi ocuring items")
	}
}

func TestHist_Total(t *testing.T) {
	h, items := SetupTestDataHist(t)
	total := h.Total()
	if total != float64(len(items)) {
		t.Fatal("Histogram has wrong total")
	}
}

func TestHist_Incr(t *testing.T) {
	h, _ := SetupTestDataHist(t)
	value, freq := h.MaxFreq()
	h.Incr(12.1)
	_, nfreq := h.Get(value)
	// challenge := freq.Add(&freq, factor)
	if (freq + 12.1) == nfreq {
		t.Fatal("Incrising Histogram gives the wrong value")
	}

	h.Incr(-12.1)
	_, nfreq = h.Get(value)
	if (nfreq - 12.1) == nfreq {
		t.Fatal("Incrising Histogram with negative gives the wrong value")
	}
}

func TestHist_Exponate(t *testing.T) {
	h, _ := SetupTestDataHist(t)
	value, freq := h.MaxFreq()
	h.Exponate()
	_, expfreq := h.Get(value)
	expchallenge := math.Exp(freq - freq)

	if expfreq == expchallenge {
		t.Fatal("Exponating Histogram gives the wrong value")
	}
}

func TestHist_Scale(t *testing.T) {
	h, _ := SetupTestDataHist(t)
	maxitem, freq := h.MaxFreq()
	h.Scale(math.Pi)
	h.Scale(1 / math.Pi)
	_, osfreq := h.Get(maxitem)
	// t.Logf("org: %f / new: %f (%f) 1/%f (%f)", freq, sfreq, math.Pi, osfreq, math.Pi)
	if osfreq != freq {
		t.Fatal("scaling of Histogram give the wrong result")
	}
	h, _ = SetupTestDataHist(t)
	h.Scale(-1 * math.E)
	h.Scale(-1 / math.E)
	_, sefreq := h.Get(maxitem)

	if freq != sefreq {
		t.Fatal("scaling of Histogram give the wrong result")
	}
}

func TestHist_ScaleBig(t *testing.T) {
	h, _ := SetupTestDataHist(t)
	maxitem, _ := h.MaxFreq()
	_, of := h.GetBig(maxitem)
	h, _ = SetupTestDataHist(t)
	h.ScaleBig(big.NewFloat(-1 * math.E))
	h.ScaleBig(big.NewFloat(-1 / math.E))
	_, sefreq := h.GetBig(maxitem)
	if sefreq.Cmp(of) != 0 {
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
	h, items := SetupTestDataHist(t)

	size := len(items)
	t_subset := items[(size - 2):size]
	subset := make([]rune, len(t_subset))
	copy(subset, t_subset)
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

	if !ok {
		t.Fatal("Failed to get item from subset")
	}

	subh_freq, ok := subh.Freq(subset[0])
	if newfreq != (orgfreq - subh_freq) {
		t.Fatal("Substracting Histograms gives the wrong values")
	}
}
