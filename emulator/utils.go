package emulator

import "math/rand/v2"

func randomByte() byte {
	return byte(rand.IntN(255))
}
