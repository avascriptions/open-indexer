package utils

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"github.com/wealdtech/go-merkletree/keccak256"
	"math/big"
	"strconv"
	"strings"
)

func Keccak256(str string) string {
	h := keccak256.New()
	bytes := h.Hash([]byte(str))
	return hex.EncodeToString(bytes)
}

func HexToUint64(hex string) uint64 {
	num, ok := new(big.Int).SetString(hex[2:], 16)
	if ok {
		return num.Uint64()
	} else {
		return 0
	}
}

func HexToUint32(hex string) uint32 {
	return uint32(HexToUint64(hex))
}

func HexStrToBytes(hexStr string) []byte {
	if strings.HasPrefix(hexStr, "0x") {
		hexStr = hexStr[2:]
	}
	data, err := hex.DecodeString(hexStr)
	if err != nil {
		return make([]byte, 0)
	}
	return data
}

func BytesToHexStr(hexBytes []byte) string {
	data := hex.EncodeToString(hexBytes)
	return "0x" + data
}

func BoolToUint32(b bool) uint32 {
	if b {
		return 1
	}
	return 0
}

func ParseInt32(str string) int32 {
	if strings.Contains(str, ".") {
		str = strings.Split(str, ".")[0]
	}
	rst, err := strconv.ParseInt(str, 10, 32)
	if err != nil {
		return 0
	} else {
		return int32(rst)
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

func Uint64ToBytes(value uint64) []byte {
	buffer := new(bytes.Buffer)
	binary.Write(buffer, binary.BigEndian, value)
	return buffer.Bytes()
}

func BytesToUint64(_bytes []byte) uint64 {
	var value uint64
	err := binary.Read(bytes.NewReader(_bytes), binary.BigEndian, &value)
	if err != nil {
		return 0
	}
	return value
}

func TopicToAddress(str string) string {
	return "0x" + str[26:]
}

func TopicToBigInt(str string) *big.Int {
	var start = 0
	for i, c := range str {
		if i > 1 && c != '0' {
			start = i
			break
		}
	}
	if start == 0 {
		return new(big.Int).SetInt64(0)
	}
	hex := str[start:]

	num, ok := new(big.Int).SetString(hex, 16)
	if !ok {
		return new(big.Int).SetInt64(0)
	}
	return num
}
