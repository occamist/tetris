package game

import "math/rand/v2"

// Randomizer supplies the sequence of pieces.
type Randomizer interface {
	Next() Kind
}

// NESRandomizer picks pieces the way the NES did: roll one of eight options,
// and if the roll is the unused eighth option or repeats the previous piece,
// roll once more among the seven real pieces and accept that result. Repeats
// are therefore possible but rarer than with a plain uniform pick.
type NESRandomizer struct {
	intN func(n int) int
	last Kind
}

// NewNESRandomizer returns a randomizer seeded with seed.
func NewNESRandomizer(seed uint64) *NESRandomizer {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	return newNESRandomizer(r.IntN)
}

func newNESRandomizer(intN func(n int) int) *NESRandomizer {
	return &NESRandomizer{intN: intN}
}

// Next returns the next piece kind.
func (r *NESRandomizer) Next() Kind {
	k := Kind(r.intN(NumKinds+1)) + 1
	if k > NumKinds || k == r.last {
		k = Kind(r.intN(NumKinds)) + 1
	}
	r.last = k
	return k
}
