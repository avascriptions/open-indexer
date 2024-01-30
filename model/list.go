package model

import (
	"open-indexer/model/serialize"
	"open-indexer/utils"
)

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
		InsId:     utils.HexStrToBytes(l.InsId),
		Owner:     utils.HexStrToBytes(l.Owner),
		Exchange:  utils.HexStrToBytes(l.Exchange),
		Tick:      l.Tick,
		Amount:    l.Amount.String(),
		Precision: uint32(l.Precision),
	}
	return protoRecord
}

func ListFromProto(l *serialize.ProtoList) *List {
	amount, _, _ := NewDecimalFromString(l.Amount)
	asc20 := &List{
		InsId:     utils.BytesToHexStr(l.InsId),
		Owner:     utils.BytesToHexStr(l.Owner),
		Exchange:  utils.BytesToHexStr(l.Exchange),
		Tick:      l.Tick,
		Amount:    amount,
		Precision: int(l.Precision),
	}
	return asc20
}
