package model

import "open-indexer/model/serialize"

type List struct {
	InsId     string
	Owner     string
	Exchange  string
	Tick      string
	Amount    *DDecimal
	Precision int
}

func (l *List) ToProtoList() *serialize.ProtoList {
	protoRecord := &serialize.ProtoList{
		InsId:     l.InsId,
		Owner:     l.Owner,
		Exchange:  l.Exchange,
		Tick:      l.Tick,
		Amount:    l.Amount.String(),
		Precision: uint32(l.Precision),
	}
	return protoRecord
}

func ListFromProto(l *serialize.ProtoList) *List {
	amount, _, _ := NewDecimalFromString(l.Amount)
	asc20 := &List{
		InsId:     l.InsId,
		Owner:     l.Owner,
		Exchange:  l.Exchange,
		Tick:      l.Tick,
		Amount:    amount,
		Precision: int(l.Precision),
	}
	return asc20
}
