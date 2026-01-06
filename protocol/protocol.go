package protocol

import (
	"encoding/json"
)

type (
	// This is a wrapper for all message types between the server and the client
	TransportMessage struct {
		Type string          `json:"type"`
		Data json.RawMessage `json:"data,omitempty"`
	}
)

const (
	// server to client
	MsgGameState = "game_state"
	MsgTableList = "table_list"
	// client to server
	MsgPlaceBet    = "place_bet"
	MsgHit         = "hit"
	MsgStand       = "stand"
	MsgJoinTable   = "join_table"
	MsgLeaveTable  = "leave_table"
	MsgCreateTable = "create_table"
	MsgStartGame   = "start_game"
	MsgDealCards   = "deal_cards"
)

type ValueMessage struct {
	Value string `json:"value"`
}

func PackageMessage(dto any) ([]byte, error) {
	message := TransportMessage{}
	data, err := json.Marshal(dto)
	if err != nil {
		return nil, err
	}
	message.Data = data
	switch dto.(type) {
	case GameDTO:
		message.Type = MsgGameState
	case []TableDTO:
		message.Type = MsgTableList
	}

	packagedMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return packagedMessage, nil
}

func PackageClientMessage(typ, val string) ([]byte, error) {
	message := TransportMessage{}
	if val != "" {
		data, err := json.Marshal(ValueMessage{Value: val})
		if err != nil {
			return nil, err
		}
		message.Data = data
	}
	message.Type = typ
	packagedMessage, err := json.Marshal(message)
	if err != nil {
		return nil, err
	}
	return packagedMessage, nil
}
