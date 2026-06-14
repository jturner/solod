//go:build ignore

// GCC's -fanalyzer fd-ownership checker reports false leaks here. A socket fd
// reaches the caller nested inside the (T, error) Result struct returned by
// Accept/Dial, and is closed through Close's void* receiver. The analyzer
// cannot track the fd across either hop, so it reports a leak even though every
// path does close the fd. Suppress just this check for the net package.
#if defined(__clang__)
#elif defined(__GNUC__)
#pragma GCC diagnostic ignored "-Wanalyzer-fd-leak"
#endif

typedef struct sockaddr_storage sockaddr_storage;

// Ignore SIGPIPE process-wide. By default, writing to a socket whose peer has
// closed its end delivers SIGPIPE, which terminates the process. Ignoring it
// makes write() fail with EPIPE instead, which Write turns into an error.
static void __attribute__((constructor)) net_ignore_sigpipe(void) {
    signal(SIGPIPE, SIG_IGN);
}
