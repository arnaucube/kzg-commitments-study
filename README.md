# kzg-commitments-study [![GoDoc](https://godoc.org/github.com/arnaucube/kzg-commitments-study?status.svg)](https://godoc.org/github.com/arnaucube/kzg-commitments-study) [![Go Report Card](https://goreportcard.com/badge/github.com/arnaucube/kzg-commitments-study)](https://goreportcard.com/report/github.com/arnaucube/kzg-commitments-study) [![Test](https://github.com/arnaucube/kzg-commitments-study/workflows/Test/badge.svg)](https://github.com/arnaucube/kzg-commitments-study/actions?query=workflow%3ATest)

Doing this to study and learn [KZG commitments](http://cacr.uwaterloo.ca/techreports/2010/cacr2010-10.pdf), do not use in production. More details at https://arnaucube.com/blog/kzg-commitments.html .

Thanks to [Dankrad Feist](https://dankradfeist.de/ethereum/2020/06/16/kate-polynomial-commitments.html), [Alin Tomescu](https://alinush.github.io/2020/05/06/kzg-polynomial-commitments.html), [Tom Walton-Pocock](https://hackmd.io/@tompocock/Hk2A7BD6U) for their articles, which helped me understand a bit the KZG Commitments.

It uses the [ethereum bn256](https://github.com/ethereum/go-ethereum/tree/master/crypto/bn256/cloudflare).

### Usage

```go
// p(x) = x^3 + x + 5
p := []*big.Int{
	big.NewInt(5),
	big.NewInt(1), // x^1
	big.NewInt(0), // x^2
	big.NewInt(1), // x^3
}
assert.Equal(t, "1x³ + 1x¹ + 5", PolynomialToString(p))

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

// verification
v := Verify(ts, c, proof, z, y)
assert.True(t, v)
```

Batch Proofs:
```go
// zs & ys contain the f(z_i)=y_i values that will be proved inside a batch proof
zs := []*big.Int{z0, z1, z2}
ys := []*big.Int{y0, y1, y2}

// prove an evaluation of the multiple z_i & y_i
proof, err := EvaluationBatchProof(ts, p, zs, ys)
assert.Nil(t, err)

// batch proof verification
v := VerifyBatchProof(ts, c, proof, zs, ys)
assert.True(t, v)
```
