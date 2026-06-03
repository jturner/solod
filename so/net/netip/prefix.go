// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package netip

import (
	"solod.dev/so/bytealg"
	"solod.dev/so/cmp"
	"solod.dev/so/strconv"
)

// Maximum length of a prefix string.
const (
	MaxPrefix4Len = MaxAddr4Len + 1 + 2 // ip/32
	MaxPrefix6Len = MaxAddr6Len + 1 + 3 // ip/128
	MaxPrefixLen  = MaxPrefix6Len
)

// Prefix is an IP address prefix (CIDR) representing an IP network.
//
// The first [Prefix.Bits]() of [Addr]() are specified. The remaining bits match any address.
// The range of Bits() is [0,32] for IPv4 or [0,128] for IPv6.
type Prefix struct {
	ip Addr

	// bitsPlusOne stores the prefix bit length plus one.
	// A Prefix is valid if and only if bitsPlusOne is non-zero.
	bitsPlusOne uint8
}

// PrefixFrom returns a [Prefix] with the provided IP address and bit
// prefix length.
//
// It does not allocate. Unlike [Addr.Prefix], [PrefixFrom] does not mask
// off the host bits of ip.
//
// If bits is less than zero or greater than ip.BitLen, [Prefix.Bits]
// will return an invalid value -1.
func PrefixFrom(ip Addr, bits int) Prefix {
	var bitsPlusOne uint8
	if !ip.isZero() && bits >= 0 && bits <= ip.BitLen() {
		bitsPlusOne = uint8(bits) + 1
	}
	return Prefix{
		ip:          ip.withoutZone(),
		bitsPlusOne: bitsPlusOne,
	}
}

// Addr returns p's IP address.
func (p Prefix) Addr() Addr { return p.ip }

// Bits returns p's prefix length.
//
// It reports -1 if invalid.
func (p Prefix) Bits() int { return int(p.bitsPlusOne) - 1 }

// IsValid reports whether p.Bits() has a valid range for p.Addr().
// If p.Addr() is the zero [Addr], IsValid returns false.
// Note that if p is the zero [Prefix], then p.IsValid() == false.
func (p Prefix) IsValid() bool { return p.bitsPlusOne > 0 }

func (p Prefix) isZero() bool {
	var zero Prefix
	return p.Equal(zero)
}

// IsSingleIP reports whether p contains exactly one IP.
func (p Prefix) IsSingleIP() bool { return p.IsValid() && p.Bits() == p.ip.BitLen() }

// Equal reports whether p and p2 are the same prefix.
func (p Prefix) Equal(p2 Prefix) bool {
	return p.ip.Equal(p2.ip) && p.bitsPlusOne == p2.bitsPlusOne
}

// Compare returns an integer comparing two prefixes.
// The result will be 0 if p == p2, -1 if p < p2, and +1 if p > p2.
// Prefixes sort first by validity (invalid before valid), then
// address family (IPv4 before IPv6), then masked prefix address, then
// prefix length, then unmasked address.
func (p Prefix) Compare(p2 Prefix) int {
	// Aside from sorting based on the masked address, this use of
	// Addr.Compare also enforces the valid vs. invalid and address
	// family ordering for the prefix.
	if c := p.Masked().Addr().Compare(p2.Masked().Addr()); c != 0 {
		return c
	}

	bits1 := p.Bits()
	bits2 := p2.Bits()
	if c := cmp.Compare(bits1, bits2); c != 0 {
		return c
	}

	return p.Addr().Compare(p2.Addr())
}

// ParsePrefix parses s as an IP address prefix.
// The string can be in the form "192.168.1.0/24" or "2001:db8::/32",
// the CIDR notation defined in RFC 4632 and RFC 4291.
// IPv6 zones are not permitted in prefixes, and an error will be returned if a
// zone is present.
//
// Note that masked address bits are not zeroed. Use Masked for that.
func ParsePrefix(s string) (Prefix, error) {
	i := bytealg.LastIndexByteString(s, '/')
	if i < 0 {
		return Prefix{}, ErrPrefix
	}
	ip, err := ParseAddr(s[:i])
	if err != nil {
		return Prefix{}, ErrPrefix
	}
	// IPv6 zones are not allowed: https://go.dev/issue/51899
	if ip.Is6() && ip.scopeID != 0 {
		return Prefix{}, ErrPrefix
	}

	bitsStr := s[i+1:]

	// strconv.Atoi accepts a leading sign and leading zeroes, but we don't want that.
	if len(bitsStr) > 1 && (bitsStr[0] < '1' || bitsStr[0] > '9') {
		return Prefix{}, ErrPrefix
	}

	bits, err := strconv.Atoi(bitsStr)
	if err != nil {
		return Prefix{}, ErrPrefix
	}
	maxBits := 32
	if ip.Is6() {
		maxBits = 128
	}
	if bits < 0 || bits > maxBits {
		return Prefix{}, ErrPrefix
	}
	return PrefixFrom(ip, bits), nil
}

