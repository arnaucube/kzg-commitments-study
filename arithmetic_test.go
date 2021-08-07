package kzg

import (
	"bytes"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func randBI() *big.Int {
	maxbits := 256
	b := make([]byte, (maxbits/8)-1)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	r := new(big.Int).SetBytes(b)
	return new(big.Int).Mod(r, Q)
}

func neg(a *big.Int) *big.Int {
	return new(big.Int).Neg(a)
}

func TestPolynomial(t *testing.T) {
	b0 := big.NewInt(int64(0))
	b1 := big.NewInt(int64(1))
	b2 := big.NewInt(int64(2))
	b3 := big.NewInt(int64(3))
	b4 := big.NewInt(int64(4))
	b5 := big.NewInt(int64(5))
	b6 := big.NewInt(int64(6))
	b16 := big.NewInt(int64(16))

	a := []*big.Int{b1, b0, b5}
	b := []*big.Int{b3, b0, b1}

	// new Finite Field
	r, ok := new(big.Int).SetString("21888242871839275222246405745257275088548364400416034343698204186575808495617", 10) //nolint:lll
	assert.True(nil, ok)

	// polynomial multiplication
	o := polynomialMul(a, b)
	assert.Equal(t, o, []*big.Int{b3, b0, b16, b0, b5})

	// polynomial division
	quo, rem := polynomialDiv(a, b)
	assert.Equal(t, quo[0].Int64(), int64(5))
	// check the rem result without modulo
	assert.Equal(t, new(big.Int).Sub(rem[0], r).Int64(), int64(-14))

	c := []*big.Int{neg(b4), b0, neg(b2), b1}
	d := []*big.Int{neg(b3), b1}
	quo2, rem2 := polynomialDiv(c, d)
	assert.Equal(t, quo2, []*big.Int{b3, b1, b1})
	assert.Equal(t, rem2[0].Int64(), int64(5))

	// polynomial addition
	o = polynomialAdd(a, b)
	assert.Equal(t, o, []*big.Int{b4, b0, b6})

	// polynomial subtraction
	o1 := polynomialSub(a, b)
	o2 := polynomialSub(b, a)
	o = polynomialAdd(o1, o2)
	assert.True(t, bytes.Equal(b0.Bytes(), o[0].Bytes()))
	assert.True(t, bytes.Equal(b0.Bytes(), o[1].Bytes()))
	assert.True(t, bytes.Equal(b0.Bytes(), o[2].Bytes()))

	c = []*big.Int{b5, b6, b1}
	d = []*big.Int{b1, b3}
	o = polynomialSub(c, d)
	assert.Equal(t, o, []*big.Int{b4, b3, b1})

	// NewPolZeroAt
	o = newPolZeroAt(3, 4, b4)
	assert.Equal(t, polynomialEval(o, big.NewInt(3)), b4)
	o = newPolZeroAt(2, 4, b3)
	assert.Equal(t, polynomialEval(o, big.NewInt(2)), b3)

	// polynomialEval
	// p(x) = x^3 + x + 5
	p := []*big.Int{
		big.NewInt(5),
		big.NewInt(1), // x^1
		big.NewInt(0), // x^2
		big.NewInt(1), // x^3
	}
	assert.Equal(t, "x³ + x¹ + 5", PolynomialToString(p))
	assert.Equal(t, "35", polynomialEval(p, big.NewInt(3)).String())
	assert.Equal(t, "1015", polynomialEval(p, big.NewInt(10)).String())
	assert.Equal(t, "16777477", polynomialEval(p, big.NewInt(256)).String())
	assert.Equal(t, "125055", polynomialEval(p, big.NewInt(50)).String())
	assert.Equal(t, "7", polynomialEval(p, big.NewInt(1)).String())
}

func BenchmarkArithmetic(b *testing.B) {
	// generate arrays with bigint
	var p, q []*big.Int
	for i := 0; i < 1000; i++ {
		pi := randBI()
		p = append(p, pi)
	}
	for i := 1000 - 1; i >= 0; i-- {
		q = append(q, p[i])
	}

	b.Run("polynomialSub", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polynomialSub(p, q)
		}
	})
	b.Run("polynomialMul", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polynomialMul(p, q)
		}
	})
	b.Run("polynomialDiv", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			polynomialDiv(p, q)
		}
	})
}

func TestLagrangeInterpolation(t *testing.T) {
	x0 := big.NewInt(3)
	y0 := big.NewInt(35)
	x1 := big.NewInt(10)
	y1 := big.NewInt(1015)
	x2 := big.NewInt(256)
	y2 := big.NewInt(16777477)
	x3 := big.NewInt(50)
	y3 := big.NewInt(125055)

	xs := []*big.Int{x0, x1, x2, x3}
	ys := []*big.Int{y0, y1, y2, y3}

	p, err := LagrangeInterpolation(xs, ys)
	assert.Nil(t, err)
	assert.Equal(t, "x³ + x¹ + 5", PolynomialToString(p))

	assert.Equal(t, y0, polynomialEval(p, x0))
	assert.Equal(t, y1, polynomialEval(p, x1))
	assert.Equal(t, y2, polynomialEval(p, x2))
}

func TestZeroPolynomial(t *testing.T) {
	x0 := big.NewInt(1)
	x1 := big.NewInt(40)
	x2 := big.NewInt(512)
	xs := []*big.Int{x0, x1, x2}

	z := zeroPolynomial(xs)
	assert.Equal(t, "x³ "+
		"+ 21888242871839275222246405745257275088548364400416034343698204186575808495064x² "+
		"+ 21032x¹ + 21888242871839275222246405745257275088548364400416034343698204186575808475137",
		PolynomialToString(z))

	assert.Equal(t, "0", polynomialEval(z, x0).String())
	assert.Equal(t, "0", polynomialEval(z, x1).String())
	assert.Equal(t, "0", polynomialEval(z, x2).String())
}
