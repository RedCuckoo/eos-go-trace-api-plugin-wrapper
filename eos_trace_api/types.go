package eos_trace_api

import (
	"github.com/eoscanada/eos-go"
	"github.com/eoscanada/eos-go/ecc"
)

type BlockHeader struct {
	Timestamp        eos.BlockTimestamp `json:"timestamp"`
	Producer         eos.AccountName    `json:"producer"`
	Status           string             `json:"status"`
	Previous         eos.Checksum256    `json:"previous_id"`
	TransactionMRoot eos.Checksum256    `json:"transaction_mroot"`
	ActionMRoot      eos.Checksum256    `json:"action_mroot"`
	ScheduleVersion  uint32             `json:"schedule_version"`
}

type SignedTransactionHeader struct {
	Status               eos.TransactionStatus `json:"status"`
	CPUUsageMicroSeconds uint32                `json:"cpu_usage_us"`
	NetUsageWords        eos.Varuint32         `json:"net_usage_words"`
}

type PermissionLevel struct {
	Actor      eos.AccountName    `json:"account"`
	Permission eos.PermissionName `json:"permission"`
}

type Action struct {
	Account        eos.AccountName   `json:"account"`
	Name           eos.ActionName    `json:"action"`
	Receiver       eos.AccountName   `json:"receiver"`
	GlobalSequence eos.Uint64           `json:"global_sequence,omitempty"`
	Authorization  []PermissionLevel `json:"authorization,omitempty"`
	eos.ActionData
}

type Transaction struct {
	eos.TransactionHeader

	ID eos.Checksum256 `json:"id"`
	Actions []*Action `json:"actions"`
}

type SignedTransaction struct {
	SignedTransactionHeader
	*Transaction

	Signatures []ecc.Signature `json:"signatures"`
}

type Block struct {
	BlockHeader
	Transactions []SignedTransaction `json:"transactions"`
}

type BlockResp struct {
	Block
	ID       eos.Checksum256 `json:"id"`
	BlockNum uint32          `json:"number"`
}