package rnd

import (
	"fmt"
	"math/rand"
)

func GenerateRandomNumber() string {
	return fmt.Sprintf("%d", rand.Intn(900000)+100000) // Generates a random number between 0 and 100,000 (inclusive)
}
