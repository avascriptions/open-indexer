package decimal

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"strings"
)

const MAX_PRECISION = 18

var precisionFactor = new(big.Int).Exp(big.NewInt(10), big.NewInt(MAX_PRECISION), nil)

// Decimal represents a fixed-point decimal number with 18 decimal places
type Decimal struct {
	value *big.Int
}

func New() *Decimal {
	return &Decimal{value: new(big.Int).SetUint64(0)}
}

func NewCopy(other *Decimal) *Decimal {
	return &Decimal{value: new(big.Int).Set(other.value)}
}

// NewFromString creates a Decimal instance from a string
func NewFromString(s string) (*Decimal, int, error) {
	if s == "" {
		return nil, 0, errors.New("empty string")
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return nil, 0, fmt.Errorf("invalid decimal format: %s", s)
	}

	integerPartStr := parts[0]
	if integerPartStr == "" || integerPartStr[0] == '+' {
		return nil, 0, errors.New("empty integer")
	}

	integerPart, ok := new(big.Int).SetString(parts[0], 10)
	if !ok {
		return nil, 0, fmt.Errorf("invalid integer format: %s", parts[0])
	}

	currPrecision := 0
	decimalPart := big.NewInt(0)
	if len(parts) == 2 {
		decimalPartStr := parts[1]
		if decimalPartStr == "" || decimalPartStr[0] == '-' || decimalPartStr[0] == '+' {
			return nil, 0, errors.New("empty decimal")
		}

		currPrecision = len(decimalPartStr)
		if currPrecision > MAX_PRECISION {
			return nil, 0, fmt.Errorf("decimal exceeds maximum precision: %s", s)
		}
		n := MAX_PRECISION - currPrecision
		for i := 0; i < n; i++ {
			decimalPartStr += "0"
		}
		decimalPart, ok = new(big.Int).SetString(decimalPartStr, 10)
		if !ok || decimalPart.Sign() < 0 {
			return nil, 0, fmt.Errorf("invalid decimal format: %s", parts[0])
		}
	}

	value := new(big.Int).Mul(integerPart, precisionFactor)
	if value.Sign() < 0 {
		value = value.Sub(value, decimalPart)
	} else {
		value = value.Add(value, decimalPart)
	}

	return &Decimal{value: value}, currPrecision, nil
}

// String returns the string representation of a Decimal instance
func (d *Decimal) String() string {
	if d == nil {
		return "0"
	}
	value := new(big.Int).Abs(d.value)
	quotient, remainder := new(big.Int).QuoRem(value, precisionFactor, new(big.Int))
	sign := ""
	if d.value.Sign() < 0 {
		sign = "-"
	}
	if remainder.Sign() == 0 {
		return fmt.Sprintf("%s%s", sign, quotient.String())
	}
	decimalPart := fmt.Sprintf("%0*d", MAX_PRECISION, remainder)
	decimalPart = strings.TrimRight(decimalPart, "0")
	return fmt.Sprintf("%s%s.%s", sign, quotient.String(), decimalPart)
}

func (d *Decimal) GetValue() *big.Int {
	return d.value
}

func (d *Decimal) Set(other *Decimal) *Decimal {
	d.value.Set(other.GetValue())
	return d
}

// Add adds two Decimal instances and returns a new Decimal instance
func (d *Decimal) Add(other *Decimal) *Decimal {
	if d == nil && other == nil {
		value := new(big.Int).SetUint64(0)
		return &Decimal{value: value}
	}
	if other == nil {
		value := new(big.Int).Set(d.value)
		return &Decimal{value: value}
	}
	if d == nil {
		value := new(big.Int).Set(other.value)
		return &Decimal{value: value}
	}
	value := new(big.Int).Add(d.value, other.value)
	return &Decimal{value: value}
}

// Sub subtracts two Decimal instances and returns a new Decimal instance
func (d *Decimal) Sub(other *Decimal) *Decimal {
	if d == nil && other == nil {
		value := new(big.Int).SetUint64(0)
		return &Decimal{value: value}
	}
	if other == nil {
		value := new(big.Int).Set(d.value)
		return &Decimal{value: value}
	}
	if d == nil {
		value := new(big.Int).Neg(other.value)
		return &Decimal{value: value}
	}
	value := new(big.Int).Sub(d.value, other.value)
	return &Decimal{value: value}
}

func (d *Decimal) Cmp(other *Decimal) int {
	if d == nil && other == nil {
		return 0
	}
	if other == nil {
		return d.value.Sign()
	}
	if d == nil {
		return -other.value.Sign()
	}
	return d.value.Cmp(other.value)
}

func (d *Decimal) Sign() int {
	if d == nil {
		return 0
	}
	return d.value.Sign()
}

func (d *Decimal) IsOverflowUint64() bool {
	if d == nil {
		return false
	}

	integerPart := new(big.Int).SetUint64(math.MaxUint64)
	value := new(big.Int).Mul(integerPart, precisionFactor)
	if d.value.Cmp(value) > 0 {
		return true
	}
	return false
}

func (d *Decimal) Float64() float64 {
	if d == nil {
		return 0
	}
	value := new(big.Int).Abs(d.value)
	quotient, remainder := new(big.Int).QuoRem(value, precisionFactor, new(big.Int))
	f := float64(quotient.Uint64()) + float64(remainder.Uint64())/math.MaxFloat64
	if d.value.Sign() < 0 {
		return -f
	}
	return f
}
