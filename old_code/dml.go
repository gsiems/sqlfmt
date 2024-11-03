package main

/* Tag and format mostly DML statements

 */

import (
	"errors"
	"fmt"
	"strings"
)

var DmlTypes = map[string]int{
	"Unknown": Unknown,
	"SELECT":  Select,
	"INSERT":  Insert,
	"UPDATE":  Update,
	"DELETE":  Delete,
	"MERGE":   Merge,
	"UPSERT":  Upsert,
	"CASE":    CaseBlock,
}

type dmlBlock struct {
	Type      int
	pDepth    int
	minorType string // int ???
}

type dml struct {
	stack []dmlBlock
}

func (o *dml) top() (blk dmlBlock) {
	if len(o.stack) > 0 {
		blk = o.stack[len(o.stack)-1]
	}

	return blk
}

func (o *dml) pop() (blk dmlBlock) {
	if len(o.stack) > 0 {
		o.stack = o.stack[:len(o.stack)-1]
	}

	return o.top()
}

func (o *dml) push(blk dmlBlock) {
	o.stack = append(o.stack, blk)
}

func (o *dml) setType(s string) {
	switch o.stack[len(o.stack)-1].Type {
	case Unknown:
		r, ok := DmlTypes[s]
		if ok {
			o.stack[len(o.stack)-1].Type = r
		}
	}
}

/* isStart returns true if the current token appears to be the valid
starting token for a DML statement.

*/
func (o *dml) isStart(q *queue, i int) bool {

	switch strings.ToUpper(q.items[i].token.Value()) {
	case "WITH", "MERGE", "SELECT", "TRUNCATE":
		return true
	case "UPDATE", "INSERT", "UPSERT", "DELETE":
		pnc, _ := q.prevNcWu(i)
		switch strings.ToUpper(pnc.token.Value()) {
		// trigger: (BEFORE|AFTER|INSTEAD OF) (UPDATE|INSERT|DELETE|UPSERT?) ON
		case "BEFORE", "AFTER", "OF":
			return false
		}
		return true
	}
	return false
}

/* isEnd returns true if the current token appears to be the
valid ending token for a DML statement.

*/
func (o *dml) isEnd(q *queue, i int) bool {
	switch q.items[i].token.Value() {
	case ";":
		return true
	}
	return false
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to DML statements.

NB-1: there is a need to track the open parens count (pDepth) when
in a DML statement in the event that the DML is part of some PL. For
example, in:

for r in (
    select ... ) loop
    ...
end loop ;

there is no trailing semi-colon at the end of the embedded DML statement

*/
func (o *dml) tag(q *queue) (err error) {

	var lParens int
	currType := Unknown
	var lineNo int

	for i := 0; i < len(q.items); i++ {

		lineNo += strings.Count(q.items[i].token.WhiteSpace(), "\n")

		if q.items[i].Type == Unknown {
			switch {
			case currType == DML:
				lParens = q.items[i].newPDepth(lParens)
				switch {
				case lParens < 0:
					currType = Unknown
				case strings.ToUpper(q.items[i].token.Value()) == "LOOP":
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging DML", lineNo))
						return err
					}
				default:
					q.items[i].Type = currType
					if o.isEnd(q, i) {
						currType = Unknown
						if lParens > 0 {
							err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging DML", lineNo))
							return err
						}
					}
				}
			case o.isStart(q, i):
				currType = DML
				q.items[i].Type = DML
				lParens = 0
			}
		}
	}
	return err
}

func (o *dml) preStack(q *queue, i, pDepth int) (blk dmlBlock) {

	// open parens -> push stack
	// close parens -> pop stack
	// primary clause -> set stack type (select, insert, update, delete, merge, upsert)
	// secondary clause -> set stack progress/state (with, from, where, ...) ???
	// case structures -> ???
	// does any database support `if ... then` stuctures in DML?

	if len(o.stack) == 0 {
		o.push(dmlBlock{Type: Unknown, pDepth: pDepth})
	}

	curr := strings.ToUpper(q.items[i].token.Value())

	switch curr {
	case "(":
		o.push(dmlBlock{Type: Unknown, pDepth: pDepth})

	case ";":
		o.stack = nil

	case "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "UPSERT":
		o.setType(curr)
	case "UNION", "INTERSECT", "EXCEPT", "MINUS":
		o.stack[len(o.stack)-1].Type = Unknown
	case "CASE":
		o.push(dmlBlock{Type: CaseBlock, pDepth: pDepth})
	}

	return o.top()
}

