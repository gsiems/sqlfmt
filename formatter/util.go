package formatter

import "log"

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

func carp(verbose bool, msg string) {
	if verbose {
		log.Print(msg)
	}
}

func getBId(bagIds map[int]int, parensDepth int) int {

	// Get the most current bag ID, if needed/available
	// Problem: There won't be a valid bag ID for all parensDepths and
	// increasing the parensDepth doesn't signify that a new bag is needed.
	// So what IS needed is to backtrack up from the parensDepth until a valid
	// bagId is found.
	// This requires that the bagId entries be cleared up as the parensDepth is
	// decreased or when a bag is closed.

	pd := parensDepth
	testId := 0
	for pd >= 0 && testId == 0 {
		if bi, ok := bagIds[pd]; ok {
			testId = bi
		}
		pd--
	}
	return testId
}
