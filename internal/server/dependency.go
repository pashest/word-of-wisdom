package server

import "github.com/pashest/word-of-wisdom/internal/model"

type quoteService interface {
	GetRandomQuote() string
}

type requestCache interface {
	Get(key string) bool
	Set(key string)
	Delete(key string)
}

type algorithmSetting interface {
	IncreaseDifficulty()
	DecreaseDifficulty()
	IsMaxDifficulty() bool
	IsMinDifficulty() bool
	GetDifficulty() model.Difficulty
}
