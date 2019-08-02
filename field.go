package xxx

import (
	"crypto/rand"
	"io"
	"math/big"
)

var inp uint64
var modulus4 Fe256

func ceil(len int) int {
	size := 1 + (len / 64)
	if size < 5 {
		return 4
	}
	return size
}

const (
	_256BIT = 4
	_320BIT = 5
	_384BIT = 6
)

type Field interface {
	Modulus() FieldElement
	Zero() FieldElement
	One() FieldElement
	Copy(dst FieldElement, src FieldElement)
	RandElement(fe FieldElement, r io.Reader) error
	Add(c, a, b FieldElement)
}

/*

Field impl for 256 bit prime

*/

type Field256 struct {
	pBig *big.Int
	r1   FieldElement
	r2   FieldElement
	P    FieldElement
}

func NewField(p []byte) Field {
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return &Field256{}
	}
	inp = inpT.Uint64()
	var r1 FieldElement
	var r2 FieldElement
	size := ceil(len(p))
	switch size {
	case _256BIT:
		r1, r2 = &Fe256{}, &Fe256{}
		modulus4.Unmarshal(p)
		r1Big := new(big.Int).SetBit(new(big.Int), 256, 1)
		r1.SetBig(new(big.Int).Mod(r1Big, pBig))
		r2.SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
		field := &Field256{
			pBig: pBig,
			r1:   r1,
			r2:   r2,
			P:    &modulus4,
		}
		return field
	case _320BIT:
		// ...
		return nil
	case _384BIT:
		// ...
		return nil
	}
	return nil
}

func (f *Field256) Add(c, a, b FieldElement) {
	add4(c.(*Fe256), a.(*Fe256), b.(*Fe256))
}

func (f *Field256) Modulus() FieldElement {
	return f.P
}

func (f *Field256) Zero() FieldElement {
	fe := &Fe256{}
	fe.SetUint(0)
	return fe
}

func (f *Field256) One() FieldElement {
	fe := &Fe256{}
	fe.Set(f.r1)
	return fe
}

func (f *Field256) Copy(dst FieldElement, src FieldElement) {
	dst.Set(src)
}

func (f *Field256) RandElement(fe FieldElement, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}

/*

Field impl for any bit size prime

*/

type FieldImpl struct {
	pBig *big.Int
	r1   FieldElement
	r2   FieldElement
	P    FieldElement
	add  func(c, a, b FieldElement)
}

func NewFieldImpl(p []byte) *FieldImpl {
	pBig := new(big.Int).SetBytes(p)
	inpT := new(big.Int).ModInverse(new(big.Int).Neg(pBig), new(big.Int).SetBit(new(big.Int), 64, 1))
	if inpT == nil {
		return &FieldImpl{}
	}
	inp = inpT.Uint64()
	var r1 FieldElement
	var r2 FieldElement
	size := ceil(len(p))
	switch size {
	case _256BIT:
		r1, r2 = &Fe256{}, &Fe256{}
		modulus4.Unmarshal(p)
		r1Big := new(big.Int).SetBit(new(big.Int), 256, 1)
		r1.SetBig(new(big.Int).Mod(r1Big, pBig))
		r2.SetBig(new(big.Int).Exp(r1Big, new(big.Int).SetUint64(2), pBig))
		return &FieldImpl{
			pBig: pBig,
			r1:   r1,
			r2:   r2,
			P:    &modulus4,
			/*

				here we register primitive function 'add' for given bit size

			*/
			add: Add4,
		}
	case _320BIT:
		// ...
		return nil
	case _384BIT:
		// ...
		return nil
	}
	return nil
}

func (f *FieldImpl) Add(c, a, b FieldElement) {
	f.add(c, a, b)
}

func (f *FieldImpl) Modulus() FieldElement {
	return f.P
}

func (f *FieldImpl) Zero() FieldElement {
	fe := &Fe256{}
	fe.SetUint(0)
	return fe
}

func (f *FieldImpl) One() FieldElement {
	fe := &Fe256{}
	fe.Set(f.r1)
	return fe
}

func (f *FieldImpl) Copy(dst FieldElement, src FieldElement) {
	dst.Set(src)
}

func (f *FieldImpl) RandElement(fe FieldElement, r io.Reader) error {
	bi, err := rand.Int(r, f.pBig)
	if err != nil {
		return err
	}
	fe.SetBig(bi)
	return nil
}
