package main

import (
	"solod.dev/so/io"
	"solod.dev/so/net"
	"solod.dev/so/time"
)

func testTCP() {
	testTCP_ResolveNamedPort()
	testTCP_ResolveHostname()
	testTCP_ResolveFamilyMismatch()
	testTCP_Listen()
	testTCP_ListenAll()
	testTCP_Dial()
	testTCP_DialRefused()
	testTCP_ReadEOF()
	testTCP_ReadDeadline()
	testTCP_ClearDeadline()
	testTCP_AcceptDeadline()
	testTCP_CloseErrors()
}

func testTCP_ResolveNamedPort() {
	print("- TCP resolve a named port...")
	// A named port resolves via the services database (no DNS for the host).
	addr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:http")
	if err != nil || addr.Port != 80 {
		panic("failed to resolve named port")
	}
	println("ok")
}

func testTCP_ResolveHostname() {
	print("- TCP resolve a hostname...")
	// "localhost" resolves via getaddrinfo (the system resolver), without
	// any external DNS. It must come back as a loopback address.
	addr, err := net.ResolveTCPAddr("tcp", "localhost:80")
	noError(err)
	if addr.Port != 80 {
		panic("unexpected port")
	}
	if !addr.IP.IsLoopback() {
		panic("localhost should resolve to a loopback address")
	}
	println("ok")
}

func testTCP_ResolveFamilyMismatch() {
	print("- TCP resolve family mismatch...")
	// An IP literal must match the network's family: "tcp4" rejects an IPv6
	// literal, "tcp6" an IPv4 one. (Needs the real AF_* values, so this can
	// only run transpiled, not in the host test.)
	if _, err := net.ResolveTCPAddr("tcp4", "[::1]:80"); err != net.ErrNoSuitableAddr {
		panic("tcp4 should reject an IPv6 literal")
	}
	if _, err := net.ResolveTCPAddr("tcp6", "127.0.0.1:80"); err != net.ErrNoSuitableAddr {
		panic("tcp6 should reject an IPv4 literal")
	}
	println("ok")
}

func testTCP_Listen() {
	print("- TCP listen...")
	// Resolve an IP literal (no DNS).
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	if err != nil || laddr.Port != 0 {
		panic("failed to resolve listen address")
	}

	// Listen on an OS-assigned port.
	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		panic(err)
	}

	addr := ln.Addr()
	if addr.Port == 0 {
		panic("listener port not assigned")
	}

	err = ln.Close()
	if err != nil {
		panic(err)
	}
	println("ok")
}

func testTCP_ListenAll() {
	print("- TCP listen on all interfaces...")
	// A nil laddr binds the unspecified address (all interfaces), with an
	// OS-assigned port.
	ln, err := net.ListenTCP("tcp", nil)
	noError(err)
	if ln.Addr().Port == 0 {
		panic("listener port not assigned")
	}
	noError(ln.Close())
	println("ok")
}

func testTCP_Dial() {
	// A single-threaded loopback echo. Without goroutines this works because the
	// connect completes into the listener backlog and the small payload fits in
	// the kernel socket buffers, so no call blocks waiting on another thread.

	print("- TCP dial...")
	// Listen on an OS-assigned port (IP literal, no DNS).
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)

	// Connect to the listener, binding to an explicit local address (an
	// ephemeral port on the loopback interface) to exercise bind-before-connect.
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", &laddr, &raddr)
	noError(err)

	// Accept the queued connection.
	server, err := ln.Accept()
	noError(err)

	// The endpoints' addresses must line up: the client's remote address is the
	// listener, and the server's remote address is the client's local address.
	if conn.RemoteAddr().Port != raddr.Port {
		panic("client remote addr mismatch")
	}
	if conn.LocalAddr().Port == 0 || conn.LocalAddr().Port != server.RemoteAddr().Port {
		panic("local/remote addr mismatch")
	}

	// Client writes, server echoes, client reads it back.
	msg := []byte("ping")
	if _, err := conn.Write(msg); err != nil {
		panic(err)
	}

	var buf [256]byte
	n, err := server.Read(buf[:])
	noError(err)
	if _, err := server.Write(buf[:n]); err != nil {
		panic(err)
	}

	var got [256]byte
	n, err = conn.Read(got[:])
	noError(err)
	if string(got[:n]) != "ping" {
		panic("echo mismatch")
	}

	conn.Close()
	server.Close()
	ln.Close()
	println("ok")
}

