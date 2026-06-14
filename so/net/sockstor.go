package net

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
	"solod.dev/so/net/netip"
)

// sockAddr returns a *sockaddr view of the address storage.
func (stor *sockaddr_storage) sockAddr() *sockaddr {
	return c.PtrAs[sockaddr](stor)
}

// tcpAddr decodes the sockaddr in stor into a TCPAddr.
// Returns the zero TCPAddr if the family is not recognized.
func (stor *sockaddr_storage) tcpAddr() TCPAddr {
	base := stor.sockAddr()
	if base.sa_family == c_AF_INET {
		s4 := c.PtrAs[sockaddr_in](stor)
		var ip [4]byte
		mem.Copy(&ip[0], &s4.sin_addr, 4)
		ipAddr := netip.AddrFromSlice(ip[:])
		port := int(ntohs(s4.sin_port))
		return TCPAddr{IP: ipAddr, Port: port}
	}
	if base.sa_family == c_AF_INET6 {
		s6 := c.PtrAs[sockaddr_in6](stor)
		var ip [16]byte
		mem.Copy(&ip[0], &s6.sin6_addr, 16)
		ipAddr := netip.AddrFromSlice(ip[:])
		port := int(ntohs(s6.sin6_port))
		return TCPAddr{IP: ipAddr, Port: port}
	}
	return TCPAddr{}
}

// fill encodes addr into stor as a sockaddr_in or sockaddr_in6 and returns
// its length. If addr's IP is invalid (neither IPv4 nor IPv6), fill does
// nothing and returns 0.
func (stor *sockaddr_storage) fill(addr TCPAddr) c.UInt {
	var ipbuf [16]byte
	ip := addr.IP.AsSlice(ipbuf[:])
	port := uint16(addr.Port)
	mem.Clear(stor, c.Sizeof[sockaddr_storage]())
	if len(ip) == 4 {
		s4 := c.PtrAs[sockaddr_in](stor)
		s4.sin_family = c_AF_INET
		s4.sin_port = htons(port)
		mem.Copy(&s4.sin_addr, &ip[0], 4)
		return c.UInt(c.Sizeof[sockaddr_in]())
	}
	if len(ip) == 16 {
		s6 := c.PtrAs[sockaddr_in6](stor)
		s6.sin6_family = c_AF_INET6
		s6.sin6_port = htons(port)
		mem.Copy(&s6.sin6_addr, &ip[0], 16)
		return c.UInt(c.Sizeof[sockaddr_in6]())
	}
	return 0
}
