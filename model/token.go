package model

type Token struct {
	Tick        string    `json:"tick"`
	Number      uint64    `json:"number"`
	Precision   int       `json:"precision"`
	Max         *DDecimal `json:"max"`
	Limit       *DDecimal `json:"limit"`
	Minted      *DDecimal `json:"minted"`
	Progress    int32     `json:"progress"`
	Holders     int32     `json:"holders"`
	Trxs        int32     `json:"trxs"`
	CreatedAt   uint64    `json:"created_at"`
	CompletedAt uint64    `json:"completed_at"`
	Hash        string    `json:"hash"`
	Updated     bool      `json:"-"`
}
