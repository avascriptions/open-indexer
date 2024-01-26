package model

import "open-indexer/model/serialize"

type Asc20 struct {
	Id        uint64
	Number    uint64
	Tick      string
	From      string
	To        string
	Operation string
	Precision int
	Limit     *DDecimal
	Amount    *DDecimal
	Hash      string
	Block     uint64
	Timestamp uint64
	Valid     int8
}

func (a *Asc20) ToProtoRecord() *serialize.ProtoRecord {
	protoRecord := &serialize.ProtoRecord{
		Id:        a.Id,
		Number:    a.Number,
		Tick:      a.Tick,
		From:      a.From,
		To:        a.To,
		Operation: a.Operation,
		Precision: uint32(a.Precision),
		Limit:     a.Limit.String(),
		Amount:    a.Amount.String(),
		Hash:      a.Hash,
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
		From:      a.From,
		To:        a.To,
		Operation: a.Operation,
		Precision: int(a.Precision),
		Limit:     limit,
		Amount:    amount,
		Hash:      a.Hash,
		Block:     a.Block,
		Timestamp: a.Timestamp,
		Valid:     int8(a.Valid),
	}
	return asc20
}
