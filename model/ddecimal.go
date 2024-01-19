package model

import (
	"database/sql/driver"
	"open-indexer/utils/decimal"
)

type DDecimal struct {
	value *decimal.Decimal
}

func NewDecimal() *DDecimal {
	return &DDecimal{decimal.New()}
}

func NewDecimalFromString(s string) (*DDecimal, int, error) {
	d, p, e := decimal.NewFromString(s)

	return &DDecimal{d}, p, e
}

func (dd *DDecimal) Add(other *DDecimal) *DDecimal {
	d := dd.value.Add(other.value)
	return &DDecimal{d}
}

func (dd *DDecimal) Sub(other *DDecimal) *DDecimal {
	d := dd.value.Sub(other.value)
	return &DDecimal{d}
}

func (dd *DDecimal) Cmp(other *DDecimal) int {
	return dd.value.Cmp(other.value)
}

func (dd *DDecimal) Sign() int {
	return dd.value.Sign()
}

func (dd *DDecimal) String() string {
	if dd == nil {
		return "0"
	}
	return dd.value.String()
}

func (dd *DDecimal) Scan(value interface{}) error {
	str := string(value.([]byte))
	d, _, err := decimal.NewFromString(str)
	dd.value = d
	return err
}

func (dd *DDecimal) Value() (driver.Value, error) {
	if dd == nil {
		return "0", nil
	}
	return dd.value.String(), nil
}

func (dd *DDecimal) MarshalJSON() ([]byte, error) {
	return []byte(dd.String()), nil
}

func (dd *DDecimal) UnmarshalJSON(data []byte) error {
	d, _, e := decimal.NewFromString(string(data))
	dd.value = d
	return e
}
