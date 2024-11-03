package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gsiems/sql-parse/sqlparse"
)

/* A Work Unit [wu] is a container unit that contains either an
un-formatted sqlparse token or the results of formatting actions

*/
type wu struct {
	Type    int
	vertSp  int    // number of newlines (vertical space) before the work unit
	indents int    // the number of indentation units before the work unit
	leadSp  int    // the number of leading spaces before the value
	value   string // the formatted contents of the work unit
	token   sqlparse.Token
}

/* queue is an ordered list of contiguous work units that have all been
tagged as being of the same type (DML, DDL, etc.)

*/
type queue struct {
	Type  int
	items []wu
}

/* initialzeQueue takes the list of sqlparse tokens and populates the
initial work unit queue

*/
func initialzeQueue(tokens sqlparse.Tokens) (q queue, err error) {

	var pDepth int
	var lineNo int

	q.Type = Unknown

	tokens.Rewind()
	for {
		t := tokens.Next()

		s := t.Value()
		if s == "" {
			// there is nothing left to parse
			break
		}

		lineNo += strings.Count(t.WhiteSpace(), "\n")

		switch s {
		case "(":
			pDepth++
		case ")":
			pDepth--
		}

		if pDepth < 0 {
			err = errors.New(fmt.Sprintf("Extra closing parens detected on line %d", lineNo))
			return q, err
		}

		lineNo += strings.Count(s, "\n")

		q.items = append(q.items, wu{Type: Unknown, token: t})
	}

	if pDepth > 0 {
		err = errors.New("Insufficient closing parens detected at end of file")
		return q, err
	}

	q.items = append(q.items, wu{Type: Final})

	return q, err
}

/* prevWu returns the work unit prior to the specified one

*/
func (q *queue) prevWu(i int) (n wu) {
	if i > 0 {
		n = q.items[i-1]
	}
	return n
}

/* prevNcWu returns the first non-comment work unit prior to the
specified one (and the index where it was found)

 */
func (q *queue) prevNcWu(i int) (n wu, ni int) {

	for ni = i - 1; ni >= 0; ni-- {
		if !q.items[ni].isComment() {
			return q.items[ni], ni
		}
	}
	return n, ni
}

func (n *wu) isComment() bool {
	switch n.token.Type() {
	case sqlparse.LineCommentToken, sqlparse.PoundLineCommentToken, sqlparse.BlockCommentToken:
		return true
	}
	return false
}

func (n *wu) isLineComment() bool {
	switch n.token.Type() {
	case sqlparse.LineCommentToken, sqlparse.PoundLineCommentToken:
		return true
	}
	return false
}

func (n *wu) newPDepth(i int) int {

	switch n.token.Value() {
	case "(":
		i++
	case ")":
		i--
	}
	return i
}

func (n *wu) verticalSpace(maxVSp int) (vSp int) {
	vSp = strings.Count(n.token.WhiteSpace(), "\n")
	if maxVSp > 0 {
		if vSp > maxVSp {
			return maxVSp
		}
	}
	return vSp
}

func (n *wu) formatValue() (s string) {
	// placeholder for now
	return n.token.Value()
}
