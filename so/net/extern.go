package net

import "solod.dev/so/c"

//so:include.c <arpa/inet.h>
//so:include.c <errno.h>
//so:include.c <fcntl.h>
//so:include.c <netdb.h>
//so:include.c <netinet/in.h>
//so:include.c <poll.h>
//so:include.c <signal.h>
//so:include.c <sys/socket.h>
//so:include.c <unistd.h>

//so:embed net.h
var net_h string

//so:embed net.c
var net_c string

//so:extern
var errno int

// Errno constants relevant to network operations.
//
//so:extern EADDRINUSE
const eADDRINUSE = 0 // Address already in use
//so:extern EADDRNOTAVAIL
const eADDRNOTAVAIL = 0 // Cannot assign requested address
//so:extern ECONNABORTED
const eCONNABORTED = 0 // Software caused connection abort
//so:extern ECONNREFUSED
const eCONNREFUSED = 0 // Connection refused
//so:extern ECONNRESET
const eCONNRESET = 0 // Connection reset by peer
//so:extern EHOSTUNREACH
const eHOSTUNREACH = 0 // No route to host
//so:extern EINTR
const eINTR = 0 // Interrupted system call
//so:extern EMSGSIZE
const eMSGSIZE = 0 // Message too long
//so:extern ENETUNREACH
const eNETUNREACH = 0 // Network is unreachable
//so:extern EPIPE
const ePIPE = 0 // Broken pipe (peer closed the connection)
//so:extern ETIMEDOUT
const eTIMEDOUT = 0 // Operation timed out

// Address family constants.
//
//so:extern AF_UNSPEC
const c_AF_UNSPEC = 0 // unspecified; let getaddrinfo choose
//so:extern AF_INET
const c_AF_INET = 0 // IPv4
//so:extern AF_INET6
const c_AF_INET6 = 0 // IPv6

//so:extern SOCK_STREAM
const c_SOCK_STREAM = 0 // sequenced, reliable, two-way byte stream (TCP)
//so:extern SOCK_DGRAM
const c_SOCK_DGRAM = 0 // connectionless, unreliable datagrams (UDP)

//so:extern SOL_SOCKET
const c_SOL_SOCKET = 0 // socket-level option level for setsockopt
//so:extern SO_REUSEADDR
const c_SO_REUSEADDR = 0 // allow reuse of a local address

// poll event flags, used to wait until a socket is ready before a blocking op.
//
//so:extern POLLIN
const c_POLLIN = 0 // data (or a pending connection, or EOF) is ready to read
//so:extern POLLOUT
const c_POLLOUT = 0 // the socket is ready to accept a write

// fcntl commands and flags for setting close-on-exec.
//
//so:extern F_SETFD
const c_F_SETFD = 0 // set the file descriptor flags
//so:extern FD_CLOEXEC
const c_FD_CLOEXEC = 0 // close the descriptor on exec

// sockaddr is a generic address header.
//
//so:extern struct sockaddr
type sockaddr struct {
	sa_family uint16
}

// in_addr is the IPv4 address struct.
//
//so:extern struct in_addr
type in_addr struct {
	s_addr uint32
}

// sockaddr_in is the sockaddr structure for IPv4 addresses.
//
//so:extern struct sockaddr_in
type sockaddr_in struct {
	sin_family uint16  // = AF_INET
	sin_port   uint16  // port number
	sin_addr   in_addr // IPv4 address
}

// in6_addr is the IPv6 address struct.
//
//so:extern struct in6_addr
type in6_addr struct {
	s6_addr [16]byte
}

// sockaddr_in6 is the sockaddr structure for IPv6 addresses.
//
//so:extern struct sockaddr_in6
type sockaddr_in6 struct {
	sin6_family uint16   // = AF_INET6
	sin6_port   uint16   // port number
	sin6_addr   in6_addr // IPv6 address
}

// sockaddr_storage is a buffer big enough and suitably aligned for any [sockaddr].
// It is used opaquely: address structs are built into it or parsed out of it.
//
//so:extern sockaddr_storage
type sockaddr_storage struct{}

// pollfd describes a single descriptor to be watched by [poll].
//
//so:extern struct pollfd
type pollfd struct {
	fd      c.Int   // descriptor to watch
	events  c.Short // requested events (POLLIN, POLLOUT)
	revents c.Short // events that occurred (filled in by poll)
}

// addrinfo contains network address information.
//
//so:extern struct addrinfo
type addrinfo struct {
	ai_family   c.Int     // address family of socket
	ai_socktype c.Int     // socket type
	ai_addrlen  c.UInt    // length of socket address
	ai_addr     *sockaddr // socket address of socket
	ai_next     *addrinfo // pointer to next in list
}

// htons converts a 16-bit integer from host byte order to network byte order.
//
//so:extern htons
func htons(x uint16) uint16 { return x }

// ntohs converts a 16-bit integer from network byte order to host byte order.
//
//so:extern ntohs
func ntohs(x uint16) uint16 { return x }

// getaddrinfo translates the name of a service location (for example,
// a host name) and/or a service name (port) and returns a set of socket
// addresses and associated information to be used in creating a socket
// with which to address the specified service.
//
// servname is a *c.Char so that nil (a NULL service) can be passed; the
// port is resolved separately via [getservbyname].
//
// Returns zero success, or a non-zero error code on failure.
//
//so:extern getaddrinfo
func getaddrinfo(nodename string, servname *c.Char, hints *addrinfo, res **addrinfo) c.Int {
	_, _, _, _ = nodename, servname, hints, res
	return -1
}

