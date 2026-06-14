#include "main.h"

// -- Forward declarations --
static void testSplitHostPort(void);
static void testJoinHostPort(void);
static void testResolveNamedPort(void);
static void testResolveHostname(void);
static void testResolveFamilyMismatch(void);
static void testListen(void);
static void testListenAll(void);
static void testDial(void);
static void testDialRefused(void);
static void testReadEOF(void);
static void testReadDeadline(void);
static void testClearDeadline(void);
static void testAcceptDeadline(void);
static void testCloseErrors(void);
static void noError(so_Error err);

// -- Implementation --

int main(void) {
    so_println("%s", "solod.dev/so/net");
    testSplitHostPort();
    testJoinHostPort();
    testResolveNamedPort();
    testResolveHostname();
    testResolveFamilyMismatch();
    testListen();
    testListenAll();
    testDial();
    testDialRefused();
    testReadEOF();
    testReadDeadline();
    testClearDeadline();
    testAcceptDeadline();
    testCloseErrors();
    return 0;
}

static void testSplitHostPort(void) {
    so_print("%s", "- split host:port...");
    net_HostPortResult _res1 = net_SplitHostPort(so_str("127.0.0.1:8080"));
    net_HostPort hp = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || so_string_ne(hp.Host, so_str("127.0.0.1")) || so_string_ne(hp.Port, so_str("8080"))) {
        so_panic("unexpected SplitHostPort result");
    }
    so_println("%s", "ok");
}

static void testJoinHostPort(void) {
    so_print("%s", "- join host:port...");
    so_byte buf[64] = {0};
    if (so_string_ne(net_JoinHostPort(so_array_slice(so_byte, buf, 0, 64, 64), so_str("::1"), so_str("80")), so_str("[::1]:80"))) {
        so_panic("unexpected JoinHostPort result");
    }
    so_println("%s", "ok");
}

static void testResolveNamedPort(void) {
    so_print("%s", "- resolve a named port...");
    // A named port resolves via the services database (no DNS for the host).
    net_TCPAddrResult _res1 = net_ResolveTCPAddr(so_str("tcp"), so_str("127.0.0.1:http"));
    net_TCPAddr addr = _res1.val;
    so_Error err = _res1.err;
    if (err.self != NULL || addr.Port != 80) {
        so_panic("failed to resolve named port");
    }
    so_println("%s", "ok");
}

static void testResolveHostname(void) {
    so_print("%s", "- resolve a hostname...");
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

static void testResolveFamilyMismatch(void) {
    so_print("%s", "- resolve family mismatch...");
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

static void testListen(void) {
    so_print("%s", "- listen...");
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

static void testListenAll(void) {
    so_print("%s", "- listen on all interfaces...");
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

static void testDial(void) {
    // A single-threaded loopback echo. Without goroutines this works because the
    // connect completes into the listener backlog and the small payload fits in
    // the kernel socket buffers, so no call blocks waiting on another thread.
    so_print("%s", "- dial...");
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

static void testDialRefused(void) {
    so_print("%s", "- dial refused...");
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

static void testReadEOF(void) {
    so_print("%s", "- read EOF...");
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

static void testReadDeadline(void) {
    so_print("%s", "- read deadline...");
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

static void testClearDeadline(void) {
    so_print("%s", "- clear deadline...");
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

static void testAcceptDeadline(void) {
    so_print("%s", "- accept deadline...");
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

static void testCloseErrors(void) {
    so_print("%s", "- close errors...");
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

static void noError(so_Error err) {
    if (err.self != NULL) {
        so_panic(so_error_cstr(err));
    }
}
