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

	PopUpType string
)

const (
	InfoMsg PopUpType = "info"
	ErrMsg  PopUpType = "error"
	WarnMsg PopUpType = "warn"
)

const (
	// server to client
	MsgGameState = "game_state"
	MsgTableList = "table_list"
	MsgPopUp     = "pop_up"
	MsgUserStats = "user_stats"

	// client to server
	MsgPlaceBet    = "place_bet"
	MsgHit         = "hit"
	MsgStand       = "stand"
	MsgJoinTable   = "join_table"
	MsgLeaveTable  = "leave_table"
	MsgCreateTable = "create_table"
	MsgDeleteTable = "delete_table"
	MsgStartGame   = "start_game"
	MsgDealCards   = "deal_cards"
	MsgGetState    = "get_state"
	MsgGetStats    = "get_stats"

	MsgLogin      = "login"
	MsgAuthStatus = "auth_status"
)

type ValueMessage struct {
	Value string `json:"value"`
}

func PackageMessage(dto any) (*TransportMessage, error) {
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
	case PopUpDTO:
		message.Type = MsgPopUp
	case StatsDTO:
		message.Type = MsgUserStats
	}

	return &message, nil
}

func PackageClientMessage(typ, val string) *TransportMessage {
	message := TransportMessage{}
	if val != "" {
		data, err := json.Marshal(ValueMessage{Value: val})
		if err != nil {
			// eventually we can make an error message?
			return &TransportMessage{}
		}
		message.Data = data
	}
	message.Type = typ
	return &message
}
