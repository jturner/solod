package main

import (
	"pkg1"
	"pkg2"
)

func main() {
	t1 := pkg1.T1{Val: 42}
	t2 := pkg2.T2{Val: 42}
	if t1.Val != t2.Val {
		panic("t1 != t2")
	}
}
