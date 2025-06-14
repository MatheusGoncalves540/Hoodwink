package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func GenerateNewId() string {
	rand.NewSource(time.Now().UnixNano())
	randomFactor := rand.Float64() * rand.Float64()
	id := int64(randomFactor * float64(time.Now().UnixNano()))
	return strconv.FormatInt(id, 16)
}
