package utils

import (
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func GenerateShortID() string {
	n := uint64(time.Now().UnixNano())
	var res []byte
	for n > 0 {
		res = append(res, charset[n%62])
		n /= 62
	}
	return string(res[:6])
}
