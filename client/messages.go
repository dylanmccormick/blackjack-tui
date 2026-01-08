package client

import (
	"bytes"
	"encoding/json"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/dylanmccormick/blackjack-tui/protocol"
)

func ReceiveMessage(sub <-chan *protocol.TransportMessage) tea.Cmd {
	return func() tea.Msg {
		log.Println("Reading message from chan")
		msg := <-sub
		log.Println(msg)
		switch msg.Type {
		case protocol.MsgGameState:
			body := &protocol.GameDTO{}
			err := json.Unmarshal(msg.Data, body)
			if err != nil {
				log.Printf("ERROR: %v", err)
			}
			return body
		case protocol.MsgTableList:
			body := []*protocol.TableDTO{}
			err := json.Unmarshal(msg.Data, &body)
			if err != nil {
				log.Printf("ERROR: %v", err)
			}
			return body
		}

		return nil
	}
}

func ParseTransportMessage(msg []byte) []*protocol.TransportMessage {
	// Parses a transport message. Returns the type and the packaged message
	log.Println("Parsing data from transport message")
	messages := []*protocol.TransportMessage{}
	decoder := json.NewDecoder(bytes.NewReader(msg))
	for decoder.More() {
		var tm protocol.TransportMessage
		err := decoder.Decode(&tm)
		if err != nil {
			log.Println(err)
			panic(err)
		}
		messages = append(messages, &tm)
	}
	log.Println("returning messages", messages)
	return messages
}
