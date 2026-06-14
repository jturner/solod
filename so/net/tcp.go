package net

import (
	"solod.dev/so/c"
	"solod.dev/so/io"
	"solod.dev/so/mem"
	"solod.dev/so/net/netip"
	"solod.dev/so/strconv"
	"solod.dev/so/time"
)

// listenBacklog is the default connection backlog for listeners.
const listenBacklog = 128

// TCPAddr represents the address of a TCP endpoint.
type TCPAddr struct {
	IP   netip.Addr
	Port int
}

// Network returns the address's network name, "tcp".
func (TCPAddr) Network() string {
	return "tcp"
}

// String returns the address in "host:port" form (or "[host]:port" for IPv6),
// built into buf. buf must have at least netip.MaxAddrPortLen bytes of
// capacity; the returned string aliases buf.
func (a TCPAddr) String(buf []byte) string {
	ap := netip.AddrPortFrom(a.IP, uint16(a.Port))
	return ap.String(buf)
}

// family returns the address family (AF_INET or AF_INET6)
// for the address's IP. Returns AF_INET for an invalid IP.
func (a TCPAddr) family() c.Int {
	if a.IP.Is6() {
		return c_AF_INET6
	}
	return c_AF_INET
}

// ResolveTCPAddr returns the address of a TCP endpoint.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
//
// The address has the form "host:port". An empty host means the unspecified
// address (0.0.0.0 or ::). An IP literal host is parsed directly; otherwise it
// is resolved via the system resolver, and a host name with several addresses
// resolves to its first. The port may be a decimal number or a service name
// (for example "http").
//
// Examples:
//
//	ResolveTCPAddr("tcp", "golang.org:443")
//	ResolveTCPAddr("tcp", "192.0.2.1:80")
//	ResolveTCPAddr("tcp", "localhost:http")
//	ResolveTCPAddr("tcp", ":80")
func ResolveTCPAddr(network, address string) (TCPAddr, error) {
	var addrs [1]TCPAddr
	if _, err := resolveTCPAddrs(network, address, addrs[:]); err != nil {
		return TCPAddr{}, err
	}
	return addrs[0], nil
}

// resolveTCPAddrs resolves address into TCP endpoints, storing up to len(dst)
// of them in dst and returning how many were stored (at least one on success).
// An empty host or IP literal yields a single address; a host name may yield
// several, in the order they should be tried.
func resolveTCPAddrs(network, address string, dst []TCPAddr) (int, error) {
	family := familyFor(network)
	if family == afInvalid {
		return 0, ErrUnknownNetwork
	}
	hp, err := SplitHostPort(address)
	if err != nil {
		return 0, err
	}
	port, ok := lookupPort(hp.Port)
	if !ok {
		return 0, ErrInvalidPort
	}

	// Empty host: the unspecified address for the family.
	if len(hp.Host) == 0 {
		dst[0] = unspecifiedAddr(family, port)
		return 1, nil
	}

	// IP literal: no resolution needed, but it must match the network's family
	// ("tcp4" rejects an IPv6 literal and vice versa).
	if ip, perr := netip.ParseAddr(hp.Host); perr == nil {
		if !familyMatch(family, ip) {
			return 0, ErrNoSuitableAddr
		}
		dst[0] = TCPAddr{IP: ip, Port: port}
		return 1, nil
	}

	// Host name: resolve via getaddrinfo and decode each result.
	hints := addrinfo{ai_family: family, ai_socktype: c_SOCK_STREAM}
	var ai *addrinfo
	if getaddrinfo(hp.Host, nil, &hints, &ai) != 0 || ai == nil {
		return 0, ErrNoSuchHost
	}
	n := 0
	for p := ai; p != nil && n < len(dst); p = p.ai_next {
		var stor sockaddr_storage
		mem.Copy(&stor, p.ai_addr, int(p.ai_addrlen))
		addr := stor.tcpAddr()
		if !addr.IP.IsValid() {
			continue
		}
		addr.Port = port
		dst[n] = addr
		n++
	}
	freeaddrinfo(ai)
	if n == 0 {
		return 0, ErrNoSuchHost
	}
	return n, nil
}

// lookupPort resolves a port string to a numeric port. An empty string means
// port 0. Otherwise the string may be a decimal number or a service name from
// the services database (for example "http"). Returns the port and true on
// success, or false if the number is out of range or the service name is unknown.
func lookupPort(port string) (int, bool) {
	if port == "" {
		return 0, true
	}
	n, err := strconv.Atoi(port)
	if err == nil {
		return n, n >= 0 && n <= 65535
	}
	se := getservbyname(port, "tcp")
	if se == nil {
		return 0, false
	}
	return int(ntohs(uint16(se.s_port))), true
}

