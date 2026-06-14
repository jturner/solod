package net_test

import "solod.dev/so/net"

// Connect to a TCP server, send a request, and read the reply.
func ExampleDialTCP() {
	// Resolve the server address. An IP literal needs no DNS lookup.
	raddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	// Connect. A nil laddr lets the system choose the local address.
	conn, err := net.DialTCP("tcp", nil, &raddr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	// Send a request.
	if _, err := conn.Write([]byte("ping")); err != nil {
		panic(err)
	}

	// Read the reply.
	var buf [256]byte
	n, err := conn.Read(buf[:])
	if err != nil {
		panic(err)
	}
	println(string(buf[:n]))
}

// Announce on a local TCP address and serve connections one
// at a time, echoing back whatever each client sends. This package
// has no goroutines, so connections are handled sequentially.
func ExampleListenTCP() {
	// Resolve the local address to listen on (IP literal, no DNS).
	laddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}

	ln, err := net.ListenTCP("tcp", &laddr)
	if err != nil {
		panic(err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			panic(err)
		}

		// Echo one message back to the client, then close the connection.
		var buf [256]byte
		n, err := conn.Read(buf[:])
		if err == nil {
			conn.Write(buf[:n])
		}
		conn.Close()
	}
}
