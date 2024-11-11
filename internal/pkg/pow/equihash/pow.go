package equihash

import (
	"encoding/binary"

	"github.com/dchest/blake2b"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
)

type Proof struct {
	n      uint32
	k      uint32
	seed   [seedLength]uint32
	nonce  uint32
	inputs []uint32
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
		n:      uint32(n),
		k:      uint32(k),
		seed:   s,
		nonce:  uint32(nonce),
		inputs: inp,
	}
}

func (p *Proof) GetInputsBytes() []byte {
	return utils.Uint32ArrayToBytes(p.inputs, binary.LittleEndian)
}

func (p *Proof) ValidateSolution() bool {
	input := make([]uint32, seedLength+2)
	copy(input[:seedLength], p.seed[:])
	input[seedLength] = p.nonce
	buf := make([]uint32, maxN/4)
	blocks := make([]uint32, p.k+1)

	h := blake2b.New256()

	if len(p.inputs) == 0 {
		return false
	}

	for i := range p.inputs {
		input[seedLength+1] = uint32(p.inputs[i])

		h.Reset()
		h.Write(utils.Uint32ArrayToBytes(input[:], binary.LittleEndian))
		buf = utils.BytesToUint32Array(h.Sum(nil), binary.LittleEndian)

		for j := uint32(0); j < (p.k + 1); j++ {
			blocks[j] ^= buf[j] >> (32 - p.n/(p.k+1))
		}
	}

	for _, block := range blocks {
		if block != 0 {
			return false
		}
	}

	return true
}
