package equihash

import (
	"encoding/binary"
	"testing"

	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
	"github.com/stretchr/testify/assert"
)

func Test_equihashFindProof(t *testing.T) {
	tests := []struct {
		name     string
		n        int
		k        int
		seed     []uint32
		expected Proof
	}{
		{
			name: "60_3_1",
			n:    60,
			k:    3,
			seed: []uint32{1, 1, 1, 1},
			expected: *NewProof(
				60,
				3,
				utils.Uint32ArrayToBytes([]uint32{1, 1, 1, 1}, binary.LittleEndian),
				2,
				utils.Uint32ArrayToBytes([]uint32{0x46c3, 0x4cb5, 0x6072, 0x812e, 0xa3ec, 0xad88, 0xbc6a, 0xe480}, binary.LittleEndian),
			),
		},
		{
			name: "60_3_1-8",
			n:    60,
			k:    3,
			seed: []uint32{1, 2, 3, 4, 5, 6, 7, 8},
			expected: *NewProof(
				60,
				3,
				utils.Uint32ArrayToBytes([]uint32{1, 2, 3, 4, 5, 6, 7, 8}, binary.LittleEndian),
				3,
				utils.Uint32ArrayToBytes([]uint32{0x4b02, 0x4b64, 0x653b, 0x6b5e, 0x77e6, 0x9708, 0xd873, 0xf39f}, binary.LittleEndian),
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq := NewEquihash(model.Difficulty{
				NumOfBits: tt.n,
				Length:    tt.k,
			}, utils.Uint32ArrayToBytes(tt.seed, binary.LittleEndian))

			proof := eq.FindProof()

			assert.Equal(t, tt.expected, proof)
		})
	}
}

func Test_validateSolution(t *testing.T) {
	tests := []struct {
		name     string
		proof    Proof
		expected bool
	}{
		{
			name: "60_3",
			proof: *NewProof(
				60,
				3,
				utils.Uint32ArrayToBytes([]uint32{1, 1, 1, 1}, binary.LittleEndian),
				2,
				utils.Uint32ArrayToBytes([]uint32{0x46c3, 0x4cb5, 0x6072, 0x812e, 0xa3ec, 0xad88, 0xbc6a, 0xe480}, binary.LittleEndian),
			),
			expected: true,
		},
		{
			name: "60_3_wrong",
			proof: *NewProof(
				60,
				3,
				utils.Uint32ArrayToBytes([]uint32{1, 1, 1, 1}, binary.LittleEndian),
				2,
				utils.Uint32ArrayToBytes([]uint32{0x610, 0x1626, 0x1c37, 0x20cb, 0x241d, 0x30d7, 0x3811, 0x395c}, binary.LittleEndian),
			),
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.proof.ValidateSolution()
			assert.Equal(t, test.expected, result)
		})
	}
}
