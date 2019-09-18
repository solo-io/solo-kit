package helpers

import (
	"fmt"
	"math/rand"
	"time"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz1234567890")

func RandString(length int) string {
	b := make([]rune, length)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func RandStringGinkgo(length int, ginkgoSeed int64, node int) string {
	seedStr := RandString(length)
	return fmt.Sprintf("%s-%d-%d", seedStr, ginkgoSeed, node)
}
