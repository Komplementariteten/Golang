package img

import "testing"

func Test_64bit_opperations(t *testing.T) {
	base := uint64(0x0001000200030004)
	t.Log(base)
	firstBytes := uint64(0xFFFF)
	secondBytes := uint64(0xFFFF0000)
	thirdBytes := uint64(0xFFFF00000000)
	fourthBytes := uint64(0xFFFF000000000000)

	t.Log("Bitwise Link")
	t.Log(base & firstBytes)
	t.Log(base & secondBytes)
	t.Log(base & thirdBytes)
	t.Log(base & fourthBytes)

	t.Log("Link & Schift")

	t.Log((base & firstBytes)  >> 00)
	t.Log((base & secondBytes) >> 020)
	t.Log((base & thirdBytes) >> 040)
	t.Log((base & fourthBytes) >> 060)

	t.Log("------------------")

	smallBase := 0xFFFFFF
	t.Logf("%T\n", smallBase)
}