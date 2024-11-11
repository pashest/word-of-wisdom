package quotes

import (
	"math/rand"

	"github.com/pashest/word-of-wisdom/config"
)

// Service of quotes
type Service struct {
	quotes []string
}

// New creates service of quotes
func New(cfg *config.Config) *Service {
	return &Service{
		quotes: cfg.Quotes,
	}
}

// GetRandomQuote returns random quote
func (s Service) GetRandomQuote() string {
	return s.quotes[rand.Intn(len(s.quotes))]
}
