package equihash

import (
	"encoding/binary"
	"sort"

	"github.com/dchest/blake2b"
	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
)

const (
	seedLength = 4
	maxN       = 32
	listLength = 5
	maxNonce   = 0xFFFFF

	forkMultiplier = 3
)

type fork struct {
	ref1 uint32
	ref2 uint32
}

type tuple struct {
	blocks    []uint32
	reference uint32
}

type Equihash struct {
	n          uint32
	k          uint32
	seed       [seedLength]uint32
	nonce      uint32
	tupleList  [][]tuple
	filledList []uint32
	solutions  []Proof
	forks      [][]fork
}

func NewEquihash(difficulty model.Difficulty, input []byte) *Equihash {
	var s [seedLength]uint32
	for i, c := range utils.BytesToUint32Array(input, binary.LittleEndian) {
		if i >= seedLength {
			break
		}

		s[i] = c
	}

	return &Equihash{
		n:     uint32(difficulty.NumOfBits),
		k:     uint32(difficulty.Length),
		seed:  s,
		nonce: 1,
	}
}

func (eq *Equihash) initializeMemory() {
	tupleN := 1 << (eq.n / (eq.k + 1))
	eq.tupleList = make([][]tuple, tupleN)
	for i := range eq.tupleList {
		defTuples := make([]tuple, listLength)
		for i := range defTuples {
			defTuples[i] = tuple{blocks: make([]uint32, eq.k)}
		}

		eq.tupleList[i] = defTuples
	}
	eq.filledList = make([]uint32, tupleN)
	eq.solutions = make([]Proof, 0)
	eq.forks = make([][]fork, 0)
}

func (eq *Equihash) fillMemory(length uint32) {
	input := make([]uint32, seedLength+2)
	copy(input[:seedLength], eq.seed[:])
	input[seedLength] = eq.nonce
	buf := make([]uint32, maxN/4)
	h := blake2b.New256()
	for i := uint32(0); i < length; i++ {
		input[seedLength+1] = i

		h.Reset()
		h.Write(utils.Uint32ArrayToBytes(input[:], binary.LittleEndian))
		buf = utils.BytesToUint32Array(h.Sum(nil), binary.LittleEndian)

		index := buf[0] >> (32 - eq.n/(eq.k+1))
		count := eq.filledList[index]
		if count < listLength {
			for j := uint32(1); j < (eq.k + 1); j++ {
				eq.tupleList[index][count].blocks[j-1] = buf[j] >> (32 - eq.n/(eq.k+1))
			}
			eq.tupleList[index][count].reference = i
			eq.filledList[index]++
		}
	}
}

func (eq *Equihash) resolveCollisions(store bool) {
	tableLength := len(eq.tupleList)
	maxNewCollisions := len(eq.tupleList) * forkMultiplier
	newBlocks := len(eq.tupleList[0][0].blocks) - 1
	newForks := make([]fork, maxNewCollisions)
	collisionList := make([][]tuple, tableLength)
	for i := range collisionList {
		tableRow := make([]tuple, listLength)
		for i := range tableRow {
			tableRow[i] = tuple{blocks: make([]uint32, newBlocks)}
		}
		collisionList[i] = tableRow
	}
	newFilledList := make([]uint32, tableLength)
	newColls := uint32(0)
	for i := uint32(0); i < uint32(tableLength); i++ {
		for j := uint32(0); j < eq.filledList[i]; j++ {
			for m := j + 1; m < eq.filledList[i]; m++ {
				newIndex := eq.tupleList[i][j].blocks[0] ^ eq.tupleList[i][m].blocks[0]
				newFork := fork{ref1: eq.tupleList[i][j].reference, ref2: eq.tupleList[i][m].reference}
				if store {
					if newIndex == 0 {
						solutionInputs := eq.resolveTree(newFork)
						eq.solutions = append(
							eq.solutions,
							Proof{N: eq.n, K: eq.k, Seed: eq.seed, Nonce: eq.nonce, Inputs: solutionInputs},
						)
					}
				} else {
					if newFilledList[newIndex] < listLength && newColls < uint32(maxNewCollisions) {
						for l := 0; l < newBlocks; l++ {
							collisionList[newIndex][newFilledList[newIndex]].blocks[l] = eq.tupleList[i][j].blocks[l+1] ^ eq.tupleList[i][m].blocks[l+1]
						}
						newForks[newColls] = newFork
						collisionList[newIndex][newFilledList[newIndex]].reference = newColls
						newFilledList[newIndex]++
						newColls++
					}
				}
			}
		}
	}
	eq.forks = append(eq.forks, newForks)
	eq.tupleList, collisionList = collisionList, eq.tupleList
	eq.filledList, newFilledList = newFilledList, eq.filledList
}

func (eq *Equihash) resolveTreeByLevel(fork fork, level uint32) []uint32 {
	if level == 0 {
		return []uint32{fork.ref1, fork.ref2}
	}
	v1 := eq.resolveTreeByLevel(eq.forks[level-1][fork.ref1], level-1)
	v2 := eq.resolveTreeByLevel(eq.forks[level-1][fork.ref2], level-1)

	result := make([]uint32, len(v1)+len(v2))
	for i, el := range append(v1, v2...) {
		result[i] = el
	}
	return result
}

func (eq *Equihash) resolveTree(fork fork) []uint32 {
	return eq.resolveTreeByLevel(fork, uint32(len(eq.forks)))
}

func (eq *Equihash) FindProof() Proof {
	eq.nonce = 1
	for eq.nonce < maxNonce {
		eq.nonce++
		eq.initializeMemory()
		eq.fillMemory(4 << (eq.n/(eq.k+1) - 1))
		for i := uint32(1); i <= eq.k; i++ {
			toStore := (i == eq.k)
			eq.resolveCollisions(toStore)
		}
		for i := range eq.solutions {
			vec := eq.solutions[i].Inputs
			sort.Slice(vec, func(i, j int) bool { return vec[i] < vec[j] })
			dup := false
			for k := range vec[:len(vec)-1] {
				if vec[k] == vec[k+1] {
					dup = true
				}
			}
			if !dup {
				return eq.solutions[i]
			}
		}
	}
	return Proof{N: eq.n, K: eq.k, Seed: eq.seed, Nonce: eq.nonce, Inputs: []uint32{}}
}
