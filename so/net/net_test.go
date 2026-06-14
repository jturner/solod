package net

import (
	"testing"

	"solod.dev/so/net/netip"
)

// TestResolveTCPAddr covers the parts of ResolveTCPAddr that complete without a
// syscall: network validation, port parsing, and the IP literal path. The
// host-name (getaddrinfo) path needs a real resolver, and family matching
// ("tcp4" vs "tcp6") needs the AF_* externs, which are all 0 on the host; both
// are exercised by the std/net integration test instead.
func TestResolveTCPAddr(t *testing.T) {
	// IP literal: resolved directly, no DNS.
	addr, err := ResolveTCPAddr("tcp", "127.0.0.1:80")
	if err != nil {
		t.Fatalf("ResolveTCPAddr literal: %v", err)
	}
	if addr.Port != 80 || !addr.IP.Equal(netip.MustParseAddr("127.0.0.1")) {
		var buf [netip.MaxAddrPortLen]byte
		t.Errorf("ResolveTCPAddr literal = %q; want 127.0.0.1:80", addr.String(buf[:]))
	}

	// Empty host: the unspecified address for the family.
	addr, err = ResolveTCPAddr("tcp", ":80")
	if err != nil {
		t.Fatalf("ResolveTCPAddr empty host: %v", err)
	}
	if addr.Port != 80 || !addr.IP.IsUnspecified() {
		var buf [netip.MaxAddrPortLen]byte
		t.Errorf("ResolveTCPAddr empty host = %q; want 0.0.0.0:80", addr.String(buf[:]))
	}

	for _, tt := range []struct {
		network string
		address string
		err     error
	}{
		{"udp", "127.0.0.1:80", ErrUnknownNetwork},
		{"tcp", "127.0.0.1", ErrMissingPort},
		{"tcp", "127.0.0.1:99999", ErrInvalidPort},
	} {
		if _, err := ResolveTCPAddr(tt.network, tt.address); err != tt.err {
			t.Errorf("ResolveTCPAddr(%q, %q) = %v; want %v", tt.network, tt.address, err, tt.err)
		}
	}
}

func TestSplitHostPort(t *testing.T) {
	for _, tt := range []struct {
		hostPort string
		host     string
		port     string
	}{
		// Host name
		{"localhost:http", "localhost", "http"},
		{"localhost:80", "localhost", "80"},

		// Go-specific host name with zone identifier
		{"localhost%lo0:http", "localhost%lo0", "http"},
		{"localhost%lo0:80", "localhost%lo0", "80"},
		{"[localhost%lo0]:http", "localhost%lo0", "http"}, // Go 1 behavior
		{"[localhost%lo0]:80", "localhost%lo0", "80"},     // Go 1 behavior

		// IP literal
		{"127.0.0.1:http", "127.0.0.1", "http"},
		{"127.0.0.1:80", "127.0.0.1", "80"},
		{"[::1]:http", "::1", "http"},
		{"[::1]:80", "::1", "80"},

		// IP literal with zone identifier
		{"[::1%lo0]:http", "::1%lo0", "http"},
		{"[::1%lo0]:80", "::1%lo0", "80"},

		// Go-specific wildcard for host name
		{":http", "", "http"}, // Go 1 behavior
		{":80", "", "80"},     // Go 1 behavior

		// Go-specific wildcard for service name or transport port number
		{"golang.org:", "golang.org", ""}, // Go 1 behavior
		{"127.0.0.1:", "127.0.0.1", ""},   // Go 1 behavior
		{"[::1]:", "::1", ""},             // Go 1 behavior

		// Opaque service name
		{"golang.org:https%foo", "golang.org", "https%foo"}, // Go 1 behavior
	} {
		if hp, err := SplitHostPort(tt.hostPort); hp.Host != tt.host || hp.Port != tt.port || err != nil {
			t.Errorf("SplitHostPort(%q) = %q, %q, %v; want %q, %q, nil", tt.hostPort, hp.Host, hp.Port, err, tt.host, tt.port)
		}
	}

	for _, tt := range []struct {
		hostPort string
		err      error
	}{
		{"golang.org", ErrMissingPort},
		{"127.0.0.1", ErrMissingPort},
		{"[::1]", ErrMissingPort},
		{"[fe80::1%lo0]", ErrMissingPort},
		{"[localhost%lo0]", ErrMissingPort},
		{"localhost%lo0", ErrMissingPort},

		{"::1", ErrTooManyColons},
		{"fe80::1%lo0", ErrTooManyColons},
		{"fe80::1%lo0:80", ErrTooManyColons},

		// Test cases that didn't fail in Go 1

		{"[foo:bar]", ErrMissingPort},
		{"[foo:bar]baz", ErrMissingPort},
		{"[foo]bar:baz", ErrMissingPort},

		{"[foo]:[bar]:baz", ErrTooManyColons},

		{"[foo]:[bar]baz", ErrUnexpectedBracket},
		{"foo[bar]:baz", ErrUnexpectedBracket},

		{"foo]bar:baz", ErrUnexpectedBracket},
	} {
		if hp, err := SplitHostPort(tt.hostPort); err == nil {
			t.Errorf("SplitHostPort(%q) should have failed", tt.hostPort)
		} else {
			if err != tt.err {
				t.Errorf("SplitHostPort(%q) = _, _, %q; want %q", tt.hostPort, err, tt.err)
			}
			if hp.Host != "" || hp.Port != "" {
				t.Errorf("SplitHostPort(%q) = %q, %q, err; want %q, %q, err on failure", tt.hostPort, hp.Host, hp.Port, "", "")
			}
		}
	}
}

func TestJoinHostPort(t *testing.T) {
	for _, tt := range []struct {
		host     string
		port     string
		hostPort string
	}{
		// Host name
		{"localhost", "http", "localhost:http"},
		{"localhost", "80", "localhost:80"},

		// Go-specific host name with zone identifier
		{"localhost%lo0", "http", "localhost%lo0:http"},
		{"localhost%lo0", "80", "localhost%lo0:80"},

		// IP literal
		{"127.0.0.1", "http", "127.0.0.1:http"},
		{"127.0.0.1", "80", "127.0.0.1:80"},
		{"::1", "http", "[::1]:http"},
		{"::1", "80", "[::1]:80"},

		// IP literal with zone identifier
		{"::1%lo0", "http", "[::1%lo0]:http"},
		{"::1%lo0", "80", "[::1%lo0]:80"},

		// Go-specific wildcard for host name
		{"", "http", ":http"}, // Go 1 behavior
		{"", "80", ":80"},     // Go 1 behavior

		// Go-specific wildcard for service name or transport port number
		{"golang.org", "", "golang.org:"}, // Go 1 behavior
		{"127.0.0.1", "", "127.0.0.1:"},   // Go 1 behavior
		{"::1", "", "[::1]:"},             // Go 1 behavior

		// Opaque service name
		{"golang.org", "https%foo", "golang.org:https%foo"}, // Go 1 behavior
	} {
		var buf [64]byte
		if hostPort := JoinHostPort(buf[:], tt.host, tt.port); hostPort != tt.hostPort {
			t.Errorf("JoinHostPort(%q, %q) = %q; want %q", tt.host, tt.port, hostPort, tt.hostPort)
		}
	}
}
