package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/pashest/word-of-wisdom/internal/model"
	"github.com/pashest/word-of-wisdom/internal/pkg/pow/equihash"
	"github.com/pashest/word-of-wisdom/internal/pkg/utils"
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

	err = sendMsg(model.Message{
		Type: model.RequestChallenge,
	}, conn)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err send request: %v", err))
		return
	}

	var response string
	response, err = reader.ReadString('\n')
	msg, err := utils.ParseMessage(response)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err parse msg: %v", err))
		return
	}
	var challenge model.Challenge
	err = json.Unmarshal([]byte(msg.Payload), &challenge)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err parse challenge: %v", err))
		return
	}
	log.Info().Msg("Received challenge")

	eq := equihash.NewEquihash(challenge.Difficulty, challenge.Input)
	proof := eq.FindProof()

	proofBytes, err := json.Marshal(proof)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err marshal proof: %v", err))
		return
	}

	err = sendMsg(model.Message{
		Type:      model.RequestResource,
		RequestID: msg.RequestID,
		Payload:   string(proofBytes),
	}, conn)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err send resourse request: %v", err))
	}
	log.Info().Msg("Proof sent to server")

	response, err = reader.ReadString('\n')
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err parse msg: %v", err))
		return
	}
	msg, err = utils.ParseMessage(response)
	if err != nil {
		log.Error().Msg(fmt.Sprintf("err parse msg: %v", err))
		return
	}
	if msg.Type == model.SuccessResponseResource {
		log.Info().Msg(msg.Payload)
	} else {
		log.Error().Msg(msg.Payload)
	}

}

func sendMsg(msg model.Message, conn io.Writer) error {
	msgStr := fmt.Sprintf("%s\n", msg.Stringify())
	_, err := conn.Write([]byte(msgStr))
	return err
}
