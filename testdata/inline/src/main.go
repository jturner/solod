package main

// Rect is a rectangle.
type Rect struct {
	W int
	H int
}

// Area returns the area of the rectangle.
//
//so:inline
func (r Rect) Area() int {
	return r.W * r.H
}

// Scale scales the rectangle by a factor.
func (r *Rect) Scale(factor int) {
	r.W *= factor
	r.H *= factor
}

//so:inline
func add(a, b int) int {
	return a + b
}

func main() {
	r := Rect{W: 3, H: 4}
	_ = r.Area()
	r.Scale(2)
	_ = add(1, 2)
}
