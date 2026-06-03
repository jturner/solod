// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import "solod.dev/so/math/bits"

// uint128 represents a uint128 using two uint64s.
//
// When the methods below mention a bit number, bit 0 is the most
// significant bit (in hi) and bit 127 is the lowest (lo&1).
//
//so:extern so_uint128
type uint128 struct {
	hi uint64
	lo uint64
}

// mask6 returns a uint128 bitmask with the topmost n bits of a
// 128-bit number.
func mask6(n int) uint128 {
	// n must be in the range [0, 128]. Boundary cases are handled
	// explicitly to avoid C undefined behavior from shifts >= 64.
	if n == 128 {
		return uint128{^uint64(0), ^uint64(0)}
	}
	if n > 64 {
		return uint128{^uint64(0), ^uint64(0) << (128 - n)}
	}
	if n == 64 {
		return uint128{^uint64(0), 0}
	}
	if n == 0 {
		return uint128{0, 0}
	}
	return uint128{^(^uint64(0) >> n), 0}
}

// isZero reports whether u == 0.
func (u uint128) isZero() bool { return u.hi|u.lo == 0 }

// and returns the bitwise AND of u and m (u&m).
func (u uint128) and(m uint128) uint128 {
	return uint128{u.hi & m.hi, u.lo & m.lo}
}

// xor returns the bitwise XOR of u and m (u^m).
func (u uint128) xor(m uint128) uint128 {
	return uint128{u.hi ^ m.hi, u.lo ^ m.lo}
}

// subOne returns u - 1.
func (u uint128) subOne() uint128 {
	lo, borrow := bits.Sub64(u.lo, 1, 0)
	return uint128{u.hi - borrow, lo}
}

// addOne returns u + 1.
func (u uint128) addOne() uint128 {
	lo, carry := bits.Add64(u.lo, 1, 0)
	return uint128{u.hi + carry, lo}
}

// equal reports whether u == u2.
func (u uint128) equal(u2 uint128) bool {
	return u.hi == u2.hi && u.lo == u2.lo
}
