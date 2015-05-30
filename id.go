package main

import (
	"math/rand"
	"strconv"
	"time"
)

// 8位36进制能有2821109907455个可用ID
const idLen = 8

// 用于生成随机数
var r = rand.New(rand.NewSource(time.Now().UnixNano()))

// 生成随机ID
func newRandID() string { return strconv.FormatInt(r.Int63(), 36) }

func strIdAdd(src string, i uint64) string {
	n, err := strconv.ParseUint(src, 36, 64)
	if err != nil {
		panic(err)
	}
	n += i
	result := strconv.FormatUint(n, 36)
	if len(result) < idLen {
		// 补零
		zero := "000000000"
		result = zero[:idLen-len(result)] + result
	}
	return result
}
