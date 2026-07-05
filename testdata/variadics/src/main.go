package main

type Sum struct {
	v int
}

func (s *Sum) Add(nums ...int) {
	for _, num := range nums {
		s.v += num
	}
}

func sum(nums ...int) int {
	total := 0
	for _, num := range nums {
		total += num
	}
	return total
}

func main() {
	{
		// Variadic function call.
		sum(1, 2)
		total := sum(1, 2, 3)
		if total != 6 {
			panic("wrong sum")
		}

		nums := []int{1, 2, 3, 4}
		total = sum(nums...)
		if total != 10 {
			panic("wrong sum")
		}
	}
	{
		// Variadic method call.
		var s Sum
		s.Add(1, 2)
		s.Add(1, 2, 3)
		if s.v != 9 {
			panic("wrong sum")
		}

		nums := []int{1, 2, 3, 4}
		s.Add(nums...)
		if s.v != 19 {
			panic("wrong sum")
		}
	}
}
