package server

import "github.com/pashest/word-of-wisdom/internal/model"

type quoteService interface {
	GetRandomQuote() string
}

type requestCache interface {
	Get(key string) (*model.Challenge, bool)
	Set(key string, challenge *model.Challenge)
	Delete(key string)
}

type algorithmSetting interface {
	IncreaseDifficulty()
	DecreaseDifficulty()
	IsMaxDifficulty() bool
	IsMinDifficulty() bool
	GetDifficulty() model.Difficulty
}
