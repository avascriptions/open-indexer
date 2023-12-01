package model

type Asc20 struct {
	Number    uint64
	Tick      string
	Operation string
	Precision int
	Max       *DDecimal
	Limit     *DDecimal
	Amount    *DDecimal
	Valid     int8
}
