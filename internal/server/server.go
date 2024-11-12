package server

import (
	"bufio"
	_ "bufio"
	"context"
	"encoding/json"
	"fmt"
	_ "io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/pashest/word-of-wisdom/config"
	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/pashest/word-of-wisdom/internal/pkg/pow/equihash"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
)

// Server server struct
type Server struct {
	cfg *config.Config
	// handlers map[string]handler
	listener                    net.Listener
	shutdownCh                  chan struct{}
	rwMu                        sync.RWMutex
	parallelConnectionNum       atomic.Int64
	parallelConnectionThreshold int64
	requestCache                requestCache
	quoteService                quoteService
	algorithmSetting            algorithmSetting
}

func NewServer(cfg *config.Config,
	requestCache requestCache,
	quoteService quoteService,
	algorithmSetting algorithmSetting,
) *Server {
	return &Server{
		cfg:                         cfg,
		shutdownCh:                  make(chan struct{}),
		parallelConnectionThreshold: cfg.TcpServer.ParallelConnectionsThreshold,
		requestCache:                requestCache,
		quoteService:                quoteService,
		algorithmSetting:            algorithmSetting,
	}
}

func (s *Server) Run(ctx context.Context) error {
	var err error
	address := fmt.Sprintf("%s:%d", s.cfg.TcpServer.ServerHost, s.cfg.TcpServer.ServerPort)
	s.listener, err = net.Listen("tcp", address)
	if err != nil {
		return errors.Wrap(err, "failed to listen")
	}

	log.Info().Msg("server started")

	go s.controlDifficulty()

	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.shutdownCh:
				return nil
			default:
				log.Fatal().Err(err)
				return err
			}
		}

		s.parallelConnectionNum.Add(1)
		go s.handleConnection(conn)
	}
}

// Stop stops server
func (s *Server) Stop() {
	close(s.shutdownCh)
	if err := s.listener.Close(); err != nil {
		log.Fatal().Err(err)
	}
	log.Info().Msg("server stopped")
}

func (s *Server) controlDifficulty() {
	prevMaxCount := s.parallelConnectionThreshold
	for {
		select {
		case <-s.shutdownCh:
			return
		default:
			time.Sleep(250 * time.Millisecond)
			connNum := s.parallelConnectionNum.Load()
			if connNum < s.parallelConnectionThreshold {
				if s.algorithmSetting.IsMinDifficulty() {
					continue
				}
				s.rwMu.Lock()
				s.algorithmSetting.DecreaseDifficulty()
				s.rwMu.Unlock()
			} else {
				if connNum > prevMaxCount && !s.algorithmSetting.IsMaxDifficulty() {
					s.rwMu.Lock()
					s.algorithmSetting.IncreaseDifficulty()
					s.rwMu.Unlock()
				}
				prevMaxCount = connNum
			}
		}
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func() {
		conn.Close()
		s.parallelConnectionNum.Add(-1)
	}()

	err := conn.SetDeadline(time.Now().Add(20 * time.Second))
	if err != nil {
		log.Error().Err(err).Msg("failed to set deadline")
	}
	reader := bufio.NewReader(conn)

	for {
		select {
		case <-s.shutdownCh:
			return
		default:
		}

		req, err := reader.ReadString('\n')
		if err != nil {
			log.Error().Err(err).Msg("failed to read connection")
			return
		}
		msg, err := s.ProcessRequest(req)
		if err != nil {
			log.Error().Err(err).Msg("failed to process request")
			return
		}

		if msg == nil {
			continue
		}

		msgStr := fmt.Sprintf("%s\n", msg.Stringify())
		_, err = conn.Write([]byte(msgStr))
		if err != nil {
			log.Error().Err(err).Msg("failed to send message")
		}
	}
}

func (s *Server) ProcessRequest(msgStr string) (*model.Message, error) {
	msg, err := utils.ParseMessage(msgStr)
	if err != nil {
		return nil, err
	}

	switch msg.Type {
	case model.RequestChallenge:
		s.rwMu.RLock()
		currentDifficulty := s.algorithmSetting.GetDifficulty()
		s.rwMu.RUnlock()

		challenge := model.NewChallenge(equihash.EquihashAlgorithm, currentDifficulty)

		reqID := utils.GetRandomString(20)
		s.requestCache.Set(reqID, &challenge)
		if err != nil {
			return nil, fmt.Errorf("err add rand to cache: %w", err)
		}

		challengeMarshaled, err := json.Marshal(challenge)
		if err != nil {
			return nil, fmt.Errorf("err marshal challenge: %v", err)
		}
		msg := model.Message{
			Type:      model.ResponseChallenge,
			RequestID: reqID,
			Payload:   string(challengeMarshaled),
		}
		return &msg, nil
	case model.RequestResource:
		reqID := msg.RequestID
		var proof equihash.Proof
		err := json.Unmarshal([]byte(msg.Payload), &proof)
		if err != nil {
			return nil, fmt.Errorf("err unmarshal proof: %v", err)
		}

		ch, exists := s.requestCache.Get(reqID)
		if !exists || ch == nil {
			return &model.Message{
				Type:      model.FailedResponseResource,
				RequestID: reqID,
				Payload:   "challenge expired or not sent",
			}, nil
		}
		s.requestCache.Delete(reqID)

		if ok := proof.ValidateChallenge(*ch); !ok {
			return &model.Message{
				Type:      model.FailedResponseResource,
				RequestID: reqID,
				Payload:   "invalid proof",
			}, nil
		}

		return &model.Message{
			Type:      model.SuccessResponseResource,
			RequestID: reqID,
			Payload:   s.quoteService.GetRandomQuote(),
		}, nil
	default:
		return &model.Message{
			Type: model.Unknown,
		}, nil
	}
}
