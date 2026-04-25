package bench

func sumIndex(xs []int) int {
	s := 0
	for i := 0; i < len(xs); i++ {
		s += xs[i]
	}
	return s
}

func sumRange(xs []int) int {
	s := 0
	for _, v := range xs {
		s += v
	}
	return s
}

func sumUnroll(xs []int) int {
	s := 0
	i := 0
	for ; i+4 <= len(xs); i += 4 {
		s += xs[i] + xs[i+1] + xs[i+2] + xs[i+3]
	}
	for ; i < len(xs); i++ {
		s += xs[i]
	}
	return s
}