func (o *dml) postStack(q *queue, i, pDepth int) {

	blk := o.top()
	curr := strings.ToUpper(q.items[i].token.Value())

	switch curr {
	case "END":
		// if type is CASE then pop
		blk = o.top()
		switch blk.Type {
		case CaseBlock:
			o.pop()
		}
	case ")":
		o.pop()
	}

}

func (o *dml) nlCheck(q *queue, i, pDepth int) (nlChk int) {

	blk := o.top()
	nlChk = NoNewLine
	pnc, _ := q.prevNcWu(i)

	switch {
	case o.isStart(q, i):
		return NewLineRequired
	case pnc.token.Value() == ",":
		if pDepth == blk.pDepth && blk.Type != Unknown {
			return NewLineRequired
		}
	}

	curr := strings.ToUpper(q.items[i].token.Value())

	switch blk.Type {
	case Unknown:
		switch curr {
		case "UNION", "INTERSECT", "EXCEPT", "MINUS":
			if pDepth == blk.pDepth {
				return NewLineRequired
			}
		}
	case CaseBlock:
		switch curr {
		case "CASE", "WHEN", "ELSE", "END":
			return NewLineRequired
		}
	}
	switch curr {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "UPSERT":
		return NewLineRequired
	case "FROM", "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "OFFSET":
		if pDepth == blk.pDepth {
			return NewLineRequired
		}
	case "INTO":
		switch strings.ToUpper(pnc.token.Value()) {
		case "INSERT":
			// nada
		default:
			if pDepth == blk.pDepth {
				return NewLineRequired
			}
		}
	case "INNER", "LEFT", "RIGHT", "FULL", "CROSS", "NATURAL", "LATERAL", "ON", "USING":
		if pDepth == blk.pDepth {
			return NewLineRequired
		}
	case "JOIN":
		switch strings.ToUpper(pnc.token.Value()) {
		case "OUTER":
			// nada
		default:
			if pDepth == blk.pDepth {
				return NewLineRequired
			}
		}
	case "AND", "OR":
		return NewLineRequired
	}

	return chkCommentNL(q, i, nlChk)
}

func (o *dml) vertSp(q *queue, i, nlChk int) int {

	vertSp := q.items[i].verticalSpace(2)
	switch nlChk {
	case NewLineRequired:
		return maxInt(vertSp, 1)
	case NewLineAllowed:
		if vertSp > 0 {
			return maxInt(vertSp, 1)
		}
	}

	return 0
}

func (o *dml) indents(q *queue, i, nlChk int) (indents int) {

	// if is comment then want to use previous indent?

	for _, v := range o.stack {
		indents++
		if v.Type != Unknown {
			indents++
		}
	}

	curr := strings.ToUpper(q.items[i].token.Value())
	switch curr {
	case "WITH", "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "UPSERT", "TRUNCATE":
		indents -= 2
	case "UNION", "INTERSECT", "EXCEPT", "MINUS":
		indents -= 2
	case "CASE":
		indents -= 2
	case "WHEN", "ELSE", "END":
		indents--
	case "JOIN", "INNER", "LEFT", "RIGHT", "FULL", "CROSS", "NATURAL", "LATERAL":
		indents--
	case "INTO", "FROM", "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "OFFSET":
		indents--
	}

	return maxInt(indents, 0)
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as DML statements

NB: DML is the one type of statement that may have an initial indent as
a result of being embedded in PL. This also means that DML needs to be
formatted after any PL blocks.

*/
func (o *dml) format(q *queue) (err error) {

	var lParens int
	var baseIndents int

	for i := 0; i < len(q.items); i++ {

		if q.items[i].Type != DML {
			baseIndents = q.items[i].indents
		}

		if q.items[i].Type == DML {
			lParens = q.items[i].newPDepth(lParens)
			//indents := 1

			_ = o.preStack(q, i, lParens)

			nlChk := NoNewLine
			if i > 0 {
				nlChk = o.nlCheck(q, i, lParens)
			}

			vertSp := o.vertSp(q, i, nlChk)

			if vertSp == 0 {
				pnc, _ := q.prevNcWu(i)
				switch {
				case q.items[i].token.Value() == ",":
					// nada
				case strings.HasPrefix(q.items[i].token.Value(), "."):
					// nada
				case strings.HasSuffix(pnc.token.Value(), "."):
					// nada
				case i == 0:
					// nada
				default:
					q.items[i].leadSp = 1
				}

			} else {
				// calc indents based on stack
				indents := o.indents(q, i, nlChk)
				q.items[i].vertSp = vertSp
				q.items[i].indents = indents + baseIndents
			}

			q.items[i].value = q.items[i].formatValue()

			o.postStack(q, i, lParens)
		}
	}

	return err
}
