package main

import (
	"context"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/pashest/word-of-wisdom/config"
	"github.com/rs/zerolog/log"
)

const baseDifficulty = 4
const maxDifficulty = 8
const threshold = 3 // Threshold for request rate to detect possible DDoS
const monitorInterval = 10 * time.Second

var difficulty = baseDifficulty
var requestCount = int64(0)
var mu sync.RWMutex
var quotes []string

func regSignalHandler(ctx context.Context) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		defer signal.Stop(done)
		<-done
		log.Info().Msg("stop signal received")
		cancel()
	}()

	return ctx
}

func runServer(ctx context.Context) error {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		return err
	}
	defer listener.Close()
	log.Info().Msg("server started")

	go adjustDifficulty()
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				log.Error().Err(err).Msg("error accepting")
				return
			}
			go handleConnection(conn)
		}
	}()

	select {
	case <-ctx.Done():
		log.Info().Msg("server was closed")
		break
	}
	return nil

}

func generateChallenge() string {
	rand.Seed(time.Now().UnixNano())
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	challenge := make([]rune, 20)
	for i := range challenge {
		challenge[i] = letters[rand.Intn(len(letters))]
	}
	return string(challenge)
}

func verifyProofOfWork(challenge string, nonce int, difficulty int) bool {
	data := fmt.Sprintf("%s%d", challenge, nonce)
	hash := sha256.Sum256([]byte(data))
	hashString := fmt.Sprintf("%x", hash)
	return strings.HasPrefix(hashString, strings.Repeat("0", difficulty))
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	atomic.AddInt64(&requestCount, 1)
	mu.RLock()
	currentDifficulty := difficulty
	mu.RUnlock()
	log.Info().Msg(fmt.Sprintf("current difficulty: %d", currentDifficulty))

	challenge := generateChallenge()
	fmt.Fprintf(conn, "POW: Find a nonce such that sha256( %s + nonce) has %d leading zeros\n", challenge, currentDifficulty)

	var nonce int
	_, err := fmt.Fscanf(conn, "%d\n", &nonce)
	if err != nil {
		fmt.Println("Failed to read nonce:", err)
		return
	}

	if verifyProofOfWork(challenge, nonce, currentDifficulty) {
		quote := quotes[rand.Intn(len(quotes))]
		fmt.Fprintf(conn, "SUCCESS: Here is your quote: %s\n", quote)
	} else {
		fmt.Fprintln(conn, "FAILURE: Proof of work verification failed.")
	}
}

func adjustDifficulty() {
	for {
		time.Sleep(monitorInterval)

		mu.Lock()
		if requestCount > threshold && difficulty < maxDifficulty {
			difficulty++
			log.Info().Msg(fmt.Sprintf("difficulty: %d", difficulty))
		} else if requestCount <= threshold && difficulty > baseDifficulty {
			difficulty--
			log.Info().Msg(fmt.Sprintf("difficulty: %d", difficulty))
		}
		mu.Unlock()
		atomic.StoreInt64(&requestCount, 0)
	}
}

func main() {
	cfg, err := config.GetConfig()
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	quotes = cfg.Quotes
	ctx := regSignalHandler(context.Background())

	if err := runServer(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server stopped on error")
	}
}
