package pow

type PoWChalenge interface {
	GenerateChallenge() []byte
	ValidateSolution(challenge []byte, nonce []byte) bool
	Solve(challenge []byte) []byte
}
