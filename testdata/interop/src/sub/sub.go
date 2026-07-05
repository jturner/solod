package sub

//so:embed sub.h
var sub_h string

//so:extern Stream
type Stream struct {
	Write func(format string, args ...any)
}

//so:extern Discard
func Discard(format string, args ...any) {}
