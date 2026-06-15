#include "main.h"

// -- Forward declarations --
static void testSplitHostPort(void);
static void testJoinHostPort(void);
static void noError(so_Error err);
static void testTCP(void);
static void testTCP_ResolveNamedPort(void);
static void testTCP_ResolveHostname(void);
static void testTCP_ResolveFamilyMismatch(void);
static void testTCP_Listen(void);
static void testTCP_ListenAll(void);
static void testTCP_Dial(void);
static void testTCP_DialRefused(void);
static void testTCP_ReadEOF(void);
static void testTCP_ReadDeadline(void);
static void testTCP_ClearDeadline(void);
static void testTCP_AcceptDeadline(void);
static void testTCP_CloseErrors(void);
static void testUDP(void);
static void testUDP_ResolveAddr(void);
static void testUDP_Listen(void);
static void testUDP_Dial(void);
static void testUDP_ReadFromWriteTo(void);
static void testUDP_ReadDeadline(void);
static void testUDP_CloseErrors(void);

// -- main.go --

int main(void) {
    so_println("%s", "solod.dev/so/net");
    testSplitHostPort();
    testJoinHostPort();
    testTCP();
    testUDP();
    return 0;
}

static void testSplitHostPort(void) {
    so_print("%s", "- split host-port...");
    net_HostPortResult _res1 = net_SplitHostPort(so_str("127.0.0.1:8080"));
    net_HostPort hp = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || so_string_ne(hp.Host, so_str("127.0.0.1")) || so_string_ne(hp.Port, so_str("8080"))) {
        so_panic("unexpected SplitHostPort result");
    }
    so_println("%s", "ok");
}

static void testJoinHostPort(void) {
    so_print("%s", "- join host-port...");
    so_byte buf[64] = {0};
    if (so_string_ne(net_JoinHostPort(so_array_slice(so_byte, buf, 0, 64, 64), so_str("::1"), so_str("80")), so_str("[::1]:80"))) {
        so_panic("unexpected JoinHostPort result");
    }
    so_println("%s", "ok");
}

static void noError(so_Error err) {
    if (err.self != NULL) {
        so_panic(so_error_cstr(err));
    }
}

// -- tcp.go --

static void testTCP(void) {
    testTCP_ResolveNamedPort();
    testTCP_ResolveHostname();
    testTCP_ResolveFamilyMismatch();
    testTCP_Listen();
    testTCP_ListenAll();
    testTCP_Dial();
    testTCP_DialRefused();
    testTCP_ReadEOF();
    testTCP_ReadDeadline();
    testTCP_ClearDeadline();
    testTCP_AcceptDeadline();
    testTCP_CloseErrors();
}

static void testTCP_ResolveNamedPort(void) {
    so_print("%s", "- TCP resolve a named port...");
    // A named port resolves via the services database (no DNS for the host).
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:http"));
    net_TCPAddr addr = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || addr.Port != 80) {
        so_panic("failed to resolve named port");
    }
    so_println("%s", "ok");
}

static void testTCP_ResolveHostname(void) {
    so_print("%s", "- TCP resolve a hostname...");
    // "localhost" resolves via getaddrinfo (the system resolver), without
    // any external DNS. It must come back as a loopback address.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("localhost:80"));
    net_TCPAddr addr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    if (addr.Port != 80) {
        so_panic("unexpected port");
    }
    if (!netip_Addr_IsLoopback(addr.IP)) {
        so_panic("localhost should resolve to a loopback address");
    }
    so_println("%s", "ok");
}

static void testTCP_ResolveFamilyMismatch(void) {
    so_print("%s", "- TCP resolve family mismatch...");
    // An IP literal must match the network's family: "tcp4" rejects an IPv6
    // literal, "tcp6" an IPv4 one. (Needs the real AF_* values, so this can
    // only run transpiled, not in the host test.)
    {
        net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp4"), so_str("[::1]:80"));
        so_Error err = _res1.err;
        if (err.self != net_ErrNoSuitableAddr.self) {
            so_panic("tcp4 should reject an IPv6 literal");
        }
    }
    {
        net_TCPAddrResult _res2 = net_ResolveTCPAddr(so_str("tcp6"), so_str("127.0.0.1:80"));
        so_Error err = _res2.err;
        if (err.self != net_ErrNoSuitableAddr.self) {
            so_panic("tcp6 should reject an IPv4 literal");
        }
    }
    so_println("%s", "ok");
}

