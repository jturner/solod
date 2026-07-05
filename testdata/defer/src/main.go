package main

var state int = 0

func xopen(x *int) {
	(*x)++
}

func xclose(a any) {
	x := a.(*int)
	(*x)--
}

func funcScope() {
	xopen(&state)
	defer xclose(&state)
	if state != 1 {
		panic("unexpected state")
	}
}

func funcWithReturn() int {
	xopen(&state)
	defer xclose(&state)
	if state != 1 {
		panic("unexpected state")
	}
	return 42
}

func funcReturnCall() (int, error) {
	xopen(&state)
	defer xclose(&state)
	return funcCalc()
}

func funcReturnVar() int {
	xopen(&state)
	defer xclose(&state)
	return state
}

func funcCalc() (int, error) {
	if state != 1 {
		panic("unexpected state")
	}
	return 42, nil
}

func main() {
	funcScope()
	if state != 0 {
		panic("unexpected state")
	}
	funcWithReturn()
	if state != 0 {
		panic("unexpected state")
	}
	funcReturnCall()
	if state != 0 {
		panic("unexpected state")
	}
	if funcReturnVar() != 1 {
		panic("unexpected return value")
	}
	if state != 0 {
		panic("unexpected state")
	}
}
