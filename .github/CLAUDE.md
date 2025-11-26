# CLAUDE.md

## Project Overview

This is Trail of Bits' pure Go implementation of **ML-DSA (FIPS-204)**, a NIST-standardized post-quantum digital signature algorithm. The implementation has been specifically engineered to be **constant-time** to prevent timing side-channel attacks.

## Critical Security Property: Constant-Time Execution

**All code in this repository MUST execute in constant time with respect to secret data.** This means the execution time must not vary based on the values of secret keys, signatures during generation, or any intermediate values derived from secrets.

Violating constant-time properties can enable attacks like [KyberSlash](https://kyberslash.cr.yp.to/), which exploited timing variations in division operations to recover secret keys.

---

## Cryptocoding Rules

The following rules are adapted from the [cryptocoding guidelines](https://github.com/veorq/cryptocoding) and apply to all code in this repository.

### Rule 1: Compare Secret Strings in Constant Time

String comparisons performed byte-per-byte may be exploited in timing attacks, for example to forge MACs or signatures.

**❌ FORBIDDEN:**
```go
// Built-in comparisons are NOT constant-time
if bytes.Equal(secretA, secretB) { ... }
if string(secretA) == string(secretB) { ... }
if hmacA == hmacB { ... }
```

**✅ REQUIRED:**
```go
import "crypto/subtle"

// Use constant-time comparison for all secret data
if subtle.ConstantTimeCompare(secretA, secretB) == 1 { ... }
```

For fixed-size comparisons, you can also use XOR accumulation:
```go
func constantTimeEqual(a, b []byte) int {
    if len(a) != len(b) {
        return 0
    }
    var result byte
    for i := 0; i < len(a); i++ {
        result |= a[i] ^ b[i]
    }
    // Returns 1 if equal, 0 otherwise
    return int(1 & ((uint32(result) - 1) >> 8))
}
```

### Rule 2: Avoid Branches Controlled by Secret Data

If a conditional branch (`if`, `switch`, `for`, `while`) depends on secret data, then the code executed and its execution time depend on the secret data as well.

**❌ FORBIDDEN:**
```go
// Branch prediction leaks timing information
if secretValue > threshold {
    doSomething()
} else {
    doSomethingElse()
}

// Switch on secret data
switch secretByte {
case 0: ...
case 1: ...
}
```

**✅ REQUIRED - Constant-Time Selection:**
```go
// Compute BOTH branches, then select using bit masking
resultA := computeA()
resultB := computeB()

// Create mask: 0xFFFFFFFF if condition is true, 0x00000000 otherwise
// condition must be 0 or 1
mask := uint32(-condition)
result := resultA ^ ((resultA ^ resultB) & mask)
```

**Constant-time comparison primitives for 32-bit values:**
```go
// Return 1 if x != 0, 0 otherwise
func ctIsNonZero(x uint32) uint32 {
    return (x | -x) >> 31
}

// Return 1 if x == 0, 0 otherwise
func ctIsZero(x uint32) uint32 {
    return 1 ^ ctIsNonZero(x)
}

// Return 1 if x != y, 0 otherwise
func ctNeq(x, y uint32) uint32 {
    return ((x - y) | (y - x)) >> 31
}

// Return 1 if x == y, 0 otherwise
func ctEq(x, y uint32) uint32 {
    return 1 ^ ctNeq(x, y)
}

// Return 1 if x < y (unsigned), 0 otherwise
func ctLt(x, y uint32) uint32 {
    return (x ^ ((x ^ y) | ((x - y) ^ y))) >> 31
}

// Return 1 if x > y (unsigned), 0 otherwise
func ctGt(x, y uint32) uint32 {
    return ctLt(y, x)
}

// Generate mask: 0xFFFFFFFF if bit != 0, 0 otherwise
func ctMask(bit uint32) uint32 {
    return uint32(-int32(ctIsNonZero(bit)))
}

// Select x if bit != 0, y otherwise (constant-time ternary)
func ctSelect(x, y, bit uint32) uint32 {
    m := ctMask(bit)
    return (x & m) | (y & ^m)
}
```

**Using `math/bits` for constant-time comparisons:**
```go
import "math/bits"

// bits.Sub32 returns (difference, borrow)
// borrow = 1 if a < b, borrow = 0 if a >= b
_, borrow := bits.Sub32(a, b, 0)
```

### Rule 3: Avoid Table Look-Ups Indexed by Secret Data

The access time of a table element can vary with its index due to CPU cache effects. This has been exploited in cache-timing attacks on AES.

**❌ FORBIDDEN:**
```go
// Cache timing attacks can recover secretIndex
value := lookupTable[secretIndex]
```

**✅ REQUIRED:**
Access ALL table entries and select the correct one using constant-time masking:
```go
func constantTimeLookup(table []uint32, secretIndex int) uint32 {
    var result uint32
    for i := 0; i < len(table); i++ {
        // mask is 0xFFFFFFFF when i == secretIndex, 0 otherwise
        mask := ctMask(ctEq(uint32(i), uint32(secretIndex)))
        result |= table[i] & mask
    }
    return result
}
```

Alternatively, use bitsliced implementations where table lookups are replaced with sequences of constant-time logical operations.

### Rule 4: Avoid Secret-Dependent Loop Bounds

Loops with bounds derived from secret values directly expose timing information.

**❌ FORBIDDEN:**
```go
// Loop count reveals information about secretLength
for i := 0; i < secretLength; i++ {
    process(data[i])
}

// Finding the first set bit reveals bit position
for i := 31; i >= 0; i-- {
    if (secret >> i) & 1 == 1 {
        break
    }
}
```

**✅ REQUIRED:**
```go
// Always iterate the maximum number of times
for i := 0; i < MAX_LENGTH; i++ {
    // Use constant-time selection to conditionally process
    shouldProcess := ctLt(uint32(i), uint32(secretLength))
    // ... process with masking based on shouldProcess
}
```

### Rule 5: Avoid Division and Modulo on Secret Data

Hardware division instructions take variable time depending on operand values. This was the root cause of the KyberSlash attack.

**❌ FORBIDDEN:**
```go
// Division timing depends on operand values
quotient := secretValue / divisor
remainder := secretValue % divisor
```

**✅ REQUIRED - Barrett Reduction for Known Divisors:**

For ML-DSA, division is always by 2γ₂, which is a known constant per parameter set:
- ML-DSA-44: γ₂ = 95232, so 2γ₂ = 190464
- ML-DSA-65/87: γ₂ = 261888, so 2γ₂ = 523776

Use Barrett reduction with precomputed reciprocals:
```go
import "math/bits"

func DivBarrett(numerator, denominator uint32) (quotient, remainder uint32) {
    var reciprocal uint64
    switch denominator {
    case 190464: // 2 * 95232 (ML-DSA-44)
        reciprocal = 96851604889688
    case 523776: // 2 * 261888 (ML-DSA-65/87)
        reciprocal = 35184372088832
    default:
        return DivConstTime32(numerator, denominator)
    }
    
    // Barrett reduction
    hi, _ := bits.Mul64(uint64(numerator), reciprocal)
    quo := uint32(hi)
    r := numerator - quo*denominator
    
    // Constant-time correction steps
    for i := 0; i < 2; i++ {
        newR, borrow := bits.Sub32(r, denominator, 0)
        correction := borrow ^ 1
        mask := uint32(-correction)
        quo += mask & 1
        r ^= mask & (newR ^ r)
    }
    return quo, r
}
```

**✅ REQUIRED - General Constant-Time Division (slow fallback):**
```go
func DivConstTime32(n, d uint32) (quotient, remainder uint32) {
    var q, r uint32
    for i := 31; i >= 0; i-- {
        r <<= 1
        r |= (n >> i) & 1
        
        rPrime, borrow := bits.Sub32(r, d, 0)
        swap := borrow ^ 1
        
        qPrime := q | (1 << i)
        
        mask := uint32(-swap)
        r ^= (rPrime ^ r) & mask
        q ^= (qPrime ^ q) & mask
    }
    return q, r
}
```

### Rule 6: Prevent Compiler Interference with Security-Critical Operations

Compilers may optimize out operations they deem useless, including security-critical memory clearing. The Go compiler is generally better about this than C compilers, but caution is still warranted.

**Concerns:**
- Memory clearing may be optimized away if the variable is not used afterward
- Compilers may introduce branches into "branchless" code
- Dead code elimination may remove security checks

**Mitigations:**
- Check generated assembly for critical functions
- Use `//go:noinline` directive to prevent inlining of security-critical functions
- Use `runtime.KeepAlive()` to prevent premature garbage collection
- Consider using `crypto/subtle` functions which are designed to resist optimization

```go
//go:noinline
func secureZero(b []byte) {
    for i := range b {
        b[i] = 0
    }
    runtime.KeepAlive(b)
}
```

### Rule 7: Clean Memory of Secret Data

Secret data should be cleared from memory as soon as it's no longer needed to minimize the window of exposure.

**Note:** Go's garbage collector makes this challenging. Unlike C, you cannot guarantee when memory will be reclaimed or that it won't be copied.

**Best Effort in Go:**
```go
func clearSecret(secret []byte) {
    for i := range secret {
        secret[i] = 0
    }
    runtime.KeepAlive(secret)
}

// Use defer to ensure cleanup
func processSecret(secret []byte) {
    defer clearSecret(secret)
    // ... use secret
}
```

**Caution:** 
- Go strings are immutable; never store secrets in strings
- Slices may be copied by the runtime during garbage collection
- Consider using fixed-size arrays on the stack for short-lived secrets

### Rule 8: Use Strong Randomness

Cryptographic systems require high-quality random numbers. Weak randomness has led to catastrophic failures (e.g., Debian OpenSSL bug).

**❌ FORBIDDEN:**
```go
import "math/rand"
// NEVER use math/rand for cryptographic purposes
randomValue := rand.Int()
```

**✅ REQUIRED:**
```go
import "crypto/rand"

// Always use crypto/rand for cryptographic randomness
randomBytes := make([]byte, 32)
if _, err := rand.Read(randomBytes); err != nil {
    // Handle error - DO NOT continue with weak randomness
    panic("crypto/rand failed")
}
```

**Additional Requirements:**
- Always check return values from the RNG
- Never seed a PRNG with predictable values (timestamps, PIDs, etc.)
- Do not implement custom PRNGs
- Do not reuse randomness across different operations

### Rule 9: Avoid Early Returns Based on Secret Data

Early returns create different execution paths with different timing characteristics.

**❌ FORBIDDEN:**
```go
func verify(secret, input []byte) bool {
    for i := 0; i < len(secret); i++ {
        if secret[i] != input[i] {
            return false  // Early exit leaks position of first difference
        }
    }
    return true
}
```

**✅ REQUIRED:**
```go
func verify(secret, input []byte) bool {
    if len(secret) != len(input) {
        return false
    }
    var result byte
    for i := 0; i < len(secret); i++ {
        result |= secret[i] ^ input[i]
    }
    return result == 0
}
```

### Rule 10: Use Unsigned Types for Binary Data

Signed integer operations can have undefined or implementation-defined behavior, particularly with bit shifts.

**❌ PROBLEMATIC:**
```go
// Signed shift behavior may vary
var signed int8 = -1
shifted := signed >> 4  // Implementation-defined in some contexts
```

**✅ REQUIRED:**
```go
// Use unsigned types for all binary/cryptographic data
var unsigned uint8 = 0xFF
shifted := unsigned >> 4  // Well-defined: 0x0F
```

### Rule 11: Always Typecast Shifted Values

When combining bytes into larger integers, ensure proper typecasting to avoid undefined behavior and sign extension issues.

**❌ PROBLEMATIC:**
```go
// Without explicit cast, bytes may be promoted incorrectly
func combine(b0, b1, b2, b3 byte) uint32 {
    return (b0 << 24) | (b1 << 16) | (b2 << 8) | b3
}
```

**✅ REQUIRED:**
```go
func combine(b0, b1, b2, b3 byte) uint32 {
    return (uint32(b0) << 24) | (uint32(b1) << 16) | (uint32(b2) << 8) | uint32(b3)
}
```

### Rule 12: Prevent API Confusion Between Secure and Insecure Functions

When APIs have both secure and insecure variants, make the distinction obvious and make insecure usage difficult.

**Guidelines:**
- Name insecure functions explicitly (e.g., `UnsafeCompare` not `FastCompare`)
- Internal helper functions that bypass security checks should be clearly marked
- Document security requirements for each public function

---

## ML-DSA Specific Considerations

### Parameter Sets

| Parameter Set | γ₂ | 2γ₂ | Barrett Reciprocal (2⁶⁴/2γ₂) |
|--------------|------|--------|------------------------------|
| ML-DSA-44 | 95232 | 190464 | 96851604889688 |
| ML-DSA-65 | 261888 | 523776 | 35184372088832 |
| ML-DSA-87 | 261888 | 523776 | 35184372088832 |

### Key Functions Requiring Constant-Time Implementation

- `Decompose` - Converts field elements; requires constant-time division
- `HighBits` / `LowBits` - Extract components using `Decompose`
- `MakeHint` / `UseHint` - Hint computation for signatures
- All polynomial arithmetic operating on secret coefficients
- Signature generation (involves rejection sampling with secret data)

### The Decompose Algorithm

The `Decompose` function is particularly sensitive because it requires division by 2γ₂. The naive implementation:

```go
// ❌ NOT CONSTANT-TIME - DO NOT USE
func DecomposeUnsafe(r, alpha int32) (r1, r0 int32) {
    r = r % q
    if r < 0 { r += q }
    if r > (q-1)/2 { r = r - q }
    // ... branches and division on secret data
}
```

Must be replaced with Barrett reduction and branchless conditionals as shown in Rule 5.

---

## Code Quality Requirements

### Linting

**All code MUST pass `golangci-lint` before being merged.** The linter runs automatically in CI on every push and pull request.

To run the linter locally:
```bash
golangci-lint run ./...
```

Common issues to avoid:
- **errcheck**: Always handle or explicitly ignore error returns (use `_ = funcThatReturnsError()` if intentionally ignoring)
- **unused**: Remove unused variables, functions, and imports
- **ineffassign**: Don't assign to variables that are never used afterward
- **staticcheck**: Follow Go best practices flagged by staticcheck

The linter configuration is in `.golangci.yml`. Do not disable linter rules without discussion.

---

## Code Review Checklist

Before approving any changes, verify:

### Code Quality

- [ ] `golangci-lint run ./...` passes with no errors
- [ ] No new linter warnings introduced
- [ ] Error returns are handled or explicitly ignored

### Timing Side-Channels
- [ ] No `/` or `%` operators on secret-dependent values
- [ ] No `if`/`else`/`switch` branching on secret-dependent values  
- [ ] No early `return` based on secret-dependent conditions
- [ ] No secret-dependent array/slice indexing
- [ ] No variable-length loops where iteration count depends on secrets
- [ ] No use of `bytes.Equal`, `bytes.Compare`, `==` on secrets
- [ ] All comparisons use `subtle.ConstantTimeCompare` or bit operations

### Implementation Quality
- [ ] Uses `crypto/rand` for all randomness (never `math/rand`)
- [ ] RNG return values are checked
- [ ] Secret data is zeroed when no longer needed
- [ ] Unsigned types used for all binary/cryptographic data
- [ ] Explicit typecasts on all bit shift operations
- [ ] No string type used for secret data

### Code Patterns
- [ ] All conditional assignments use constant-time selection (bit masking)
- [ ] Barrett reduction used for division by known constants
- [ ] `DivConstTime32` used for any other required division
- [ ] Assembly output reviewed for critical functions (if feasible)

---

## References

- [FIPS 204 (ML-DSA) Specification](https://csrc.nist.gov/pubs/fips/204/final)
- [KyberSlash Attack](https://kyberslash.cr.yp.to/)
- [Trail of Bits Blog: Avoiding Side-Channels in Post-Quantum Go Libraries](https://blog.trailofbits.com/2025/11/14/how-we-avoided-side-channels-in-our-new-post-quantum-go-cryptography-libraries/)
- [Cryptocoding Guidelines](https://github.com/veorq/cryptocoding)
- [BearSSL Constant-Time Documentation](https://www.bearssl.org/constanttime.html)
- [Go crypto/subtle Package](https://pkg.go.dev/crypto/subtle)
