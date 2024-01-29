package model

import (
	"open-indexer/model/serialize"
	"open-indexer/utils"
)

type Token struct {
	Tick        string    `json:"tick"`
	Number      uint64    `json:"number"`
	Precision   int       `json:"precision"`
	Max         *DDecimal `json:"max"`
	Limit       *DDecimal `json:"limit"`
	Minted      *DDecimal `json:"minted"`
	Progress    uint32    `json:"progress"`
	Holders     uint32    `json:"holders"`
	Trxs        uint32    `json:"trxs"`
	CreatedAt   uint64    `json:"created_at"`
	CompletedAt uint64    `json:"completed_at"`
	Hash        string    `json:"hash"`
}

func (t *Token) ToProtoToken() *serialize.ProtoToken {
	protoToken := &serialize.ProtoToken{
		Tick:        t.Tick,
		Number:      t.Number,
		Precision:   uint32(t.Precision),
		Max:         t.Max.String(),
		Limit:       t.Limit.String(),
		Minted:      t.Minted.String(),
		Progress:    t.Progress,
		Holders:     t.Holders,
		Trxs:        t.Trxs,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
		Hash:        utils.HexStrToBytes(t.Hash),
	}
	return protoToken
}

func TokenFromProto(t *serialize.ProtoToken) *Token {
	max, _, _ := NewDecimalFromString(t.Max)
	limit, _, _ := NewDecimalFromString(t.Limit)
	minted, _, _ := NewDecimalFromString(t.Minted)
	token := &Token{
		Tick:        t.Tick,
		Number:      t.Number,
		Precision:   int(t.Precision),
		Max:         max,
		Limit:       limit,
		Minted:      minted,
		Progress:    t.Progress,
		Holders:     t.Holders,
		Trxs:        t.Trxs,
		CreatedAt:   t.CreatedAt,
		CompletedAt: t.CompletedAt,
		Hash:        utils.BytesToHexStr(t.Hash)[2:],
	}
	return token
}
