package bayes

import "testing"

func SetupTestDataSh(t *testing.T) (ch *SortedHist) {
	items := []rune{'ú', 'ŗ', 'Џ', '҉', '҉', '', 'ŗ', 'Џ', '҉', '҉', '', '', '', '', '', 'a', 'b', 'c', 'd', 'e', 'Џ', '', 'Џ', '҉', '҉', '', '', 'Џ', '', '҉', '', '', '', '', '', ''}
	ch, ok := NewSortedHist(items)
	if !ok {
		t.Fatal("Failed to create Sorted Test List")
	}
	return
}

func TestNewOrderedHistFromFloat(t *testing.T) {
	floatlist := []float32{15.99, 1, 0.22, 3, 0.912, 12.56, 1}
	ordered, ok := NewOrderedHistFromFloat(floatlist)
	if !ok {
		t.Fatal("failed to create Ordered List")
	}
	keys := ordered.Keys()
	if keys[0] >= keys[1] {
		t.Fatal("Sort failed")
	}
}

func TestCreateSortedHist(t *testing.T) {
	sl := SetupTestDataSh(t)
	fs := sl.Freqs()
	if fs[sl.size-2] >= fs[sl.size-1] {
		t.Fatal("Failed to sort List")
	}
}

func TestSortedHist_Scale(t *testing.T) {
	sl := SetupTestDataSh(t)
	fs := sl.Freqs()
	sl.Scale(0.1)
	ss := sl.Freqs()
	if fs[0] == ss[0] {
		t.Fatal("Frequences match after scaling")
	}
}
