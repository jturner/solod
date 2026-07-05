package main

func main() {
	{
		// Empty switch statement.
		switch {
		}
	}
	{
		// Switch on int with cases and default.
		i := 2
		switch i {
		case 1:
			panic("unexpected i == 1")
		case 2:
			println("i == 2")
		case 3:
			panic("unexpected i == 3")
		default:
			panic("unexpected default")
		}
	}
	{
		// Tagless switch (bool conditions).
		x := 10
		switch {
		case x > 100:
			panic("unexpected x > 100")
		case x > 0:
			println("x > 0")
		default:
			panic("unexpected default")
		}
	}
	{
		// Multiple values per case.
		y := 3
		switch y {
		case 1, 2, 3:
			println("y == 3")
		case 4, 5, 6:
			panic("unexpected y == 4, 5, 6")
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch with init statement.
		switch n := 42; n {
		case 42:
			println("n == 42")
		default:
			panic("unexpected default")
		}
	}
	{
		// Switch on string.
		s := "hello"
		switch s {
		case "hello":
			println("s == hello")
		case "bye":
			panic("unexpected s == bye")
		default:
			panic("unexpected default")
		}
	}
	{
		// Cases without default.
		z := 5
		switch z {
		case 1:
			panic("unexpected z == 1")
		case 5:
			println("z == 5")
		}
	}
}
