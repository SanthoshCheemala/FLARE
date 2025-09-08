# The Mathematics of Lattice-Based Cryptography in Your Implementation

This document explains the fundamental concepts behind your lattice-based cryptography implementation, focusing on Ring-LWE, noise, and how encryption/decryption works.

## 1. Lattices and Rings

### What is a Lattice?
A lattice is a regular arrangement of points in n-dimensional space. Mathematically, it's the set of all integer linear combinations of some basis vectors.

### What is a Ring?
In your code, a "ring" refers to a polynomial ring, specifically:

```
R = ℤ_Q[X]/(X^D + 1)
```

Where:
- **ℤ_Q**: Integers modulo Q (180143985094819841 in your case)
- **D**: Polynomial degree (4096 in your case)
- **X^D + 1**: The polynomial used to define the quotient ring (a cyclotomic polynomial)

In your code: `leParams.R` represents this mathematical structure.

## 2. Ring-LWE (Learning With Errors)

Ring-LWE is a computational problem that forms the basis of your cryptosystem:

### Problem Statement
Given samples of the form (a, b = a·s + e), where:
- **a** is a random polynomial in the ring
- **s** is a secret polynomial
- **e** is a "small" error/noise polynomial
- All operations are in the polynomial ring

The task of recovering s from these samples is believed to be computationally hard, forming the security basis of your system.

## 3. Noise Generation and Parameters

In your implementation:
```go
e0[j] = matrix.NewNoiseVec(le.M, le.R, le.PRNG, le.Sigma, le.Bound).NTT(le.R)
```

This creates noise with:
- **Sigma**: Standard deviation of a Gaussian distribution (controls noise magnitude)
- **Bound**: Maximum absolute value for noise coefficients
- **Noise Distribution**: Typically Gaussian or bounded uniform

These parameters ensure:
1. Noise is small enough for correct decryption
2. Noise is large enough for security

## 4. Key Generation

Your code generates:
```go
pkTests[i], skTests[i] = le.KeyGen()
```

Mathematically:
1. **Secret key (sk)**: A random "small" polynomial s
2. **Public key (pk)**: A pair (a, b = a·s + e) where:
   - a is random in the ring
   - e is small noise
   - s is the secret key

## 5. Encryption

In `Laconic_PSI_server`:
```go
c0, c1, c, d := LE.Enc(le, pp, idSet[i], msg, r, e0, e1, e)
```

This is performing:
1. Sample random r
2. Add carefully calibrated noise (e0, e1, e)
3. Create ciphertext components that encode the message while hiding it with noise

For a message m, simplified encryption looks like:
- c0 = a·r + e0
- c1 = b·r + e1 + ⌊q/2⌋·m

where ⌊q/2⌋·m scales the binary message to approximately q/2 for 1s and 0 for 0s.

## 6. Decryption

```go
m := LE.Dec(le, skTests[i], vec1, vec2, c0, c1, c, d)
```

Mathematically:
1. **Compute**: result = c1 - s·c0
2. **This gives**: result ≈ ⌊q/2⌋·m + noise
3. **Decode**: If coefficient is close to 0, it's 0; if close to q/2, it's 1

This works because:
```
c1 - s·c0 = (b·r + e1 + ⌊q/2⌋·m) - s·(a·r + e0)
          = (a·s + e)·r + e1 + ⌊q/2⌋·m - s·a·r - s·e0
          = e·r + e1 - s·e0 + ⌊q/2⌋·m
          ≈ ⌊q/2⌋·m + (small noise)
```

## 7. Noise Impact on Decryption

The `CorrectnessCheck` function handles noise in decryption:
```go
if decrypted.Coeffs[0][i] < q14 || decrypted.Coeffs[0][i] > q34 {
    binaryDecrypted.Coeffs[0][i] = 0
} else {
    binaryDecrypted.Coeffs[0][i] = 1
}
```

This checks if each coefficient is closer to 0 or q/2 (accounting for wraparound at q).

## 8. Number Theoretic Transform (NTT)

You're using NTT operations frequently:
```go
pp := LE.ReadFromDB(db,0,0,leParams).NTT(leParams.R)
```

NTT is the finite field equivalent of the Fast Fourier Transform, allowing efficient polynomial multiplication:
- **Converting to NTT domain**: O(n log n) complexity
- **Multiplication in NTT domain**: Simple coefficient-wise multiplication O(n)
- **Converting back**: O(n log n) using inverse NTT

## 9. Laconic Private Set Intersection

Your code implements PSI using lattice cryptography:

1. Clients and servers hash their data to create element identifiers
2. Elements are encrypted using the lattice scheme
3. Due to the homomorphic properties, when elements match, the noise remains manageable
4. When elements don't match, noise becomes too large for proper decryption
5. The comparison function detects matches by checking if noise is within acceptable bounds

## 10. Parameter Selection Trade-offs

Your parameters:
- **Q = 180143985094819841** (≈ 2^58): Large modulus allows more noise before decryption fails
- **D = 4096**: Higher dimension means better security and more noise tolerance
- **N = 4**: Matrix dimension for the scheme
- **Layers = 50**: Tree depth for the laconic structure

These represent trade-offs between:
- **Security** (larger parameters = more security)
- **Performance** (larger parameters = slower operations)
- **Correctness** (larger parameters = more noise tolerance)

## Conclusion

The Laconic PSI approach you're implementing uses these lattice techniques to allow private set intersection with communication complexity independent of the server's set size, making it highly efficient for asymmetric set sizes.

---

*These notes provide a mathematical foundation for understanding the lattice-based cryptographic operations in your implementation.*