// MustParsePrefix calls [ParsePrefix](s) and panics on error.
// It is intended for use in tests with hard-coded strings.
func MustParsePrefix(s string) Prefix {
	ip, err := ParsePrefix(s)
	if err != nil {
		panic(err)
	}
	return ip
}

// Masked returns p in its canonical form, with all but the high
// p.Bits() bits of p.Addr() masked off.
//
// If p is zero or otherwise invalid, Masked returns the zero [Prefix].
func (p Prefix) Masked() Prefix {
	m, _ := p.ip.Prefix(p.Bits())
	return m
}

// Contains reports whether the network p includes ip.
//
// An IPv4 address will not match an IPv6 prefix.
// An IPv4-mapped IPv6 address will not match an IPv4 prefix.
// A zero-value IP will not match any prefix.
// If ip has an IPv6 zone, Contains returns false,
// because Prefixes strip zones.
func (p Prefix) Contains(ip Addr) bool {
	if !p.IsValid() || ip.hasZone() {
		return false
	}
	if f1, f2 := p.ip.BitLen(), ip.BitLen(); f1 == 0 || f2 == 0 || f1 != f2 {
		return false
	}
	if ip.Is4() {
		// xor the IP addresses together; mismatched bits are now ones.
		// Shift away the number of bits we don't care about.
		// Shifts in Go are more efficient if the compiler can prove
		// that the shift amount is smaller than the width of the shifted type (64 here).
		// We know that p.bits is in the range 0..32 because p is Valid;
		// the compiler doesn't know that, so mask with 63 to help it.
		// Now truncate to 32 bits, because this is IPv4.
		// If all the bits we care about are equal, the result will be zero.
		return uint32((ip.addr.lo^p.ip.addr.lo)>>((32-p.Bits())&63)) == 0
	} else {
		// xor the IP addresses together.
		// Mask away the bits we don't care about.
		// If all the bits we care about are equal, the result will be zero.
		return ip.addr.xor(p.ip.addr).and(mask6(p.Bits())).isZero()
	}
}

// Overlaps reports whether p and o contain any IP addresses in common.
//
// If p and o are of different address families or either have a zero
// IP, it reports false. Like the Contains method, a prefix with an
// IPv4-mapped IPv6 address is still treated as an IPv6 mask.
func (p Prefix) Overlaps(o Prefix) bool {
	if !p.IsValid() || !o.IsValid() {
		return false
	}
	if p.Equal(o) {
		return true
	}
	if p.ip.Is4() != o.ip.Is4() {
		return false
	}
	var minBits int
	if pb, ob := p.Bits(), o.Bits(); pb < ob {
		minBits = pb
	} else {
		minBits = ob
	}
	if minBits == 0 {
		return true
	}
	// One of these Prefix calls might look redundant, but we don't require
	// that p and o values are normalized (via Prefix.Masked) first,
	// so the Prefix call on the one that's already minBits serves to zero
	// out any remaining bits in IP.
	var err error
	if p, err = p.ip.Prefix(minBits); err != nil {
		return false
	}
	if o, err = o.ip.Prefix(minBits); err != nil {
		return false
	}
	return p.ip.Equal(o.ip)
}

// AppendText implements the [encoding.TextAppender] interface.
// Requires at least [MaxPrefixLen] bytes of spare capacity in b.
func (p Prefix) AppendText(b []byte) ([]byte, error) {
	return p.appendTo(b), nil
}

// String returns the CIDR notation of p: "<ip>/<bits>".
// buf length must be at least [MaxPrefixLen].
func (p Prefix) String(buf []byte) string {
	if !p.IsValid() {
		return "invalid Prefix"
	}
	ip := p.ip.String(buf)
	n := len(ip)
	buf[n] = '/'
	bits := strconv.Itoa(buf[n+1:], p.Bits())
	return string(buf[:n+len(bits)+1])
}

// appendTo appends a text encoding of p
// to b and returns the extended buffer.
func (p Prefix) appendTo(b []byte) []byte {
	if p.isZero() {
		return b
	}
	if !p.IsValid() {
		return append(b, "invalid Prefix"...)
	}

	// p.ip is non-nil, because p is valid.
	if p.ip.bitlen == z4 {
		b = p.ip.appendTo4(b)
	} else {
		if p.ip.Is4In6() {
			b = append(b, "::ffff:"...)
			b = p.ip.Unmap().appendTo4(b)
		} else {
			b = p.ip.appendTo6(b)
		}
	}

	b = append(b, '/')
	b = appendDecimal(b, uint8(p.Bits()))
	return b
}
