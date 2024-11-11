package equihash

import (
	"fmt"

	"github.com/pashest/word-of-wisdom/config"
	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/rs/zerolog/log"
)

type setting struct {
	difficulty       model.Difficulty
	currentLevel     int
	difficultyLevels []config.Difficulty
}

// NewSetting struct
func NewSetting(cfg *config.Config) (*setting, error) {
	difficulties := cfg.Equihash.Difficulties
	if len(difficulties) == 0 {
		return nil, fmt.Errorf("there aren't difficulties in config")
	}

	return &setting{
		difficulty: model.Difficulty{
			NumOfBits: difficulties[0].N,
			Length:    difficulties[0].K,
		},
		currentLevel:     0,
		difficultyLevels: difficulties,
	}, nil
}

// IncreaseDifficulty method increases difficulty until max value
func (s *setting) IncreaseDifficulty() {
	if !s.IsMaxDifficulty() {
		s.currentLevel++
		s.difficulty.NumOfBits = s.difficultyLevels[s.currentLevel].N
		s.difficulty.Length = s.difficultyLevels[s.currentLevel].K
		log.Info().Msg(fmt.Sprintf("The difficulty has been increased N: %d, K: %d", s.difficulty.NumOfBits, s.difficulty.Length))
	}
}

// DecreaseDifficulty method decreases difficulty until basic value
func (s *setting) DecreaseDifficulty() {
	if !s.IsMinDifficulty() {
		s.currentLevel--
		s.difficulty.NumOfBits = s.difficultyLevels[s.currentLevel].N
		s.difficulty.Length = s.difficultyLevels[s.currentLevel].K
		log.Info().Msg(fmt.Sprintf("The difficulty has been decreased N: %d, K: %d", s.difficulty.NumOfBits, s.difficulty.Length))
	}
}

// IsMaxDifficulty method returns true if difficulty has max value
func (s *setting) IsMaxDifficulty() bool {
	return s.currentLevel == len(s.difficultyLevels)-1
}

// IsMinDifficulty method decreases difficulty
func (s *setting) IsMinDifficulty() bool {
	return s.currentLevel == 0
}

// GetDifficulty method returns difficulty
func (s *setting) GetDifficulty() model.Difficulty {
	return s.difficulty
}
