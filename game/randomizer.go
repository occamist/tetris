package game

import "math/rand/v2"

// NESRandomizer picks pieces the way the NES did: roll one of eight options,
// and if the roll is the unused eighth option or repeats the previous piece,
// roll once more among the seven real pieces and accept that result. Repeats
// are therefore possible but rarer than with a plain uniform pick.
type NESRandomizer struct {
	intN func(n int) int
	last Kind
}

func NewNESRandomizer(seed uint64) *NESRandomizer {
	r := rand.New(rand.NewPCG(seed, seed^0x9e3779b97f4a7c15))
	return &NESRandomizer{intN: r.IntN}
}

func (r *NESRandomizer) NextKind() Kind {
	k := Kind(r.intN(NumKinds+1)) + 1
	if k > NumKinds || k == r.last {
		k = Kind(r.intN(NumKinds)) + 1
	}
	r.last = k
	return k
}
