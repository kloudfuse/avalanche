package metrics

import "math/rand"

func GetRandomCountInRange(minSamples int, maxSamples int) int {
	return (minSamples + rand.Intn(maxSamples-minSamples+1))
}
