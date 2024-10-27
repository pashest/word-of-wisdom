package main

import (
	"bufio"
	"crypto/sha256"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

func findNonce(challenge string, difficulty int) int {
	for nonce := 0; ; nonce++ {
		data := fmt.Sprintf("%s%d", challenge, nonce)
		hash := sha256.Sum256([]byte(data))
		hashString := fmt.Sprintf("%x", hash)
		if strings.HasPrefix(hashString, strings.Repeat("0", difficulty)) {
			return nonce
		}
	}
}

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		log.Fatal().Err(err)
		return
	}
	defer conn.Close()
	log.Info().Msg("client connected")

	reader := bufio.NewReader(conn)
	var response string

	response, err = reader.ReadString('\n')
	fmt.Println("Received challenge:", response)

	parts := strings.Split(response, " ")
	challenge := parts[7]
	difficulty, _ := strconv.Atoi(parts[11])

	now := time.Now()
	nonce := findNonce(challenge, difficulty)
	fmt.Printf("Found nonce: %d, spent %d millisecconds\n", nonce, time.Since(now).Milliseconds())

	fmt.Fprintf(conn, "%d\n", nonce)

	response, _ = bufio.NewReader(conn).ReadString('\n')
	fmt.Println("Server response:", response)
}