// unspecifiedAddr returns the unspecified address (0.0.0.0 or ::) for the
// family, with the given port. It is used when the host part is empty.
func unspecifiedAddr(family c.Int, port int) TCPAddr {
	ip := netip.IPv4Unspecified()
	if family == c_AF_INET6 {
		ip = netip.IPv6Unspecified()
	}
	return TCPAddr{IP: ip, Port: port}
}

// TCPConn abstracts a TCP network connection.
//
// The zero value is not usable; obtain a TCPConn from [DialTCP] or
// [TCPListener.Accept]. A TCPConn must not be copied after use
// (copies share the underlying socket descriptor).
type TCPConn struct {
	fd     c.Int
	laddr  TCPAddr
	raddr  TCPAddr
	closed bool
	// Read/write deadlines; the zero Time means no deadline (block forever).
	rdeadline time.Time
	wdeadline time.Time
}

// DialTCP connects to raddr on the named TCP network.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
// Use [ResolveTCPAddr] to obtain raddr (and an optional laddr) from a
// "host:port" string.
//
// If laddr is nil, a local address is automatically chosen.
// A laddr with an invalid IP binds only its port, on the
// unspecified address of the remote's family.
func DialTCP(network string, laddr, raddr *TCPAddr) (TCPConn, error) {
	if familyFor(network) == afInvalid {
		return TCPConn{}, ErrUnknownNetwork
	}
	if raddr == nil {
		return TCPConn{}, ErrAddrNotAvail
	}

	var rstor sockaddr_storage
	rlen := rstor.fill(*raddr)
	if rlen == 0 {
		return TCPConn{}, ErrAddrNotAvail
	}

	fd := socket(raddr.family(), c_SOCK_STREAM, 0)
	if fd < 0 {
		return TCPConn{}, mapError()
	}
	closeOnExec(fd)

	// Optional local bind address (bind-before-connect).
	if laddr != nil {
		local := *laddr
		if !local.IP.IsValid() {
			// Bind only the port, on the unspecified address of the remote's family.
			local.IP = unspecifiedAddr(raddr.family(), 0).IP
		}
		var lstor sockaddr_storage
		llen := lstor.fill(local)
		if llen == 0 || bind(fd, lstor.sockAddr(), llen) != 0 {
			err := mapError()
			fd_close(fd)
			return TCPConn{}, err
		}
	}

	// Unlike read/write/accept, connect is not restarted on EINTR: a blocking
	// connect interrupted by a signal keeps completing in the background, so
	// re-calling it would return EALREADY/EISCONN. Doing it right needs
	// poll(POLLOUT) plus getsockopt(SO_ERROR); not worth it for this rare case.
	if connect(fd, rstor.sockAddr(), rlen) != 0 {
		err := mapError()
		fd_close(fd)
		return TCPConn{}, err
	}

	conn := TCPConn{fd: fd, raddr: *raddr}
	conn.laddr = sockname(fd)
	return conn, nil
}

// Read reads data from the connection into b.
// At end of stream, Read returns 0, io.EOF.
func (conn *TCPConn) Read(b []byte) (int, error) {
	if conn.closed {
		return 0, ErrClosed
	}
	if len(b) == 0 {
		return 0, nil
	}
	// Restart on EINTR: a read interrupted by a signal before any data was
	// transferred returns -1/EINTR, and is retried transparently.
	for {
		if err := waitFD(conn.fd, c_POLLIN, conn.rdeadline); err != nil {
			return 0, err
		}
		n := fd_read(conn.fd, &b[0], uintptr(len(b)))
		if n > 0 {
			return n, nil
		}
		if n == 0 {
			return 0, io.EOF
		}
		if errno != eINTR {
			return 0, mapError()
		}
	}
}

// Write writes len(b) bytes from b to the connection.
// Returns the number of bytes written and an error, if any.
func (conn *TCPConn) Write(b []byte) (int, error) {
	if conn.closed {
		return 0, ErrClosed
	}
	// Loop until all bytes are written: a single write may transfer fewer
	// bytes than requested when the socket send buffer fills up. A write
	// interrupted by a signal before any data was transferred returns
	// -1/EINTR and is restarted.
	total := 0
	for total < len(b) {
		if err := waitFD(conn.fd, c_POLLOUT, conn.wdeadline); err != nil {
			return total, err
		}
		n := fd_write(conn.fd, &b[total], uintptr(len(b)-total))
		if n < 0 {
			if errno == eINTR {
				continue
			}
			return total, mapError()
		}
		total += n
	}
	return total, nil
}

// Close closes the connection. Returns an error
// if it has already been called.
func (conn *TCPConn) Close() error {
	if conn.closed {
		return ErrClosed
	}
	conn.closed = true
	if fd_close(conn.fd) != 0 {
		return mapError()
	}
	return nil
}

