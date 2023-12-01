package utils

import (
	"math/big"
	"strconv"
	"strings"
)

func HexToUint64(hex string) uint64 {
	num, ok := new(big.Int).SetString(hex[2:], 16)
	if ok {
		return num.Uint64()
	} else {
		return 0
	}
}

func ParseInt64(str string) int64 {
	if strings.Contains(str, ".") {
		str = strings.Split(str, ".")[0]
	}
	rst, err := strconv.ParseInt(str, 10, 64)
	if err != nil {
		return 0
	} else {
		return rst
	}
}
