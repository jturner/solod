package main

import (
	"solod.dev/so/net"
	"solod.dev/so/time"
)

func testUDP() {
	testUDP_ResolveAddr()
	testUDP_Listen()
	testUDP_Dial()
	testUDP_ReadFromWriteTo()
	testUDP_ReadDeadline()
	testUDP_CloseErrors()
}

func testUDP_ResolveAddr() {
	print("- UDP resolve a named port...")
	// A named port resolves via the udp services database (no DNS for the host).
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:domain")
	if err != nil || addr.Port != 53 {
		panic("failed to resolve named UDP port")
	}
	println("ok")

	print("- UDP resolve a hostname...")
	// "localhost" resolves via getaddrinfo (the system resolver), without
	// any external DNS. It must come back as a loopback address.
	addr, err = net.ResolveUDPAddr("udp", "localhost:53")
	noError(err)
	if addr.Port != 53 {
		panic("unexpected port")
	}
	if !addr.IP.IsLoopback() {
		panic("localhost should resolve to a loopback address")
	}
	println("ok")
}

func testUDP_Listen() {
	print("- UDP listen...")
	// Resolve an IP literal (no DNS) and listen on an OS-assigned port.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	if err != nil || laddr.Port != 0 {
		panic("failed to resolve listen address")
	}

	conn, err := net.ListenUDP("udp", &laddr)
	noError(err)
	if conn.LocalAddr().Port == 0 {
		panic("listener port not assigned")
	}
	noError(conn.Close())
	println("ok")
}

func testUDP_Dial() {
	// A single-threaded loopback echo. Datagrams are buffered in the kernel, so
	// no call blocks waiting on another thread.
	print("- UDP dial...")

	// Server listens on an OS-assigned port.
	srvAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	noError(err)
	server, err := net.ListenUDP("udp", &srvAddr)
	noError(err)

	// Client connects to the server.
	raddr := server.LocalAddr()
	client, err := net.DialUDP("udp", nil, &raddr)
	noError(err)

	// The client's remote address is the server.
	if client.RemoteAddr().Port != raddr.Port {
		panic("client remote addr mismatch")
	}

	// Client writes a datagram; the server receives it and learns the client's
	// address, then echoes it back via WriteTo.
	if _, err := client.Write([]byte("ping")); err != nil {
		panic(err)
	}

	var buf [256]byte
	r, err := server.ReadFrom(buf[:])
	noError(err)
	if r.Addr.Port != client.LocalAddr().Port {
		panic("server learned wrong client addr")
	}
	if _, err := server.WriteTo(buf[:r.N], &r.Addr); err != nil {
		panic(err)
	}

	// Client reads the echo on its connected socket.
	var got [256]byte
	n, err := client.Read(got[:])
	noError(err)
	if string(got[:n]) != "ping" {
		panic("echo mismatch")
	}

	client.Close()
	server.Close()
	println("ok")
}

func testUDP_ReadFromWriteTo() {
	print("- UDP ReadFrom/WriteTo...")
	// Two unconnected sockets exchange datagrams in both directions, with each
	// receiver checking the reported source address against the sender's local
	// address.
	addrA, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	noError(err)
	a, err := net.ListenUDP("udp", &addrA)
	noError(err)

	addrB, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	noError(err)
	b, err := net.ListenUDP("udp", &addrB)
	noError(err)

	// A -> B.
	bAddr := b.LocalAddr()
	if _, err := a.WriteTo([]byte("ping"), &bAddr); err != nil {
		panic(err)
	}
	var buf [256]byte
	r, err := b.ReadFrom(buf[:])
	noError(err)
	if string(buf[:r.N]) != "ping" {
		panic("A->B payload mismatch")
	}
	if r.Addr.Port != a.LocalAddr().Port {
		panic("A->B source addr mismatch")
	}

	// B -> A, replying to the learned source address.
	if _, err := b.WriteTo([]byte("pong"), &r.Addr); err != nil {
		panic(err)
	}
	var buf2 [256]byte
	r2, err := a.ReadFrom(buf2[:])
	noError(err)
	if string(buf2[:r2.N]) != "pong" {
		panic("B->A payload mismatch")
	}
	if r2.Addr.Port != b.LocalAddr().Port {
		panic("B->A source addr mismatch")
	}

	a.Close()
	b.Close()
	println("ok")
}

func testUDP_ReadDeadline() {
	print("- UDP read deadline...")
	// A ReadFrom with a short deadline and no data must time out.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	noError(err)
	conn, err := net.ListenUDP("udp", &laddr)
	noError(err)

	noError(conn.SetReadDeadline(time.Now().Add(50 * time.Millisecond)))
	var buf [16]byte
	if _, err := conn.ReadFrom(buf[:]); err != net.ErrTimeout {
		panic("expected timeout")
	}

	noError(conn.Close())
	println("ok")
}

func testUDP_CloseErrors() {
	print("- UDP close errors...")
	// A double close, and any I/O after close, must report ErrClosed.
	laddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	noError(err)
	conn, err := net.ListenUDP("udp", &laddr)
	noError(err)

	noError(conn.Close())
	if err := conn.Close(); err != net.ErrClosed {
		panic("expected ErrClosed on double close")
	}
	var buf [16]byte
	if _, err := conn.ReadFrom(buf[:]); err != net.ErrClosed {
		panic("expected ErrClosed on ReadFrom after close")
	}
	if _, err := conn.WriteTo(buf[:], &laddr); err != net.ErrClosed {
		panic("expected ErrClosed on WriteTo after close")
	}
	println("ok")
}
