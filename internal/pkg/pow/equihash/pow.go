package equihash

import (
	"encoding/binary"

	"github.com/dchest/blake2b"
	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
)

const (
	EquihashAlgorithm model.ChallengeAlgorithm = "equihash"
)

type Proof struct {
	N      uint32             `json:"n"`
	K      uint32             `json:"k"`
	Seed   [seedLength]uint32 `json:"seed"`
	Nonce  uint32             `json:"nonce"`
	Inputs []uint32           `json:"inputs"`
}

func NewProof(
	n int,
	k int,
	seed []byte,
	nonce int,
	inputs []byte,
) *Proof {
	var s [seedLength]uint32
	for i, c := range utils.BytesToUint32Array(seed, binary.LittleEndian) {
		if i >= seedLength {
			break
		}

		s[i] = c
	}

	inp := make([]uint32, 0, len(inputs))
	for _, c := range utils.BytesToUint32Array(inputs, binary.LittleEndian) {
		inp = append(inp, c)
	}

	return &Proof{
		N:      uint32(n),
		K:      uint32(k),
		Seed:   s,
		Nonce:  uint32(nonce),
		Inputs: inp,
	}
}

func (p *Proof) ValidateChallenge(ch model.Challenge) bool {
	if ch.Algorithm != EquihashAlgorithm {
		return false
	}
	eq := NewEquihash(ch.Difficulty, ch.Input)
	if eq.seed == p.Seed && eq.n == p.N && eq.k == p.K {
		return p.ValidateSolution()
	}
	return false
}

func (p *Proof) GetInputsBytes() []byte {
	return utils.Uint32ArrayToBytes(p.Inputs, binary.LittleEndian)
}

func (p *Proof) ValidateSolution() bool {
	input := make([]uint32, seedLength+2)
	copy(input[:seedLength], p.Seed[:])
	input[seedLength] = p.Nonce
	buf := make([]uint32, maxN/4)
	blocks := make([]uint32, p.K+1)

	h := blake2b.New256()

	if len(p.Inputs) == 0 {
		return false
	}

	for i := range p.Inputs {
		input[seedLength+1] = uint32(p.Inputs[i])

		h.Reset()
		h.Write(utils.Uint32ArrayToBytes(input[:], binary.LittleEndian))
		buf = utils.BytesToUint32Array(h.Sum(nil), binary.LittleEndian)

		for j := uint32(0); j < (p.K + 1); j++ {
			blocks[j] ^= buf[j] >> (32 - p.N/(p.K+1))
		}
	}

	for _, block := range blocks {
		if block != 0 {
			return false
		}
	}

	return true
}
