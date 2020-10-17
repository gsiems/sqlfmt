package main

const (
	noNewLine = iota
	newLineAllowed
	newLineRequired
)

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func chkCommentNL(n, prev wu, nlChk int) int {

	// if a new line is already required then a new line is still required
	if nlChk == newLineRequired {
		return newLineRequired
	}

	// if prev was a line comment then require a newline/indent before
	// if prev was a block comment then allow for a newline/indent before
	// if current is any kind of comment then allow for a newline/indent before
	switch {
	case prev.isLineComment():
		return newLineRequired
	case prev.isComment():
		return newLineAllowed
	case n.isComment():
		return newLineAllowed
	}
	return nlChk
}