// LocalAddr returns the local network address.
func (conn *TCPConn) LocalAddr() TCPAddr {
	return conn.laddr
}

// RemoteAddr returns the remote network address.
func (conn *TCPConn) RemoteAddr() TCPAddr {
	return conn.raddr
}

// SetDeadline sets the read and write deadlines associated
// with the connection. It is equivalent to calling both
// SetReadDeadline and SetWriteDeadline.
//
// A deadline is an absolute time after which I/O operations
// fail instead of blocking. The deadline applies to all future I/O,
// not just the immediately following call to Read or Write.
// After a deadline has been exceeded, the connection can be
// refreshed by setting a deadline in the future.
//
// If the deadline is exceeded a call to Read or Write or to other
// I/O methods will return [ErrTimeout].
//
// An idle timeout can be implemented by repeatedly extending
// the deadline after successful Read or Write calls.
//
// A zero value for t means I/O operations will not time out.
func (conn *TCPConn) SetDeadline(t time.Time) error {
	if conn.closed {
		return ErrClosed
	}
	conn.rdeadline = t
	conn.wdeadline = t
	return nil
}

// SetReadDeadline sets the deadline for future Read calls.
// A zero value for t means Read will not time out.
func (conn *TCPConn) SetReadDeadline(t time.Time) error {
	if conn.closed {
		return ErrClosed
	}
	conn.rdeadline = t
	return nil
}

// SetWriteDeadline sets the deadline for future Write calls.
// A zero value for t means Write will not time out.
func (conn *TCPConn) SetWriteDeadline(t time.Time) error {
	if conn.closed {
		return ErrClosed
	}
	conn.wdeadline = t
	return nil
}

// TCPListener is a TCP network listener. The zero value
// is not usable; obtain one from [ListenTCP].
type TCPListener struct {
	fd     c.Int
	addr   TCPAddr
	closed bool
	// Accept deadline; the zero Time means no deadline (block forever).
	deadline time.Time
}

// ListenTCP announces on the local TCP address laddr.
//
// Known networks are "tcp", "tcp4" (IPv4-only), and "tcp6" (IPv6-only).
// Use [ResolveTCPAddr] to obtain laddr from a "host:port" string.
//
// A nil laddr, or an unspecified IP in laddr (0.0.0.0 or ::, as produced by an
// empty host), listens on all interfaces of the network's address family:
// 0.0.0.0 (all IPv4 interfaces) for "tcp" or "tcp4", :: (all IPv6 interfaces)
// for "tcp6". A zero Port lets the system pick a free port, which the returned
// listener's [TCPListener.Addr] reports.
func ListenTCP(network string, laddr *TCPAddr) (TCPListener, error) {
	family := familyFor(network)
	if family == afInvalid {
		return TCPListener{}, ErrUnknownNetwork
	}

	// A nil laddr, or one with an invalid IP, binds the unspecified address
	// (all interfaces) for the network's family, keeping any requested port.
	addr := unspecifiedAddr(family, 0)
	if laddr != nil {
		addr.Port = laddr.Port
		if laddr.IP.IsValid() {
			addr.IP = laddr.IP
		}
	}

	var stor sockaddr_storage
	slen := stor.fill(addr)
	if slen == 0 {
		return TCPListener{}, ErrAddrNotAvail
	}

	fd := socket(addr.family(), c_SOCK_STREAM, 0)
	if fd < 0 {
		return TCPListener{}, mapError()
	}
	closeOnExec(fd)

	one := c.Int(1)
	setsockopt(fd, c_SOL_SOCKET, c_SO_REUSEADDR, &one, c.UInt(c.Sizeof[c.Int]()))
	if bind(fd, stor.sockAddr(), slen) != 0 || listen(fd, listenBacklog) != 0 {
		err := mapError()
		fd_close(fd)
		return TCPListener{}, err
	}

	// Report the bound address; with port 0 the system assigns the real port.
	return TCPListener{fd: fd, addr: sockname(fd)}, nil
}

// Accept waits for and returns the next connection to the listener.
func (l *TCPListener) Accept() (TCPConn, error) {
	if l.closed {
		return TCPConn{}, ErrClosed
	}
	// Restart on EINTR: an accept interrupted by a signal returns -1/EINTR,
	// and is retried transparently. slen is in/out, so reset it each try.
	var stor sockaddr_storage
	for {
		if err := waitFD(l.fd, c_POLLIN, l.deadline); err != nil {
			return TCPConn{}, err
		}
		slen := c.UInt(c.Sizeof[sockaddr_storage]())
		fd := accept(l.fd, stor.sockAddr(), &slen)
		if fd >= 0 {
			closeOnExec(fd)
			// Report the accepted socket's real local address; for a wildcard
			// listener this is the concrete interface, not l.addr's 0.0.0.0/::.
			conn := TCPConn{fd: fd, laddr: sockname(fd)}
			conn.raddr = stor.tcpAddr()
			return conn, nil
		}
		if errno != eINTR {
			return TCPConn{}, mapError()
		}
	}
}

