typedef struct {
    void (*Write)(const char* format, ...);
} Stream;

static inline void Discard(const char* format, ...) {
    (void)format;
}
