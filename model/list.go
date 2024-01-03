package model

type List struct {
	InsId     string
	Owner     string
	Exchange  string
	Tick      string
	Amount    *DDecimal
	Precision int
}
