package model

type Token struct {
	Tick        string
	Number      uint64
	Precision   int
	Max         *DDecimal
	Limit       *DDecimal
	Minted      *DDecimal
	Progress    int32
	Holders     int32
	Trxs        int32
	CreatedAt   uint64
	CompletedAt int64
}
