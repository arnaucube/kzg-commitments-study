package kzg

import (
	"fmt"
	"math/big"

	bn256 "github.com/ethereum/go-ethereum/crypto/bn256/cloudflare"
)

// TrustedSetup also named Reference String
type TrustedSetup struct {
	Tau1 []*bn256.G1
	Tau2 []*bn256.G2
}

// NewTrustedSetup returns a new trusted setup. This step should be done in a
// secure & distributed way
func NewTrustedSetup(p []*big.Int) (*TrustedSetup, error) {
	// compute random s
	s, err := randBigInt()
	if err != nil {
		return nil, err
	}

	// Notation: [x]₁=xG ∈ 𝔾₁, [x]₂=xH ∈ 𝔾₂
	// τ₁: [x₀]₁, [x₁]₁, [x₂]₁, ..., [x n₋₁]₁
	// τ₂: [x₀]₂, [x₁]₂, [x₂]₂, ..., [x n₋₁]₂

	// sPow := make([]*big.Int, len(p))
	tauG1 := make([]*bn256.G1, len(p))
	tauG2 := make([]*bn256.G2, len(p))
	for i := 0; i < len(p); i++ {
		sPow := fExp(s, big.NewInt(int64(i)))
		tauG1[i] = new(bn256.G1).ScalarBaseMult(sPow)
		tauG2[i] = new(bn256.G2).ScalarBaseMult(sPow)
	}

	return &TrustedSetup{tauG1, tauG2}, nil
}

// Commit generates the commitment to the polynomial p(x)
func Commit(ts *TrustedSetup, p []*big.Int) *bn256.G1 {
	c := evaluateG1(ts, p)
	return c
}

func evaluateG1(ts *TrustedSetup, p []*big.Int) *bn256.G1 {
	c := new(bn256.G1).ScalarMult(ts.Tau1[0], p[0])
	for i := 1; i < len(p); i++ {
		sp := new(bn256.G1).ScalarMult(ts.Tau1[i], p[i])
		c = new(bn256.G1).Add(c, sp)
	}
	return c
}

//nolint:deadcode,unused
func evaluateG2(ts *TrustedSetup, p []*big.Int) *bn256.G2 {
	c := new(bn256.G2).ScalarMult(ts.Tau2[0], p[0])
	for i := 1; i < len(p); i++ {
		sp := new(bn256.G2).ScalarMult(ts.Tau2[i], p[i])
		c = new(bn256.G2).Add(c, sp)
	}
	return c
}

// EvaluationProof generates the evaluation proof
func EvaluationProof(ts *TrustedSetup, p []*big.Int, z, y *big.Int) (*bn256.G1, error) {
	n := polynomialSub(p, []*big.Int{y})    // p-y
	d := []*big.Int{fNeg(z), big.NewInt(1)} // x-z
	q, rem := polynomialDiv(n, d)
	if compareBigIntArray(rem, arrayOfZeroes(len(rem))) {
		return nil,
			fmt.Errorf("remainder should be 0, instead is %d", rem)
	}
	fmt.Println("q(x):", polynomialToString(q)) // TMP DBG

	// proof: e = [q(s)]₁
	e := evaluateG1(ts, q)
	return e, nil
}

// Verify computes the KZG commitment verification
func Verify(ts *TrustedSetup, c, proof *bn256.G1, z, y *big.Int) bool {
	s2 := ts.Tau2[1] // [s]₂ = sG ∈ 𝔾₂ = Tau2[1]
	zG2Neg := new(bn256.G2).Neg(
		new(bn256.G2).ScalarBaseMult(z)) // [z]₂ = zG ∈ 𝔾₂
	// [s]₂ - [z]₂
	sz := new(bn256.G2).Add(s2, zG2Neg)

	yG1Neg := new(bn256.G1).Neg(
		new(bn256.G1).ScalarBaseMult(y)) // [y]₁ = yG ∈ 𝔾₁
	// c - [y]₁
	cy := new(bn256.G1).Add(c, yG1Neg)

	h := new(bn256.G2).ScalarBaseMult(big.NewInt(1)) // H ∈ 𝔾₂

	// e(proof, [s]₂ - [z]₂) == e(c - [y]₁, H)
	e1 := bn256.Pair(proof, sz)
	e2 := bn256.Pair(cy, h)
	return e1.String() == e2.String()
}
