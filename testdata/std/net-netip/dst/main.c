#include "main.h"

// -- Implementation --

int main(void) {
    so_byte buf[54] = {0};
    {
        // Parse IPv4 address.
        netip_AddrResult _res1 = netip_ParseAddr(so_str("192.168.140.255"));
        netip_Addr ip4 = _res1.val;
        so_Error err = _res1.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
        so_byte a4[4] = {0};
        memcpy(a4, netip_Addr_As4(ip4, a4), sizeof(a4));
        if (so_mem_ne(a4, ((so_byte[4]){192, 168, 140, 255}), 4 * sizeof(so_byte))) {
            so_panic("unexpected IPv4 bytes");
        }
    }
    {
        // Parse IPv6 address.
        netip_AddrResult _res2 = netip_ParseAddr(so_str("fd7a:115c::626b:430b"));
        netip_Addr ip6 = _res2.val;
        so_Error err = _res2.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
        so_byte a16[16] = {0};
        memcpy(a16, netip_Addr_As16(ip6, a16), sizeof(a16));
        if (so_mem_ne(a16, ((so_byte[16]){0xfd, 0x7a, 0x11, 0x5c, [12] = 0x62, 0x6b, 0x43, 0x0b}), 16 * sizeof(so_byte))) {
            so_panic("unexpected IPv6 bytes");
        }
    }
    {
        // Addr.String.
        netip_Addr ip = netip_MustParseAddr(so_str("10.0.0.1"));
        if (so_string_ne(netip_Addr_String(ip, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("10.0.0.1"))) {
            so_panic("Addr.String IPv4");
        }
        ip = netip_MustParseAddr(so_str("2001:db8::1"));
        if (so_string_ne(netip_Addr_String(ip, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("2001:db8::1"))) {
            so_panic("Addr.String IPv6");
        }
    }
    {
        // Addr classification.
        netip_Addr ip4 = netip_MustParseAddr(so_str("1.2.3.4"));
        if (!netip_Addr_Is4(ip4)) {
            so_panic("Is4");
        }
        if (netip_Addr_Is6(ip4)) {
            so_panic("Is6 for v4");
        }
        netip_Addr ip6 = netip_MustParseAddr(so_str("::1"));
        if (netip_Addr_Is4(ip6)) {
            so_panic("Is4 for v6");
        }
        if (!netip_Addr_Is6(ip6)) {
            so_panic("Is6");
        }
    }
    {
        // Addr properties.
        if (!netip_Addr_IsLoopback(netip_MustParseAddr(so_str("127.0.0.1")))) {
            so_panic("IsLoopback v4");
        }
        if (!netip_Addr_IsLoopback(netip_MustParseAddr(so_str("::1")))) {
            so_panic("IsLoopback v6");
        }
        if (!netip_Addr_IsPrivate(netip_MustParseAddr(so_str("10.0.0.1")))) {
            so_panic("IsPrivate");
        }
        if (!netip_Addr_IsMulticast(netip_MustParseAddr(so_str("224.0.0.1")))) {
            so_panic("IsMulticast");
        }
    }
    {
        // Addr.Compare.
        netip_Addr a = netip_MustParseAddr(so_str("1.2.3.4"));
        netip_Addr b = netip_MustParseAddr(so_str("1.2.3.5"));
        if (netip_Addr_Compare(a, b) != -1) {
            so_panic("Compare less");
        }
        if (netip_Addr_Compare(b, a) != 1) {
            so_panic("Compare greater");
        }
        if (netip_Addr_Compare(a, a) != 0) {
            so_panic("Compare equal");
        }
    }
    {
        // Addr.Next and Addr.Prev.
        netip_Addr ip = netip_MustParseAddr(so_str("1.2.3.4"));
        netip_Addr next = netip_Addr_Next(ip);
        if (so_string_ne(netip_Addr_String(next, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("1.2.3.5"))) {
            so_panic("Addr.Next");
        }
        netip_Addr prev = netip_Addr_Prev(next);
        if (!netip_Addr_Equal(prev, ip)) {
            so_panic("Addr.Prev");
        }
    }
    {
        // Addr.Unmap (4-in-6).
        netip_Addr ip = netip_MustParseAddr(so_str("::ffff:1.2.3.4"));
        if (!netip_Addr_Is4In6(ip)) {
            so_panic("Is4In6");
        }
        netip_Addr unmapped = netip_Addr_Unmap(ip);
        if (!netip_Addr_Is4(unmapped)) {
            so_panic("Unmap Is4");
        }
        if (so_string_ne(netip_Addr_String(unmapped, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("1.2.3.4"))) {
            so_panic("Unmap String");
        }
    }
    {
        // AddrFrom4 and AddrFrom16.
        netip_Addr ip4 = netip_AddrFrom4((so_byte[4]){10, 20, 30, 40});
        if (so_string_ne(netip_Addr_String(ip4, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("10.20.30.40"))) {
            so_panic("AddrFrom4");
        }
        netip_Addr ip6 = netip_AddrFrom16((so_byte[16]){0x20, 0x01, 0x0d, 0xb8, [15] = 0x01});
        if (so_string_ne(netip_Addr_String(ip6, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("2001:db8::1"))) {
            so_panic("AddrFrom16");
        }
    }
    {
        // AddrPort.
        netip_AddrPortResult _res3 = netip_ParseAddrPort(so_str("192.168.1.1:8080"));
        netip_AddrPort ap = _res3.val;
        so_Error err = _res3.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
        netip_Addr addr = netip_AddrPort_Addr(ap);
        if (so_string_ne(netip_Addr_String(addr, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("192.168.1.1"))) {
            so_panic("AddrPort.Addr");
        }
        if (netip_AddrPort_Port(ap) != 8080) {
            so_panic("AddrPort.Port");
        }
        if (so_string_ne(netip_AddrPort_String(ap, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("192.168.1.1:8080"))) {
            so_panic("AddrPort.String v4");
        }
    }
    {
        // AddrPort IPv6.
        netip_AddrPort ap = netip_MustParseAddrPort(so_str("[::1]:443"));
        if (so_string_ne(netip_AddrPort_String(ap, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("[::1]:443"))) {
            so_panic("AddrPort.String v6");
        }
    }
    {
        // Prefix.
        netip_PrefixResult _res4 = netip_ParsePrefix(so_str("192.168.1.0/24"));
        netip_Prefix pfx = _res4.val;
        so_Error err = _res4.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
        if (netip_Prefix_Bits(pfx) != 24) {
            so_panic("Prefix.Bits");
        }
        if (so_string_ne(netip_Prefix_String(pfx, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("192.168.1.0/24"))) {
            so_panic("Prefix.String");
        }
        if (!netip_Prefix_Contains(pfx, netip_MustParseAddr(so_str("192.168.1.100")))) {
            so_panic("Prefix.Contains true");
        }
        if (netip_Prefix_Contains(pfx, netip_MustParseAddr(so_str("192.168.2.1")))) {
            so_panic("Prefix.Contains false");
        }
    }
    {
        // Prefix.Masked.
        netip_Prefix pfx = netip_MustParsePrefix(so_str("192.168.1.1/24"));
        netip_Prefix masked = netip_Prefix_Masked(pfx);
        netip_Addr maskedAddr = netip_Prefix_Addr(masked);
        if (so_string_ne(netip_Addr_String(maskedAddr, so_array_slice(so_byte, buf, 0, 54, 54)), so_str("192.168.1.0"))) {
            so_panic("Prefix.Masked");
        }
    }
    {
        // Prefix.Overlaps.
        netip_Prefix a = netip_MustParsePrefix(so_str("192.168.0.0/16"));
        netip_Prefix b = netip_MustParsePrefix(so_str("192.168.1.0/24"));
        if (!netip_Prefix_Overlaps(a, b)) {
            so_panic("Prefix.Overlaps true");
        }
        netip_Prefix c = netip_MustParsePrefix(so_str("10.0.0.0/8"));
        if (netip_Prefix_Overlaps(a, c)) {
            so_panic("Prefix.Overlaps false");
        }
    }
    return 0;
}