static void testTCP_Listen(void) {
    so_print("%s", "- TCP listen...");
    // Resolve an IP literal (no DNS).
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr laddr = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || laddr.Port != 0) {
        so_panic("failed to resolve listen address");
    }
    // Listen on an OS-assigned port.
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &laddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    if (err.self != NULL) {
        so_panic(so_error_cstr(err));
    }
    net_TCPAddr addr = net_TCPListener_Addr(&ln);
    if (addr.Port == 0) {
        so_panic("listener port not assigned");
    }
    err = net_TCPListener_Close(&ln);
    if (err.self != NULL) {
        so_panic(so_error_cstr(err));
    }
    so_println("%s", "ok");
}

static void testTCP_ListenAll(void) {
    so_print("%s", "- TCP listen on all interfaces...");
    // A nil laddr binds the unspecified address (all interfaces), with an
    // OS-assigned port.
    net_TCPListenerResult _res1 = net_ListenTCP(so_str("tcp"), NULL);
    net_TCPListener ln = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    if (net_TCPListener_Addr(&ln).Port == 0) {
        so_panic("listener port not assigned");
    }
    noError(net_TCPListener_Close(&ln));
    so_println("%s", "ok");
}

static void testTCP_Dial(void) {
    // A single-threaded loopback echo. Without goroutines this works because the
    // connect completes into the listener backlog and the small payload fits in
    // the kernel socket buffers, so no call blocks waiting on another thread.
    so_print("%s", "- TCP dial...");
    // Listen on an OS-assigned port (IP literal, no DNS).
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    // Connect to the listener, binding to an explicit local address (an
    // ephemeral port on the loopback interface) to exercise bind-before-connect.
    net_TCPAddrResult _res3 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr laddr = _res3.val;
    err = _res3.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    net_TCPConnResult _res4 = net_DialTCP(so_str("tcp"), &laddr, &raddr);
    net_TCPConn conn = _res4.val;
    err = _res4.err;
    noError(err);
    // Accept the queued connection.
    net_TCPConnResult _res5 = net_TCPListener_Accept(&ln);
    net_TCPConn server = _res5.val;
    err = _res5.err;
    noError(err);
    // The endpoints' addresses must line up: the client's remote address is the
    // listener, and the server's remote address is the client's local address.
    if (net_TCPConn_RemoteAddr(&conn).Port != raddr.Port) {
        so_panic("client remote addr mismatch");
    }
    if (net_TCPConn_LocalAddr(&conn).Port == 0 || net_TCPConn_LocalAddr(&conn).Port != net_TCPConn_RemoteAddr(&server).Port) {
        so_panic("local/remote addr mismatch");
    }
    // Client writes, server echoes, client reads it back.
    so_Slice msg = so_string_bytes(so_str("ping"));
    {
        so_R_int_err _res6 = net_TCPConn_Write(&conn, msg);
        so_Error err = _res6.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    so_byte buf[256] = {0};
    so_R_int_err _res7 = net_TCPConn_Read(&server, so_array_slice(so_byte, buf, 0, 256, 256));
    so_int n = _res7.val;
    err = _res7.err;
    noError(err);
    {
        so_R_int_err _res8 = net_TCPConn_Write(&server, so_array_slice(so_byte, buf, 0, n, 256));
        so_Error err = _res8.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    so_byte got[256] = {0};
    so_R_int_err _res9 = net_TCPConn_Read(&conn, so_array_slice(so_byte, got, 0, 256, 256));
    n = _res9.val;
    err = _res9.err;
    noError(err);
    if (so_string_ne(so_bytes_string(so_array_slice(so_byte, got, 0, n, 256)), so_str("ping"))) {
        so_panic("echo mismatch");
    }
    net_TCPConn_Close(&conn);
    net_TCPConn_Close(&server);
    net_TCPListener_Close(&ln);
    so_println("%s", "ok");
}

static void testTCP_DialRefused(void) {
    so_print("%s", "- TCP dial refused...");
    // Bind a port, learn its address, then close the listener so nothing is
    // listening there. Dialing it must be refused.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    noError(net_TCPListener_Close(&ln));
    {
        net_TCPConnResult _res3 = net_DialTCP(so_str("tcp"), NULL, &raddr);
        so_Error err = _res3.err;
        if (err.self != net_ErrConnRefused.self) {
            so_panic("expected connection refused");
        }
    }
    so_println("%s", "ok");
}

static void testTCP_ReadEOF(void) {
    so_print("%s", "- TCP read EOF...");
    // Connect a pair, then close the server end. The client's next read must
    // report end of stream.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    net_TCPConnResult _res3 = net_DialTCP(so_str("tcp"), NULL, &raddr);
    net_TCPConn conn = _res3.val;
    err = _res3.err;
    noError(err);
    net_TCPConnResult _res4 = net_TCPListener_Accept(&ln);
    net_TCPConn server = _res4.val;
    err = _res4.err;
    noError(err);
    noError(net_TCPConn_Close(&server));
    so_byte buf[16] = {0};
    {
        so_R_int_err _res5 = net_TCPConn_Read(&conn, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res5.err;
        if (err.self != io_EOF.self) {
            so_panic("expected EOF");
        }
    }
    net_TCPConn_Close(&conn);
    net_TCPListener_Close(&ln);
    so_println("%s", "ok");
}

static void testTCP_ReadDeadline(void) {
    so_print("%s", "- TCP read deadline...");
    // Set up a connected pair, then read on the server side with no data sent.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    net_TCPConnResult _res3 = net_DialTCP(so_str("tcp"), NULL, &raddr);
    net_TCPConn conn = _res3.val;
    err = _res3.err;
    noError(err);
    net_TCPConnResult _res4 = net_TCPListener_Accept(&ln);
    net_TCPConn server = _res4.val;
    err = _res4.err;
    noError(err);
    // Nothing is written, so a read with a short deadline must time out.
    noError(net_TCPConn_SetReadDeadline(&server, time_Time_Add(time_Now(), 50 * time_Millisecond)));
    so_byte buf[16] = {0};
    {
        so_R_int_err _res5 = net_TCPConn_Read(&server, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res5.err;
        if (err.self != net_ErrTimeout.self) {
            so_panic("expected timeout");
        }
    }
    net_TCPConn_Close(&conn);
    net_TCPConn_Close(&server);
    net_TCPListener_Close(&ln);
    so_println("%s", "ok");
}

static void testTCP_ClearDeadline(void) {
    so_print("%s", "- TCP clear deadline...");
    // After a read deadline fires, clearing it must leave the connection usable.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    net_TCPConnResult _res3 = net_DialTCP(so_str("tcp"), NULL, &raddr);
    net_TCPConn conn = _res3.val;
    err = _res3.err;
    noError(err);
    net_TCPConnResult _res4 = net_TCPListener_Accept(&ln);
    net_TCPConn server = _res4.val;
    err = _res4.err;
    noError(err);
    // Arm a short deadline and let it elapse with no data.
    noError(net_TCPConn_SetReadDeadline(&server, time_Time_Add(time_Now(), 50 * time_Millisecond)));
    so_byte buf[16] = {0};
    {
        so_R_int_err _res5 = net_TCPConn_Read(&server, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res5.err;
        if (err.self != net_ErrTimeout.self) {
            so_panic("expected timeout");
        }
    }
    // Clearing the deadline must let a read of already-sent data succeed
    // instead of timing out. (Data is sent first because there is no second
    // thread to write during a blocking read.)
    so_R_int_err _res6 = net_TCPConn_Write(&conn, so_string_bytes(so_str("hi")));
    err = _res6.err;
    noError(err);
    noError(net_TCPConn_SetReadDeadline(&server, (time_Time){}));
    so_R_int_err _res7 = net_TCPConn_Read(&server, so_array_slice(so_byte, buf, 0, 16, 16));
    so_int n = _res7.val;
    err = _res7.err;
    noError(err);
    if (so_string_ne(so_bytes_string(so_array_slice(so_byte, buf, 0, n, 16)), so_str("hi"))) {
        so_panic("read after clearing deadline failed");
    }
    net_TCPConn_Close(&conn);
    net_TCPConn_Close(&server);
    net_TCPListener_Close(&ln);
    so_println("%s", "ok");
}

static void testTCP_AcceptDeadline(void) {
    so_print("%s", "- TCP accept deadline...");
    // A listener with a short deadline and no incoming connection must time out.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    noError(net_TCPListener_SetDeadline(&ln, time_Time_Add(time_Now(), 50 * time_Millisecond)));
    {
        net_TCPConnResult _res3 = net_TCPListener_Accept(&ln);
        so_Error err = _res3.err;
        if (err.self != net_ErrTimeout.self) {
            so_panic("expected timeout");
        }
    }
    noError(net_TCPListener_Close(&ln));
    so_println("%s", "ok");
}

static void testTCP_CloseErrors(void) {
    so_print("%s", "- TCP close errors...");
    // A double close, and any I/O after close, must report ErrClosed on both
    // connections and listeners.
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:0"));
    net_TCPAddr lnAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_TCPListenerResult _res2 = net_ListenTCP(so_str("tcp"), &lnAddr);
    net_TCPListener ln = _res2.val;
    err = _res2.err;
    noError(err);
    net_TCPAddr raddr = net_TCPListener_Addr(&ln);
    net_TCPConnResult _res3 = net_DialTCP(so_str("tcp"), NULL, &raddr);
    net_TCPConn conn = _res3.val;
    err = _res3.err;
    noError(err);
    net_TCPConnResult _res4 = net_TCPListener_Accept(&ln);
    net_TCPConn server = _res4.val;
    err = _res4.err;
    noError(err);
    noError(net_TCPConn_Close(&conn));
    {
        so_Error err = net_TCPConn_Close(&conn);
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on double close");
        }
    }
    so_byte buf[16] = {0};
    {
        so_R_int_err _res5 = net_TCPConn_Read(&conn, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res5.err;
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on read after close");
        }
    }
    {
        so_R_int_err _res6 = net_TCPConn_Write(&conn, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res6.err;
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on write after close");
        }
    }
    noError(net_TCPListener_Close(&ln));
    {
        so_Error err = net_TCPListener_Close(&ln);
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on double close (listener)");
        }
    }
    {
        net_TCPConnResult _res7 = net_TCPListener_Accept(&ln);
        so_Error err = _res7.err;
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on accept after close");
        }
    }
    net_TCPConn_Close(&server);
    so_println("%s", "ok");
}

// -- udp.go --

static void testUDP(void) {
    testUDP_ResolveAddr();
    testUDP_Listen();
    testUDP_Dial();
    testUDP_ReadFromWriteTo();
    testUDP_ReadDeadline();
    testUDP_CloseErrors();
}

static void testUDP_ResolveAddr(void) {
    so_print("%s", "- UDP resolve a named port...");
    // A named port resolves via the udp services database (no DNS for the host).
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:domain"));
    net_UDPAddr addr = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || addr.Port != 53) {
        so_panic("failed to resolve named UDP port");
    }
    so_println("%s", "ok");
    so_print("%s", "- UDP resolve a hostname...");
    // "localhost" resolves via getaddrinfo (the system resolver), without
    // any external DNS. It must come back as a loopback address.
    net_UDPAddrResult _res2 = net_ResolveUDPAddr(so_str("udp"), so_str("localhost:53"));
    addr = _res2.val;
    err = _res2.err;
    noError(err);
    if (addr.Port != 53) {
        so_panic("unexpected port");
    }
    if (!netip_Addr_IsLoopback(addr.IP)) {
        so_panic("localhost should resolve to a loopback address");
    }
    so_println("%s", "ok");
}

static void testUDP_Listen(void) {
    so_print("%s", "- UDP listen...");
    // Resolve an IP literal (no DNS) and listen on an OS-assigned port.
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr laddr = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || laddr.Port != 0) {
        so_panic("failed to resolve listen address");
    }
    net_UDPConnResult _res2 = net_ListenUDP(so_str("udp"), &laddr);
    net_UDPConn conn = _res2.val;
    err = _res2.err;
    noError(err);
    if (net_UDPConn_LocalAddr(&conn).Port == 0) {
        so_panic("listener port not assigned");
    }
    noError(net_UDPConn_Close(&conn));
    so_println("%s", "ok");
}

