package model

import (
	"open-indexer/model/serialize"
	"open-indexer/utils"
)

type Asc20 struct {
	Id        uint64    `json:"id"`
	Number    uint64    `json:"number"`
	Tick      string    `json:"tick"`
	From      string    `json:"from"`
	To        string    `json:"to"`
	Operation string    `json:"operation"`
	Precision int       `json:"precision"`
	Limit     *DDecimal `json:"limit"`
	Amount    *DDecimal `json:"amount"`
	Hash      string    `json:"hash"`
	Block     uint64    `json:"block"`
	Timestamp uint64    `json:"timestamp"`
	Valid     int8      `json:"valid"`
}

func (a *Asc20) ToProtoRecord() *serialize.ProtoRecord {
	protoRecord := &serialize.ProtoRecord{
		Id:        a.Id,
		Number:    a.Number,
		Tick:      a.Tick,
		From:      utils.HexStrToBytes(a.From),
		To:        utils.HexStrToBytes(a.To),
		Operation: a.Operation,
		Precision: uint32(a.Precision),
		Limit:     a.Limit.String(),
		Amount:    a.Amount.String(),
		Hash:      utils.HexStrToBytes(a.Hash),
		Block:     a.Block,
		Timestamp: a.Timestamp,
		Valid:     int32(a.Valid),
	}
	return protoRecord
}

func Asc20FromProto(a *serialize.ProtoRecord) *Asc20 {
	limit, _, _ := NewDecimalFromString(a.Limit)
	amount, _, _ := NewDecimalFromString(a.Amount)
	asc20 := &Asc20{
		Id:        a.Id,
		Number:    a.Number,
		Tick:      a.Tick,
		From:      utils.BytesToHexStr(a.From),
		To:        utils.BytesToHexStr(a.To),
		Operation: a.Operation,
		Precision: int(a.Precision),
		Limit:     limit,
		Amount:    amount,
		Hash:      utils.BytesToHexStr(a.Hash)[2:],
		Block:     a.Block,
		Timestamp: a.Timestamp,
		Valid:     int8(a.Valid),
	}
	return asc20
}
