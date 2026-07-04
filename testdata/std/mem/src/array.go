package main

import (
	"solod.dev/so/c"
	"solod.dev/so/mem"
)

func arrayTest() {
	arr := mem.NewArray(mem.System, c.Sizeof[Point](), 3)
	if arr.Len() != 3 {
		panic("want arr.Len() == 3")
	}

	var p Point
	arr.Load(1, &p)
	if p.x != 0 || p.y != 0 {
		panic("want arr[1] == {0, 0}")
	}

	p1 := Point{x: 11, y: 22}
	p2 := Point{x: 33, y: 44}
	p3 := Point{x: 55, y: 66}
	arr.Store(0, &p1)
	arr.Store(1, &p2)
	arr.Store(2, &p3)

	arr.Load(0, &p)
	if p.x != 11 || p.y != 22 {
		panic("want arr[0] == {11, 22}")
	}

	arr.Load(1, &p)
	if p.x != 33 || p.y != 44 {
		panic("want arr[1] == {33, 44}")
	}

	// At returns a pointer into the storage.
	pp := arr.At(2).(*Point)
	if pp.x != 55 || pp.y != 66 {
		panic("want arr[2] == {55, 66}")
	}

	arr.Free()
}
