package kzg

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"math/big"
	"strconv"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// Q is the order of the integer field (Zq) that fits inside the snark
var Q, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088696311157297823662689037894645226208583", 10)

// R is the mod of the finite field
var R, _ = new(big.Int).SetString(
	"21888242871839275222246405745257275088548364400416034343698204186575808495617", 10)

func randBigInt() (*big.Int, error) {
	maxbits := R.BitLen()
	b := make([]byte, (maxbits/8)-1)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	r := new(big.Int).SetBytes(b)
	rq := new(big.Int).Mod(r, R)

	return rq, nil
}

func arrayOfZeroes(n int) []*big.Int {
	r := make([]*big.Int, n)
	for i := 0; i < n; i++ {
		r[i] = new(big.Int).SetInt64(0)
	}
	return r[:]
}

//nolint:deadcode,unused
func arrayOfZeroesG1(n int) []*bn256.G1 {
	r := make([]*bn256.G1, n)
	for i := 0; i < n; i++ {
		r[i] = new(bn256.G1).ScalarBaseMult(big.NewInt(0))
	}
	return r[:]
}

//nolint:deadcode,unused
func arrayOfZeroesG2(n int) []*bn256.G2 {
	r := make([]*bn256.G2, n)
	for i := 0; i < n; i++ {
		r[i] = new(bn256.G2).ScalarBaseMult(big.NewInt(0))
	}
	return r[:]
}

func compareBigIntArray(a, b []*big.Int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := 0; i < len(a); i++ {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func fAdd(a, b *big.Int) *big.Int {
	ab := new(big.Int).Add(a, b)
	return ab.Mod(ab, R)
}

func fSub(a, b *big.Int) *big.Int {
	ab := new(big.Int).Sub(a, b)
	return new(big.Int).Mod(ab, R)
}

func fMul(a, b *big.Int) *big.Int {
	ab := new(big.Int).Mul(a, b)
	return ab.Mod(ab, R)
}

func fDiv(a, b *big.Int) *big.Int {
	ab := new(big.Int).Mul(a, new(big.Int).ModInverse(b, R))
	return new(big.Int).Mod(ab, R)
}

//nolint:unused,deadcode // TODO check
func fNeg(a *big.Int) *big.Int {
	return new(big.Int).Mod(new(big.Int).Neg(a), R)
}

//nolint:deadcode,unused
func fInv(a *big.Int) *big.Int {
	return new(big.Int).ModInverse(a, R)
}

func fExp(base *big.Int, e *big.Int) *big.Int {
	res := big.NewInt(1)
	rem := new(big.Int).Set(e)
	exp := base

	for !bytes.Equal(rem.Bytes(), big.NewInt(int64(0)).Bytes()) {
		// if BigIsOdd(rem) {
		if rem.Bit(0) == 1 { // .Bit(0) returns 1 when is odd
			res = fMul(res, exp)
		}
		exp = fMul(exp, exp)
		rem.Rsh(rem, 1)
	}
	return res
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func polynomialAdd(a, b []*big.Int) []*big.Int {
	r := arrayOfZeroes(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = fAdd(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		r[i] = fAdd(r[i], b[i])
	}
	return r
}

func polynomialSub(a, b []*big.Int) []*big.Int {
	r := arrayOfZeroes(max(len(a), len(b)))
	for i := 0; i < len(a); i++ {
		r[i] = fAdd(r[i], a[i])
	}
	for i := 0; i < len(b); i++ {
		r[i] = fSub(r[i], b[i])
	}
	return r
}

func polynomialMul(a, b []*big.Int) []*big.Int {
	r := arrayOfZeroes(len(a) + len(b) - 1)
	for i := 0; i < len(a); i++ {
		for j := 0; j < len(b); j++ {
			r[i+j] = fAdd(r[i+j], fMul(a[i], b[j]))
		}
	}
	return r
}

func polynomialDiv(a, b []*big.Int) ([]*big.Int, []*big.Int) {
	// https://en.wikipedia.org/wiki/Division_algorithm
	r := arrayOfZeroes(len(a) - len(b) + 1)
	rem := a
	for len(rem) >= len(b) {
		l := fDiv(rem[len(rem)-1], b[len(b)-1])
		pos := len(rem) - len(b)
		r[pos] = l
		aux := arrayOfZeroes(pos)
		aux1 := append(aux, l)
		aux2 := polynomialSub(rem, polynomialMul(b, aux1))
		rem = aux2[:len(aux2)-1]
	}
	return r, rem
}

// polynomialEval evaluates the polinomial over the Finite Field at the given value x
func polynomialEval(p []*big.Int, x *big.Int) *big.Int {
	r := big.NewInt(int64(0))
	for i := 0; i < len(p); i++ {
		xi := fExp(x, big.NewInt(int64(i)))
		elem := fMul(p[i], xi)
		r = fAdd(r, elem)
	}
	return r
}

// newPolZeroAt generates a new polynomial that has value zero at the given value
func newPolZeroAt(pointPos, totalPoints int, height *big.Int) []*big.Int {
	fac := 1
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			fac = fac * (pointPos - i)
		}
	}
	facBig := big.NewInt(int64(fac))
	hf := fDiv(height, facBig)
	r := []*big.Int{hf}
	for i := 1; i < totalPoints+1; i++ {
		if i != pointPos {
			ineg := big.NewInt(int64(-i))
			b1 := big.NewInt(int64(1))
			r = polynomialMul(r, []*big.Int{ineg, b1})
		}
	}
	return r
}

var sNums = map[string]string{
	"0": "⁰",
	"1": "¹",
	"2": "²",
	"3": "³",
	"4": "⁴",
	"5": "⁵",
	"6": "⁶",
	"7": "⁷",
	"8": "⁸",
	"9": "⁹",
}

func intToSNum(n int) string {
	s := strconv.Itoa(n)
	sN := ""
	for i := 0; i < len(s); i++ {
		sN += sNums[string(s[i])]
	}
	return sN
}

func polynomialToString(p []*big.Int) string {
	s := ""
	for i := len(p) - 1; i >= 1; i-- {
		if !bytes.Equal(p[i].Bytes(), big.NewInt(0).Bytes()) {
			s += fmt.Sprintf("%sx%s + ", p[i], intToSNum(i))
		}
	}
	s += p[0].String()
	return s
}

// TODO add method to 'clean' the polynomial, to remove right-zeroes
