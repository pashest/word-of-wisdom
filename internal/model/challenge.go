package model

import "crypto/rand"

type Difficulty struct {
	// The width (number of bits)
	NumOfBits int
	// The length of the generalized birthday problem
	Length int
}

type ChallengeAlgorithm string

const (
	EquihashAlgorithm ChallengeAlgorithm = "equihash"
)

type Challenge struct {
	Algorithm  ChallengeAlgorithm
	Difficulty Difficulty
	Input      []byte
}

func NewChallenge(algo ChallengeAlgorithm, dfc Difficulty) Challenge {
	seed := make([]byte, 32)
	_, _ = rand.Read(seed)
	return Challenge{
		Algorithm:  algo,
		Difficulty: dfc,
		Input:      seed,
	}
}