static void testUDP_Dial(void) {
    // A single-threaded loopback echo. Datagrams are buffered in the kernel, so
    // no call blocks waiting on another thread.
    so_print("%s", "- UDP dial...");
    // Server listens on an OS-assigned port.
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr srvAddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_UDPConnResult _res2 = net_ListenUDP(so_str("udp"), &srvAddr);
    net_UDPConn server = _res2.val;
    err = _res2.err;
    noError(err);
    // Client connects to the server.
    net_UDPAddr raddr = net_UDPConn_LocalAddr(&server);
    net_UDPConnResult _res3 = net_DialUDP(so_str("udp"), NULL, &raddr);
    net_UDPConn client = _res3.val;
    err = _res3.err;
    noError(err);
    // The client's remote address is the server.
    if (net_UDPConn_RemoteAddr(&client).Port != raddr.Port) {
        so_panic("client remote addr mismatch");
    }
    // Client writes a datagram; the server receives it and learns the client's
    // address, then echoes it back via WriteTo.
    {
        so_R_int_err _res4 = net_UDPConn_Write(&client, so_string_bytes(so_str("ping")));
        so_Error err = _res4.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    so_byte buf[256] = {0};
    net_UDPReadResult _res5 = net_UDPConn_ReadFrom(&server, so_array_slice(so_byte, buf, 0, 256, 256));
    net_UDPRead r = _res5.val;
    err = _res5.err;
    noError(err);
    if (r.Addr.Port != net_UDPConn_LocalAddr(&client).Port) {
        so_panic("server learned wrong client addr");
    }
    {
        so_R_int_err _res6 = net_UDPConn_WriteTo(&server, so_array_slice(so_byte, buf, 0, r.N, 256), &r.Addr);
        so_Error err = _res6.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    // Client reads the echo on its connected socket.
    so_byte got[256] = {0};
    so_R_int_err _res7 = net_UDPConn_Read(&client, so_array_slice(so_byte, got, 0, 256, 256));
    so_int n = _res7.val;
    err = _res7.err;
    noError(err);
    if (so_string_ne(so_bytes_string(so_array_slice(so_byte, got, 0, n, 256)), so_str("ping"))) {
        so_panic("echo mismatch");
    }
    net_UDPConn_Close(&client);
    net_UDPConn_Close(&server);
    so_println("%s", "ok");
}

static void testUDP_ReadFromWriteTo(void) {
    so_print("%s", "- UDP ReadFrom/WriteTo...");
    // Two unconnected sockets exchange datagrams in both directions, with each
    // receiver checking the reported source address against the sender's local
    // address.
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr addrA = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_UDPConnResult _res2 = net_ListenUDP(so_str("udp"), &addrA);
    net_UDPConn a = _res2.val;
    err = _res2.err;
    noError(err);
    net_UDPAddrResult _res3 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr addrB = _res3.val;
    err = _res3.err;
    noError(err);
    net_UDPConnResult _res4 = net_ListenUDP(so_str("udp"), &addrB);
    net_UDPConn b = _res4.val;
    err = _res4.err;
    noError(err);
    // A -> B.
    net_UDPAddr bAddr = net_UDPConn_LocalAddr(&b);
    {
        so_R_int_err _res5 = net_UDPConn_WriteTo(&a, so_string_bytes(so_str("ping")), &bAddr);
        so_Error err = _res5.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    so_byte buf[256] = {0};
    net_UDPReadResult _res6 = net_UDPConn_ReadFrom(&b, so_array_slice(so_byte, buf, 0, 256, 256));
    net_UDPRead r = _res6.val;
    err = _res6.err;
    noError(err);
    if (so_string_ne(so_bytes_string(so_array_slice(so_byte, buf, 0, r.N, 256)), so_str("ping"))) {
        so_panic("A->B payload mismatch");
    }
    if (r.Addr.Port != net_UDPConn_LocalAddr(&a).Port) {
        so_panic("A->B source addr mismatch");
    }
    // B -> A, replying to the learned source address.
    {
        so_R_int_err _res7 = net_UDPConn_WriteTo(&b, so_string_bytes(so_str("pong")), &r.Addr);
        so_Error err = _res7.err;
        if (err.self != NULL) {
            so_panic(so_error_cstr(err));
        }
    }
    so_byte buf2[256] = {0};
    net_UDPReadResult _res8 = net_UDPConn_ReadFrom(&a, so_array_slice(so_byte, buf2, 0, 256, 256));
    net_UDPRead r2 = _res8.val;
    err = _res8.err;
    noError(err);
    if (so_string_ne(so_bytes_string(so_array_slice(so_byte, buf2, 0, r2.N, 256)), so_str("pong"))) {
        so_panic("B->A payload mismatch");
    }
    if (r2.Addr.Port != net_UDPConn_LocalAddr(&b).Port) {
        so_panic("B->A source addr mismatch");
    }
    net_UDPConn_Close(&a);
    net_UDPConn_Close(&b);
    so_println("%s", "ok");
}

static void testUDP_ReadDeadline(void) {
    so_print("%s", "- UDP read deadline...");
    // A ReadFrom with a short deadline and no data must time out.
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr laddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_UDPConnResult _res2 = net_ListenUDP(so_str("udp"), &laddr);
    net_UDPConn conn = _res2.val;
    err = _res2.err;
    noError(err);
    noError(net_UDPConn_SetReadDeadline(&conn, time_Time_Add(time_Now(), 50 * time_Millisecond)));
    so_byte buf[16] = {0};
    {
        net_UDPReadResult _res3 = net_UDPConn_ReadFrom(&conn, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res3.err;
        if (err.self != net_ErrTimeout.self) {
            so_panic("expected timeout");
        }
    }
    noError(net_UDPConn_Close(&conn));
    so_println("%s", "ok");
}

static void testUDP_CloseErrors(void) {
    so_print("%s", "- UDP close errors...");
    // A double close, and any I/O after close, must report ErrClosed.
    net_UDPAddrResult _res1 = net_ResolveUDPAddr(so_str("udp"), so_str("127.0.0.1:0"));
    net_UDPAddr laddr = _res1.val;
    so_Error err = _res1.err;
    noError(err);
    net_UDPConnResult _res2 = net_ListenUDP(so_str("udp"), &laddr);
    net_UDPConn conn = _res2.val;
    err = _res2.err;
    noError(err);
    noError(net_UDPConn_Close(&conn));
    {
        so_Error err = net_UDPConn_Close(&conn);
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on double close");
        }
    }
    so_byte buf[16] = {0};
    {
        net_UDPReadResult _res3 = net_UDPConn_ReadFrom(&conn, so_array_slice(so_byte, buf, 0, 16, 16));
        so_Error err = _res3.err;
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on ReadFrom after close");
        }
    }
    {
        so_R_int_err _res4 = net_UDPConn_WriteTo(&conn, so_array_slice(so_byte, buf, 0, 16, 16), &laddr);
        so_Error err = _res4.err;
        if (err.self != net_ErrClosed.self) {
            so_panic("expected ErrClosed on WriteTo after close");
        }
    }
    so_println("%s", "ok");
}
