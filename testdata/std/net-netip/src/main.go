package main

import "solod.dev/so/net/netip"

func main() {
	var buf [netip.MaxAddrPortLen]byte
	{
		// Parse IPv4 address.
		ip4, err := netip.ParseAddr("192.168.140.255")
		if err != nil {
			panic(err)
		}
		var a4 [4]byte
		a4 = ip4.As4(a4)
		if a4 != [4]byte{192, 168, 140, 255} {
			panic("unexpected IPv4 bytes")
		}
	}
	{
		// Parse IPv6 address.
		ip6, err := netip.ParseAddr("fd7a:115c::626b:430b")
		if err != nil {
			panic(err)
		}
		var a16 [16]byte
		a16 = ip6.As16(a16)
		if a16 != [16]byte{0xfd, 0x7a, 0x11, 0x5c, 12: 0x62, 0x6b, 0x43, 0x0b} {
			panic("unexpected IPv6 bytes")
		}
	}
	{
		// Addr.String.
		ip := netip.MustParseAddr("10.0.0.1")
		if ip.String(buf[:]) != "10.0.0.1" {
			panic("Addr.String IPv4")
		}
		ip = netip.MustParseAddr("2001:db8::1")
		if ip.String(buf[:]) != "2001:db8::1" {
			panic("Addr.String IPv6")
		}
	}
	{
		// Addr classification.
		ip4 := netip.MustParseAddr("1.2.3.4")
		if !ip4.Is4() {
			panic("Is4")
		}
		if ip4.Is6() {
			panic("Is6 for v4")
		}
		ip6 := netip.MustParseAddr("::1")
		if ip6.Is4() {
			panic("Is4 for v6")
		}
		if !ip6.Is6() {
			panic("Is6")
		}
	}
	{
		// Addr properties.
		if !netip.MustParseAddr("127.0.0.1").IsLoopback() {
			panic("IsLoopback v4")
		}
		if !netip.MustParseAddr("::1").IsLoopback() {
			panic("IsLoopback v6")
		}
		if !netip.MustParseAddr("10.0.0.1").IsPrivate() {
			panic("IsPrivate")
		}
		if !netip.MustParseAddr("224.0.0.1").IsMulticast() {
			panic("IsMulticast")
		}
	}
	{
		// Addr.Compare.
		a := netip.MustParseAddr("1.2.3.4")
		b := netip.MustParseAddr("1.2.3.5")
		if a.Compare(b) != -1 {
			panic("Compare less")
		}
		if b.Compare(a) != 1 {
			panic("Compare greater")
		}
		if a.Compare(a) != 0 {
			panic("Compare equal")
		}
	}
	{
		// Addr.Next and Addr.Prev.
		ip := netip.MustParseAddr("1.2.3.4")
		next := ip.Next()
		if next.String(buf[:]) != "1.2.3.5" {
			panic("Addr.Next")
		}
		prev := next.Prev()
		if !prev.Equal(ip) {
			panic("Addr.Prev")
		}
	}
	{
		// Addr.Unmap (4-in-6).
		ip := netip.MustParseAddr("::ffff:1.2.3.4")
		if !ip.Is4In6() {
			panic("Is4In6")
		}
		unmapped := ip.Unmap()
		if !unmapped.Is4() {
			panic("Unmap Is4")
		}
		if unmapped.String(buf[:]) != "1.2.3.4" {
			panic("Unmap String")
		}
	}
	{
		// AddrFrom4 and AddrFrom16.
		ip4 := netip.AddrFrom4([4]byte{10, 20, 30, 40})
		if ip4.String(buf[:]) != "10.20.30.40" {
			panic("AddrFrom4")
		}
		ip6 := netip.AddrFrom16([16]byte{0x20, 0x01, 0x0d, 0xb8, 15: 0x01})
		if ip6.String(buf[:]) != "2001:db8::1" {
			panic("AddrFrom16")
		}
	}
	{
		// AddrPort.
		ap, err := netip.ParseAddrPort("192.168.1.1:8080")
		if err != nil {
			panic(err)
		}
		addr := ap.Addr()
		if addr.String(buf[:]) != "192.168.1.1" {
			panic("AddrPort.Addr")
		}
		if ap.Port() != 8080 {
			panic("AddrPort.Port")
		}
		if ap.String(buf[:]) != "192.168.1.1:8080" {
			panic("AddrPort.String v4")
		}
	}
	{
		// AddrPort IPv6.
		ap := netip.MustParseAddrPort("[::1]:443")
		if ap.String(buf[:]) != "[::1]:443" {
			panic("AddrPort.String v6")
		}
	}
	{
		// Prefix.
		pfx, err := netip.ParsePrefix("192.168.1.0/24")
		if err != nil {
			panic(err)
		}
		if pfx.Bits() != 24 {
			panic("Prefix.Bits")
		}
		if pfx.String(buf[:]) != "192.168.1.0/24" {
			panic("Prefix.String")
		}
		if !pfx.Contains(netip.MustParseAddr("192.168.1.100")) {
			panic("Prefix.Contains true")
		}
		if pfx.Contains(netip.MustParseAddr("192.168.2.1")) {
			panic("Prefix.Contains false")
		}
	}
	{
		// Prefix.Masked.
		pfx := netip.MustParsePrefix("192.168.1.1/24")
		masked := pfx.Masked()
		maskedAddr := masked.Addr()
		if maskedAddr.String(buf[:]) != "192.168.1.0" {
			panic("Prefix.Masked")
		}
	}
	{
		// Prefix.Overlaps.
		a := netip.MustParsePrefix("192.168.0.0/16")
		b := netip.MustParsePrefix("192.168.1.0/24")
		if !a.Overlaps(b) {
			panic("Prefix.Overlaps true")
		}
		c := netip.MustParsePrefix("10.0.0.0/8")
		if a.Overlaps(c) {
			panic("Prefix.Overlaps false")
		}
	}
}