// Close stops listening on the TCP address.
// Already accepted connections are not closed.
func (l *TCPListener) Close() error {
	if l.closed {
		return ErrClosed
	}
	l.closed = true
	if fd_close(l.fd) != 0 {
		return mapError()
	}
	return nil
}

// Addr returns the listener's network address.
func (l *TCPListener) Addr() TCPAddr {
	return l.addr
}

// SetDeadline sets the deadline for future Accept calls. An Accept that has no
// connection ready before t fails with [ErrTimeout]. The zero value for t
// clears the deadline (Accept blocks until a connection arrives).
func (l *TCPListener) SetDeadline(t time.Time) error {
	if l.closed {
		return ErrClosed
	}
	l.deadline = t
	return nil
}

// sockname returns the local address of fd, or the zero TCPAddr on error.
func sockname(fd c.Int) TCPAddr {
	var stor sockaddr_storage
	slen := c.UInt(c.Sizeof[sockaddr_storage]())
	if getsockname(fd, stor.sockAddr(), &slen) != 0 {
		return TCPAddr{}
	}
	return stor.tcpAddr()
}

// waitFD blocks until fd is ready for the requested events (c_POLLIN for a
// read or accept, c_POLLOUT for a write) or the deadline passes, in which case
// it returns ErrTimeout. A zero deadline means no deadline: waitFD returns nil
// immediately and lets the following syscall block indefinitely.
func waitFD(fd c.Int, events c.Short, deadline time.Time) error {
	if deadline.IsZero() {
		return nil
	}
	for {
		d := time.Until(deadline)
		if d <= 0 {
			return ErrTimeout
		}
		// poll is used rather than SO_RCVTIMEO/SO_SNDTIMEO because, although those
		// timeouts bound read and write on every platform, accept() ignores them on
		// macOS/BSD, so a listener deadline there would never fire. poll bounds all
		// three uniformly.
		pfd := pollfd{fd: fd, events: events}
		n := poll(&pfd, 1, pollTimeout(d))
		if n > 0 {
			return nil
		}
		// n == 0 means this poll slice elapsed: loop to recheck the deadline
		// (it may not have arrived yet if pollTimeout clamped a long wait).
		// EINTR likewise retries. Any other error is reported.
		if n < 0 && errno != eINTR {
			return mapError()
		}
	}
}

// pollTimeout converts a positive duration to a poll timeout in milliseconds,
// rounding sub-millisecond values up to 1 (never 0, which would make poll
// return immediately) and clamping to the int range so a very distant deadline
// cannot overflow into a negative (block-forever) value.
func pollTimeout(d time.Duration) c.Int {
	ms := d.Milliseconds()
	if ms < 1 {
		return 1
	}
	const maxMS = 1<<31 - 1
	if ms > maxMS {
		return maxMS
	}
	return c.Int(ms)
}

// closeOnExec sets the close-on-exec flag on fd so the socket is not inherited
// by child processes spawned via exec. It is best-effort: errors are ignored,
// and there is a small window between creating the fd and this call in which a
// concurrent exec could still inherit it (the atomic SOCK_CLOEXEC/accept4 forms
// are Linux-only, so this portable fallback is used everywhere).
func closeOnExec(fd c.Int) {
	fcntl(fd, c_F_SETFD, c_FD_CLOEXEC)
}

// afInvalid is returned by familyFor for an unsupported network. Real address
// families are non-negative, so -1 is a safe sentinel.
const afInvalid = -1

// familyFor maps a network name to an address family, or afInvalid if the
// network is not a supported TCP network ("tcp", "tcp4", "tcp6").
func familyFor(network string) c.Int {
	if network == "tcp" {
		return c_AF_UNSPEC
	}
	if network == "tcp4" {
		return c_AF_INET
	}
	if network == "tcp6" {
		return c_AF_INET6
	}
	return afInvalid
}

// familyMatch reports whether ip is usable on the given address family.
// AF_UNSPEC (the "tcp" network) accepts any IP; AF_INET requires an IPv4
// address and AF_INET6 an IPv6 one.
func familyMatch(family c.Int, ip netip.Addr) bool {
	if family == c_AF_INET {
		return ip.Is4()
	}
	if family == c_AF_INET6 {
		return ip.Is6()
	}
	return true
}
