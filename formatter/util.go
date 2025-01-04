package formatter

type intStack struct {
	ints []int
}

func (t *intStack) Push(i int) {
	t.ints = append(t.ints, i)
}

func (t *intStack) Pop() int {
	ret := 0
	if len(t.ints) > 0 {
		n := len(t.ints) - 1
		ret = t.ints[n]
		switch n {
		case 0:
			t.ints = nil
		default:
			t.ints[n] = 0
			t.ints = t.ints[:n]
		}
	}
	return ret
}

func (t *intStack) Peek() int {
	if len(t.ints) > 0 {
		return t.ints[len(t.ints)-1]
	}
	return 0
}

func (t *intStack) Len() int {
	return len(t.ints)
}

func (t *intStack) Reset() {
	t.ints = nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
