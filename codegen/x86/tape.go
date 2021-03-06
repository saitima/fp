package main

import (
	. "github.com/mmcloughlin/avo/build"
	. "github.com/mmcloughlin/avo/operand"
	. "github.com/mmcloughlin/avo/reg"
)

func isMem(op Op) bool { return IsM64(op) }

func isGp(op Op) bool {
	return IsR64(op)
}

type tape struct {
	gpSet *gpSet
	stack *stack
	swap  GPPhysical
}

func newTape(swap GPPhysical, reservedGps ...Op) *tape {
	gpSet := newGpSet(R8, R9, R10, R11, R12, R13, R14, R15, RCX, RAX, RDX, RBX, RDI, RSI)
	gpSet.reserve(reservedGps...)
	gpSet.reserve(swap)
	stack := newStack()
	return &tape{gpSet, stack, swap}
}

func (t tape) newReprNoAlloc(size int) *repr {
	return newReprEmpty(size, t.swap)
}

func (t tape) newReprAlloc(size int) *repr {
	r := newReprEmpty(size, t.swap)
	for i := 0; i < size; i++ {
		r.limbs[i].set(t.next(_ALLOC))
	}
	return r
}

func (t *tape) newReprAtParam(size int, param string, dst Register) *repr {
	if _, ok := dst.(GPPhysical); ok {
		t.reserveGp(dst.(GPPhysical))
	}
	return t.newReprAtMemory(size, Mem{Base: Load(Param(param), dst)})
}

func (t *tape) newReprAtMemory(size int, base Mem) *repr {
	number := make([]limb, size)
	for i := 0; i < size; i++ {
		number[i] = newLimb(base.Offset(int(i*8)), t.swap)
	}
	return &repr{number, 0, size, base.Base, t.swap}
}

func (t *tape) next(allocated bool) Op {
	if op := t.gpSet.next(allocated); op != nil {
		return op
	}
	return t.stack.next(allocated)
}

func (t tape) free(ops ...Op) {
	for i := 0; i < len(ops); i++ {
		op := ops[i]
		if isLimb(op) {
			op = op.(limb).s
		}
		if isGp(op) {
			t.gpSet.free(op)
		} else if isMem(op) {
			t.stack.free(op)
		}
	}
}

func (t tape) donate(r *repr) Op {
	var a Op
	for j := r.size - 1; ; j-- {
		if r.limbs[j].atReg() {
			a = r.limbs[j].s
			r.limbs[j].moveTo(t.next(_ALLOC), _ASSIGN)
			return a
		}
	}
	return nil
}

func (t tape) freeGp(gps ...Op) []Op {
	return t.gpSet.free(gps...)
}

func (t tape) reserveGp(gps ...Op) []Op {
	return t.gpSet.reserve(gps...)
}

func (t tape) sizeFreeGp() int {
	return t.gpSet.sizeFree()
}

func (t tape) ret() {
	t.stack.allocLocal()
}

type gpSet struct {
	allocated map[GPPhysical]bool
	regs      map[int]GPPhysical
	size      int
}

func newGpSet(regs ...Op) *gpSet {
	allocated := make(map[GPPhysical]bool)
	regs_ := make(map[int]GPPhysical)
	for i, reg := range regs {
		if reg, ok := reg.(GPPhysical); ok {
			allocated[reg] = false
			regs_[i] = reg
		} else {
			panic("bad operand for general purpose set")
		}
	}
	return &gpSet{allocated: allocated, regs: regs_, size: len(regs)}
}

func (set *gpSet) allocate(size int) ([]Op, int) {
	allocated := []Op{}
	var i int
	for i = 0; i < size; i++ {
		r := set.next(_ALLOC)
		if r == nil {
			break
		}
		allocated = append(allocated, r)
	}
	return allocated, i
}

func (set *gpSet) reserve(regs ...Op) []Op {
	regs_ := []Op{}
	for _, reg := range regs {
		if reg, ok := reg.(GPPhysical); ok {
			set.allocated[reg] = true
			regs_ = append(regs_, reg)
		} else {
			panic("bad operand for general purpose set")
		}
	}
	return regs_
}

func (set *gpSet) free(regs ...Op) []Op {
	regs_ := []Op{}
	for _, reg := range regs {
		if isLimb(reg) {
			reg = reg.(limb).s
		}
		if reg, ok := reg.(GPPhysical); ok {
			set.allocated[reg] = false
			regs_ = append(regs_, reg)
		}
	}
	return regs
}

func (set *gpSet) freeAll() {
	for r := range set.allocated {
		set.allocated[r] = false
	}
}

func (set *gpSet) next(allocate bool) GPPhysical {
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if !set.allocated[r] {
			set.allocated[r] = allocate
			return r
		}
	}
	return nil
}

func (set *gpSet) sizeFree() int {
	c := 0
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if !set.allocated[r] {
			c++
		}
	}
	return c
}

func (set *gpSet) sizeAllocated() int {
	c := 0
	for i := 0; i < set.size; i++ {
		r := set.regs[i]
		if set.allocated[r] {
			c++
		}
	}
	return c
}

func (set *gpSet) slice() []Op {
	regs := make([]Op, set.size)
	for i, r := range set.regs {
		regs[i] = r
	}
	return regs
}

// stack manager with 8 byte slots
type stack struct {
	head      Mem
	allocated map[int]bool
	size      int
}

func newStack() *stack {
	allocated := make(map[int]bool)
	return &stack{
		head:      NewStackAddr(0),
		allocated: allocated,
		size:      0,
	}
}

func (s *stack) allocLocal() {
	AllocLocal(s.size * 8)
}

func (s *stack) extend(size int, allocate bool) Mem {
	offset := s.size * 8
	for i := s.size; i < s.size+size; i++ {
		s.allocated[i] = allocate
	}
	s.size += size
	return s.head.Offset(offset)
}

func (s *stack) next(allocate bool) Mem {
	// look up for free stack slot
	for i := 0; i < s.size; i++ {
		if !s.allocated[i] {
			s.allocated[i] = allocate
			return s.head.Offset(8 * i)
		}
	}
	// else extend by one
	return s.extend(1, allocate)
}

func (s *stack) free(mems ...Op) []Op {
	_mems := []Op{}
	for _, op := range mems {
		// todo : consider removing limb
		if isLimb(op) {
			op = op.(limb).s
		}
		if isMem(op) {
			mem := op.(Mem)
			s.allocated[mem.Disp/8] = false
			_mems = append(_mems, op)
		}
	}
	return _mems
}

func (s stack) sizeFree() int {
	c := 0
	for i := 0; i < s.size; i++ {
		if !s.allocated[i] {
			c++
		}
	}
	return c
}

func (s stack) sizeAllocated() int {
	c := 0
	for i := 0; i < s.size; i++ {
		if s.allocated[i] {
			c++
		}
	}
	return c
}
