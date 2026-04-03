package utils

import (
	"time"
)

func GenerateShortID(key string) string {
	n := uint64(time.Now().UnixNano())
	var res []byte
	for n > 0 {
		res = append(res, key[n%62])
		n /= 62
	}
	return string(res[:6])
}
