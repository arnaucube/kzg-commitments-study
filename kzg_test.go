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
	assert.Equal(t, "x³ + x¹ + 5", PolynomialToString(p))

	// TrustedSetup
	ts, err := NewTrustedSetup(len(p))
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

func TestBatchProof(t *testing.T) {
	// p(x) = x^3 + x + 5
	p := []*big.Int{
		big.NewInt(5),
		big.NewInt(1),  // x^1
		big.NewInt(0),  // x^2
		big.NewInt(1),  // x^3
		big.NewInt(10), // x^4
	}
	assert.Equal(t, "10x⁴ + x³ + x¹ + 5", PolynomialToString(p))

	// TrustedSetup
	ts, err := NewTrustedSetup(len(p))
	assert.Nil(t, err)

	// Commit
	c := Commit(ts, p)

	// 1st point: p(z)=y --> p(3)=35
	z0 := big.NewInt(3)
	y0 := polynomialEval(p, z0)

	// 2nd point: p(10)=1015
	z1 := big.NewInt(10)
	y1 := polynomialEval(p, z1)

	// 3nd point: p(256)=16777477
	z2 := big.NewInt(256)
	y2 := polynomialEval(p, z2)

	zs := []*big.Int{z0, z1, z2}
	ys := []*big.Int{y0, y1, y2}

	// prove an evaluation of the multiple z_i & y_i
	proof, err := EvaluationBatchProof(ts, p, zs, ys)
	assert.Nil(t, err)

	// batch proof verification
	v := VerifyBatchProof(ts, c, proof, zs, ys)
	assert.True(t, v)

	// changing order of the points to be verified
	zs[0], zs[1], zs[2] = zs[1], zs[2], zs[0]
	ys[0], ys[1], ys[2] = ys[1], ys[2], ys[0]
	v = VerifyBatchProof(ts, c, proof, zs, ys)
	assert.True(t, v)

	// change a value of zs and check that verification fails
	zs[0] = big.NewInt(2)
	v = VerifyBatchProof(ts, c, proof, zs, ys)
	assert.False(t, v)

	// using a value that is not in the evaluation proof should generate a
	// proof that will not correctly be verified
	zs = []*big.Int{z0, z1, z2}
	ys = []*big.Int{y0, y1, y2}
	proof, err = EvaluationBatchProof(ts, p, zs, ys)
	assert.Nil(t, err)
	zs[2] = big.NewInt(2500)
	ys[2] = polynomialEval(p, zs[2])
	v = VerifyBatchProof(ts, c, proof, zs, ys)
	assert.False(t, v)
}
