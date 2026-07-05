package main

func regularGoto() {
	fails := 0

	for i := range 10 {
		if i%2 == 0 {
			goto next
		}
	next:
		fails++
		if fails > 2 {
			goto fallback
		}
	}

fallback:
	if fails != 3 {
		panic("fails != 3")
	}
}

func labeledLoop() {
	x := 0
outer:
	for range 5 {
		x++
	}
	if x < 10 {
		goto outer
	}
	if x != 10 {
		panic("x != 10")
	}
}

func labeledBreak() {
	sum := 0
outer:
	for i := range 5 {
		for j := range 5 {
			if i+j > 3 {
				break outer
			}
			sum += i + j
		}
	}
	if sum != 6 {
		panic("sum != 6")
	}
}

func main() {
	regularGoto()
	labeledLoop()
	labeledBreak()
}
