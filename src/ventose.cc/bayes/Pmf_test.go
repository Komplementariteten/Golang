package bayes

import "testing"

func SetupTestDataPmf(t *testing.T) (ch *Pmf) {
	items := []rune{'ú', 'ŗ', 'Џ', '҉', '҉', '', 'ŗ', 'Џ', '҉', '҉', '', '', '', '', '', 'a', 'b', 'c', 'd', 'e', 'Џ', '', 'Џ', '҉', '҉', '', '', 'Џ', '', '҉', '', '', '', '', '', ''}
	ch, ok := NewPmf(items)
	if !ok {
		t.Fatal("Failed to create Sorted Test List")
	}
	return
}

func TestPmf_Normalize(t *testing.T) {
	pmf := SetupTestDataPmf(t)
	pmf.Normalize(1.0)
	total := pmf.h.Total()
	if total != 1.0 {
		t.Fatal("Totals don't add up to Normalization factor")
	}
}
