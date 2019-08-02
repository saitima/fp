package xxx

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

func testSuite(prime []byte, elements ...FieldElement) Field {
	field := NewField(prime)
	for i := 0; i < len(elements); i++ {
		if err := field.RandElement(elements[i], rand.Reader); err != nil {
			panic(err)
		}
	}
	return field
}

func testSuite2(prime []byte, elements ...FieldElement) *FieldImpl {
	field := NewFieldImpl(prime)
	for i := 0; i < len(elements); i++ {
		if err := field.RandElement(elements[i], rand.Reader); err != nil {
			panic(err)
		}
	}
	return field
}

func BenchmarkAdd256(t *testing.B) {
	a, b, c := &Fe256{}, &Fe256{}, &Fe256{}
	pStr := "0x73eda753299d7d483339d80809a1d80553bda402fffe5bfeffffffff00000001"
	prime, _ := hex.DecodeString(pStr[2:])
	field1 := testSuite(prime, a, b, c)
	field2 := testSuite(prime, a, b, c).(*Field256)
	field3 := testSuite2(prime, a, b, c)
	t.Run("interface", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field1.Add(c, a, b)
			field1.Add(a, b, c)
			field1.Add(c, a, b)
			field1.Add(a, b, c)
		}
	})
	t.Run("casted to 256 bit field", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field2.Add(c, a, b)
			field2.Add(a, b, c)
			field2.Add(c, a, b)
			field2.Add(a, b, c)
		}
	})
	t.Run("general purpose impl", func(t *testing.B) {
		t.ResetTimer()
		for i := 0; i < t.N; i++ {
			field3.Add(c, a, b)
			field3.Add(a, b, c)
			field3.Add(c, a, b)
			field3.Add(a, b, c)
		}
	})
}
