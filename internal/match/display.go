package match

import (
	"crypto/rand"
	"math/big"
)

const (
	syntheticCountMin = 3
	syntheticCountMax = 47
)

func SyntheticDisplayCount() int64 {
	n, err := rand.Int(rand.Reader, big.NewInt(syntheticCountMax-syntheticCountMin+1))
	if err != nil {
		return syntheticCountMin
	}
	return n.Int64() + syntheticCountMin
}
