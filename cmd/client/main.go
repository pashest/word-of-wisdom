package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

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

	for {
		response, err = reader.ReadString('\n')
		if err != nil {
			log.Error().Msg(fmt.Sprintf("server error: %v", err))
			return
		}
		fmt.Println("msg received:", strings.TrimSpace(response))

		time.Sleep(1 * time.Second)
	}
}
