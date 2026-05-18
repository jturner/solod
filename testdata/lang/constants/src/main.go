package main

import "example/sub"

// File-level constants.
const fInt int = 42
const fString string = "file"

// Using _ on file level is not supported.
// var _ = fInt
// var _ = fString

// Typedefed constant group.
type HttpStatus int

const (
	StatusOK       HttpStatus = 200
	StatusNotFound HttpStatus = 404
	StatusError    HttpStatus = 500
	statusSecret   HttpStatus = 999
)

// Regular constant group.
type ServerState string

const (
	StateIdle      ServerState = "idle"
	StateConnected ServerState = "connected"
	StateError     ServerState = "error"
)

// Iota constant group.
type Day int

const (
	Sunday Day = iota
	Monday
	Tuesday
)

// Using constants in other definitions.
const Zero = 42
const FortyTwo = Zero + 42

type Point struct {
	X int
	Y int
}

var PointZero = Point{X: Zero, Y: Zero}
var PointSubZero = Point{X: sub.Zero, Y: sub.Zero}

func main() {
	{
		// Local constants.
		const lInt = 500000000
		_ = lInt
		const lFloat = 3e20 / lInt
		_ = lFloat
		const lString = "local"
		_ = lString
	}
	{
		// Using constants in expressions.
		status := StatusOK
		_ = status != StatusNotFound

		secret := statusSecret
		_ = secret > StatusOK

		state := StateConnected
		_ = state == StateIdle
	}
	{
		// Using iota constants.
		day := Monday
		_ = day == Sunday
	}
	{
		// Using _ on file level is not supported,
		// so silence the unused file-level constants here.
		_ = fInt
		_ = fString
	}
}