// servent describes an entry in the services database.
//
//so:extern struct servent
type servent struct {
	s_port c.Int // port number, network byte order
}

// getservbyname looks up a service by name and protocol (for example "http"
// and "tcp") in the services database. Returns nil if no entry is found.
//
//so:extern getservbyname
func getservbyname(name string, proto string) *servent {
	_, _ = name, proto
	return nil
}

// freeaddrinfo frees one or more [addrinfo] structures returned by [getaddrinfo],
// along with any additional storage associated with those structures.
//
//so:extern freeaddrinfo
func freeaddrinfo(ai *addrinfo) { _ = ai }

// socket creates an unbound socket in a communications domain.
// Returns a file descriptor that can be used in later function calls
// that operate on sockets, or -1 on error.
//
//so:extern socket
func socket(domain c.Int, typ c.Int, protocol c.Int) c.Int {
	_, _, _ = domain, typ, protocol
	return -1
}

// connect attempts to make a connection on a connection-mode socket
// or to set or reset the peer address of a connectionless-mode socket.
// Returns zero on success, or -1 on error.
//
//so:extern connect
func connect(socket c.Int, address *sockaddr, address_len c.UInt) c.Int {
	_, _, _ = socket, address, address_len
	return -1
}

// bind assigns a local address to a socket.
// Returns zero on success, or -1 on error.
//
//so:extern bind
func bind(socket c.Int, address *sockaddr, address_len c.UInt) c.Int {
	_, _, _ = socket, address, address_len
	return -1
}

// listen marks a connection-mode socket as accepting connections.
// Returns zero on success, or -1 on error.
//
//so:extern listen
func listen(socket c.Int, backlog c.Int) c.Int {
	_, _ = socket, backlog
	return -1
}

// accept extracts the first connection on the queue of pending connections,
// creates a new socket with the same socket type protocol and address family
// as the specified socket, and allocates a new file descriptor for that socket.
//
// Returns the new socket's file descriptor on success, or -1 on error.
//
//so:extern accept
func accept(socket c.Int, address *sockaddr, address_len *c.UInt) c.Int {
	_, _, _ = socket, address, address_len
	return -1
}

// setsockopt sets a socket option with given name, value and protocol level.
// Returns zero on success, or -1 on error.
//
//so:extern setsockopt
func setsockopt(socket c.Int, level c.Int, option_name c.Int, option_value any, option_len c.UInt) c.Int {
	_, _, _, _, _ = socket, level, option_name, option_value, option_len
	return -1
}

// getsockname retrieves the locally-bound name of the specified socket,
// stores this address in the sockaddr structure pointed to by address,
// and stores the length of this address in the object pointed to by address_len.
//
//so:extern getsockname
func getsockname(socket c.Int, address *sockaddr, address_len *c.UInt) c.Int {
	_, _, _ = socket, address, address_len
	return -1
}

// read reads up to nbyte bytes from the file associated with the
// file descriptor fd, into the buffer pointed to by buf.
//
// Returns the number of bytes read, or -1 on error.
//
//so:extern read
func fd_read(fd c.Int, buf *byte, nbyte uintptr) int {
	_, _, _ = fd, buf, nbyte
	return 0
}

// write writes up to nbyte bytes from the buffer pointed to by buf
// to the file associated with the file descriptor fd.
//
// Returns the number of bytes written, or -1 on error.
//
//so:extern write
func fd_write(fd c.Int, buf *byte, nbyte uintptr) int {
	_, _, _ = fd, buf, nbyte
	return 0
}

// sendto sends a datagram on a socket. For a connectionless socket dest is the
// destination address; for a connected socket dest may be nil. Returns the
// number of bytes sent, or -1 on error.
//
//so:extern sendto
func sendto(socket c.Int, buf *byte, length uintptr, flags c.Int, dest *sockaddr, dest_len c.UInt) int {
	_, _, _, _, _, _ = socket, buf, length, flags, dest, dest_len
	return -1
}

// recvfrom receives a datagram from a socket and, if address is non-nil, stores
// the source address there. One whole datagram is returned; any excess beyond
// length is discarded. Returns the number of bytes received, or -1 on error.
//
//so:extern recvfrom
func recvfrom(socket c.Int, buf *byte, length uintptr, flags c.Int, address *sockaddr, address_len *c.UInt) int {
	_, _, _, _, _, _ = socket, buf, length, flags, address, address_len
	return -1
}

// close closes the file descriptor indicated by fd.
// Returns zero on success, or -1 on error.
//
//so:extern close
func fd_close(fd c.Int) c.Int {
	_ = fd
	return 0
}

// fcntl performs an operation on the file descriptor fd. The C function is
// variadic; the single int arg covers the commands this package uses
// (F_SETFD). Returns -1 on error.
//
//so:extern fcntl
func fcntl(fd c.Int, cmd c.Int, arg c.Int) c.Int {
	_, _, _ = fd, cmd, arg
	return -1
}

// poll waits for one of a set of file descriptors to become ready to perform
// I/O. timeout is in milliseconds; a negative value blocks indefinitely.
// Returns the number of ready descriptors, 0 on timeout, or -1 on error.
//
//so:extern poll
func poll(fds *pollfd, nfds c.ULong, timeout c.Int) c.Int {
	_, _, _ = fds, nfds, timeout
	return -1
}