func testTCP_DialRefused() {
	print("- TCP dial refused...")
	// Bind a port, learn its address, then close the listener so nothing is
	// listening there. Dialing it must be refused.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)
	raddr := ln.Addr()
	noError(ln.Close())

	if _, err := net.DialTCP("tcp", nil, &raddr); err != net.ErrConnRefused {
		panic("expected connection refused")
	}
	println("ok")
}

func testTCP_ReadEOF() {
	print("- TCP read EOF...")
	// Connect a pair, then close the server end. The client's next read must
	// report end of stream.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	noError(err)
	server, err := ln.Accept()
	noError(err)

	noError(server.Close())
	var buf [16]byte
	if _, err := conn.Read(buf[:]); err != io.EOF {
		panic("expected EOF")
	}

	conn.Close()
	ln.Close()
	println("ok")
}

func testTCP_ReadDeadline() {
	print("- TCP read deadline...")
	// Set up a connected pair, then read on the server side with no data sent.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	noError(err)
	server, err := ln.Accept()
	noError(err)

	// Nothing is written, so a read with a short deadline must time out.
	noError(server.SetReadDeadline(time.Now().Add(50 * time.Millisecond)))
	var buf [16]byte
	if _, err := server.Read(buf[:]); err != net.ErrTimeout {
		panic("expected timeout")
	}

	conn.Close()
	server.Close()
	ln.Close()
	println("ok")
}

func testTCP_ClearDeadline() {
	print("- TCP clear deadline...")
	// After a read deadline fires, clearing it must leave the connection usable.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	noError(err)
	server, err := ln.Accept()
	noError(err)

	// Arm a short deadline and let it elapse with no data.
	noError(server.SetReadDeadline(time.Now().Add(50 * time.Millisecond)))
	var buf [16]byte
	if _, err := server.Read(buf[:]); err != net.ErrTimeout {
		panic("expected timeout")
	}

	// Clearing the deadline must let a read of already-sent data succeed
	// instead of timing out. (Data is sent first because there is no second
	// thread to write during a blocking read.)
	_, err = conn.Write([]byte("hi"))
	noError(err)
	noError(server.SetReadDeadline(time.Time{}))
	n, err := server.Read(buf[:])
	noError(err)
	if string(buf[:n]) != "hi" {
		panic("read after clearing deadline failed")
	}

	conn.Close()
	server.Close()
	ln.Close()
	println("ok")
}

func testTCP_AcceptDeadline() {
	print("- TCP accept deadline...")
	// A listener with a short deadline and no incoming connection must time out.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)

	noError(ln.SetDeadline(time.Now().Add(50 * time.Millisecond)))
	if _, err := ln.Accept(); err != net.ErrTimeout {
		panic("expected timeout")
	}

	noError(ln.Close())
	println("ok")
}

func testTCP_CloseErrors() {
	print("- TCP close errors...")
	// A double close, and any I/O after close, must report ErrClosed on both
	// connections and listeners.
	lnAddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:0")
	noError(err)
	ln, err := net.ListenTCP("tcp", &lnAddr)
	noError(err)
	raddr := ln.Addr()
	conn, err := net.DialTCP("tcp", nil, &raddr)
	noError(err)
	server, err := ln.Accept()
	noError(err)

	noError(conn.Close())
	if err := conn.Close(); err != net.ErrClosed {
		panic("expected ErrClosed on double close")
	}
	var buf [16]byte
	if _, err := conn.Read(buf[:]); err != net.ErrClosed {
		panic("expected ErrClosed on read after close")
	}
	if _, err := conn.Write(buf[:]); err != net.ErrClosed {
		panic("expected ErrClosed on write after close")
	}

	noError(ln.Close())
	if err := ln.Close(); err != net.ErrClosed {
		panic("expected ErrClosed on double close (listener)")
	}
	if _, err := ln.Accept(); err != net.ErrClosed {
		panic("expected ErrClosed on accept after close")
	}

	server.Close()
	println("ok")
}
