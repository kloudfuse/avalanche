package metrics

import (
	"math/rand"
	"strconv"
)

type ValueGenerator interface {
	Generate() string
}

type RandomSetValueGenerator struct {
	numValues int
}

func NewRandomSetValueGenerator(numValues int) *RandomSetValueGenerator {
	return &RandomSetValueGenerator{numValues}
}

func (gen *RandomSetValueGenerator) Generate() string {
	r := rand.Intn(gen.numValues)
	return "values_" + strconv.Itoa(r)
}
