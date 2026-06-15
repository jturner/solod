package main

import "solod.dev/so/net"

func main() {
	println("solod.dev/so/net")
	testSplitHostPort()
	testJoinHostPort()
	testTCP()
	testUDP()
}

func testSplitHostPort() {
	print("- split host-port...")
	hp, err := net.SplitHostPort("127.0.0.1:8080")
	if err != nil || hp.Host != "127.0.0.1" || hp.Port != "8080" {
		panic("unexpected SplitHostPort result")
	}
	println("ok")
}

func testJoinHostPort() {
	print("- join host-port...")
	var buf [64]byte
	if net.JoinHostPort(buf[:], "::1", "80") != "[::1]:80" {
		panic("unexpected JoinHostPort result")
	}
	println("ok")
}

func noError(err error) {
	if err != nil {
		panic(err)
	}
}
