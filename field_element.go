package xxx

import (
	"encoding/hex"
	"fmt"
	"io"
	"math/big"
	"math/bits"
)

type Fe256 [4]uint64

type FieldElement interface {
	Marshal(out []byte) []byte
	Unmarshal(in []byte) FieldElement
	SetBig(a *big.Int)
	SetUint(a uint64)
	SetString(s string) error
	Set(fe FieldElement)
	Big() *big.Int
	String() string
	IsOdd() bool
	IsEven() bool
	IsZero() bool
	IsOne() bool
	Cmp(fe FieldElement) int64
	Equals(fe FieldElement) bool
	limb(i int) uint64
	setLimb(i int, val uint64)
}

func (fe *Fe256) Marshal(out []byte) []byte {
	var a int
	for i := 0; i < 4; i++ {
		a = 4*8 - i*8
		out[a-1] = byte(fe[i])
		out[a-2] = byte(fe[i] >> 8)
		out[a-3] = byte(fe[i] >> 16)
		out[a-4] = byte(fe[i] >> 24)
		out[a-5] = byte(fe[i] >> 32)
		out[a-6] = byte(fe[i] >> 40)
		out[a-7] = byte(fe[i] >> 48)
		out[a-8] = byte(fe[i] >> 56)
	}
	return out
}

func (fe *Fe256) Unmarshal(in []byte) FieldElement {
	size := 4 * 8
	padded := make([]byte, size)
	l := len(in)
	if l >= size {
		l = size
	}
	copy(padded[size-l:], in[:])
	var a int
	for i := 0; i < 4; i++ {
		a = size - i*8
		fe[i] = uint64(padded[a-1]) | uint64(padded[a-2])<<8 |
			uint64(padded[a-3])<<16 | uint64(padded[a-4])<<24 |
			uint64(padded[a-5])<<32 | uint64(padded[a-6])<<40 |
			uint64(padded[a-7])<<48 | uint64(padded[a-8])<<56
	}
	return fe
}

func (fe *Fe256) SetBig(a *big.Int) {
	fe.Unmarshal(a.Bytes())
}

func (fe *Fe256) SetUint(a uint64) {
	fe[0] = a
	fe[1] = 0
	fe[2] = 0
	fe[3] = 0
}

func (fe *Fe256) SetString(s string) error {
	if s[:2] == "0x" {
		s = s[2:]
	}
	h, err := hex.DecodeString(s)
	if err != nil {
		return err
	}
	fe.Unmarshal(h)
	return nil
}

func (fe *Fe256) Set(fe2 FieldElement) {
	fe[0] = fe2.limb(0)
	fe[1] = fe2.limb(1)
	fe[1] = fe2.limb(2)
	fe[1] = fe2.limb(3)
}

func (fe *Fe256) limb(i int) uint64 {
	if i >= len(fe) {
		return 0
	}
	return fe[i]
}

func (fe *Fe256) setLimb(i int, val uint64) {
	if i >= len(fe) {
		return
	}
	fe[i] = val
}

func (fe *Fe256) Big() *big.Int {
	h := [4 * 8]byte{}
	return new(big.Int).SetBytes(fe.Marshal(h[:]))
}

func (fe Fe256) String() (s string) {
	for i := 3; i >= 0; i-- {
		s = fmt.Sprintf("%s%16.16x", s, fe[i])
	}
	return "0x" + s
}

func (fe *Fe256) IsOdd() bool {
	var mask uint64 = 1
	return fe[0]&mask != 0
}

func (fe *Fe256) IsEven() bool {
	var mask uint64 = 1
	return fe[0]&mask == 0
}

func (fe *Fe256) IsZero() bool {
	return 0 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3]
}

func (fe *Fe256) IsOne() bool {
	return 1 == fe[0] && 0 == fe[1] && 0 == fe[2] && 0 == fe[3]
}

func (fe *Fe256) Cmp(fe2 FieldElement) int64 {

	if fe[3] > fe2.limb(3) {
		return 1
	} else if fe[3] < fe2.limb(3) {
		return -1
	}
	if fe[2] > fe2.limb(2) {
		return 1
	} else if fe[2] < fe2.limb(2) {
		return -1
	}
	if fe[1] > fe2.limb(1) {
		return 1
	} else if fe[1] < fe2.limb(1) {
		return -1
	}
	if fe[0] > fe2.limb(0) {
		return 1
	} else if fe[0] < fe2.limb(0) {
		return -1
	}
	return 0
}

func (fe *Fe256) Equals(fe2 FieldElement) bool {
	return fe2.limb(0) == fe[0] && fe2.limb(1) == fe[1] && fe2.limb(2) == fe[2] && fe2.limb(3) == fe[3]
}

func (fe *Fe256) div2(e uint64) {
	fe[0] = fe[0]>>1 | fe[1]<<63
	fe[1] = fe[1]>>1 | fe[2]<<63
	fe[2] = fe[2]>>1 | fe[3]<<63
	fe[3] = fe[3]>>1 | e<<63
}

func (fe *Fe256) mul2() uint64 {
	e := fe[3] >> 63
	fe[3] = fe[3]<<1 | fe[2]>>63
	fe[2] = fe[2]<<1 | fe[1]>>63
	fe[1] = fe[1]<<1 | fe[0]>>63
	fe[0] = fe[0] << 1
	return e
}

func (fe *Fe256) bit(i int) bool {
	k := i >> 6
	i = i - k<<6
	b := (fe[k] >> uint(i)) & 1
	return b != 0
}

func (fe *Fe256) bitLen() int {
	for i := len(fe) - 1; i >= 0; i-- {
		if len := bits.Len64(fe[i]); len != 0 {
			return len + 64*i
		}
	}
	return 0
}

func (f *Fe256) rand(max *Fe256, r io.Reader) error {
	bitLen := bits.Len64(max[3]) + (4-1)*64
	k := (bitLen + 7) / 8
	b := uint(bitLen % 8)
	if b == 0 {
		b = 8
	}
	bytes := make([]byte, k)
	for {
		_, err := io.ReadFull(r, bytes)
		if err != nil {
			return err
		}
		bytes[0] &= uint8(int(1<<b) - 1)
		f.Unmarshal(bytes)
		if f.Cmp(max) < 0 {
			break
		}
	}
	return nil
}
