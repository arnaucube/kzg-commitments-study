package kzg

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleFlow(t *testing.T) {
	// p(x) = x^3 + x + 5
	p := []*big.Int{
		big.NewInt(5),
		big.NewInt(1), // x^1
		big.NewInt(0), // x^2
		big.NewInt(1), // x^3
	}
	assert.Equal(t, "1x³ + 1x¹ + 5", PolynomialToString(p))

	// TrustedSetup
	ts, err := NewTrustedSetup(p)
	assert.Nil(t, err)

	// Commit
	c := Commit(ts, p)

	// p(z)=y --> p(3)=35
	z := big.NewInt(3)
	y := big.NewInt(35)

	// z & y: to prove an evaluation p(z)=y
	proof, err := EvaluationProof(ts, p, z, y)
	assert.Nil(t, err)

	v := Verify(ts, c, proof, z, y)
	assert.True(t, v)

	v = Verify(ts, c, proof, big.NewInt(4), y)
	assert.False(t, v)
}
