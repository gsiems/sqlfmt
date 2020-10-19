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
func (o *dml) isStart(items [2]wu) bool {

	switch strings.ToUpper(items[0].token.Value()) {
	case "WITH", "MERGE", "SELECT", "TRUNCATE":
		return true
	case "UPDATE", "INSERT", "UPSERT", "DELETE":
		switch strings.ToUpper(items[1].token.Value()) {
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
func (o *dml) isEnd(items [2]wu) bool {
	switch items[0].token.Value() {
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
	var items [2]wu
	currType := Unknown
	var lineNo int

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		lineNo += strings.Count(items[0].token.WhiteSpace(), "\n")

		if q.items[i].Type == Unknown {
			switch {
			case currType == DML:
				lParens = items[0].newPDepth(lParens)
				switch {
				case lParens < 0:
					currType = Unknown
				case strings.ToUpper(items[0].token.Value()) == "LOOP":
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging DML", lineNo))
						return err
					}
				default:
					q.items[i].Type = currType
					if o.isEnd(items) {
						currType = Unknown
						if lParens > 0 {
							err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging DML", lineNo))
							return err
						}
					}
				}
			case o.isStart(items):
				currType = DML
				q.items[i].Type = DML
				lParens = 0
			}
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}

	}
	return err
}

func (o *dml) preStack(items [2]wu, pDepth int) (blk dmlBlock) {

	// open parens -> push stack
	// close parens -> pop stack
	// primary clause -> set stack type (select, insert, update, delete, merge, upsert)
	// secondary clause -> set stack progress/state (with, from, where, ...) ???
	// case structures -> ???
	// does any database support `if ... then` stuctures in DML?

	if len(o.stack) == 0 {
		o.push(dmlBlock{Type: Unknown, pDepth: pDepth})
	}

	curr := strings.ToUpper(items[0].token.Value())

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

func (o *dml) postStack(items [2]wu, pDepth int) {

	blk := o.top()
	curr := strings.ToUpper(items[0].token.Value())

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

func (o *dml) nlCheck(items [2]wu, prev wu, pDepth int) (nlChk int) {

	blk := o.top()
	nlChk = NoNewLine
	switch {
	case o.isStart(items):
		return NewLineRequired
	case items[1].token.Value() == ",":
		if pDepth == blk.pDepth && blk.Type != Unknown {
			return NewLineRequired
		}
	}

	curr := strings.ToUpper(items[0].token.Value())

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
		switch strings.ToUpper(items[1].token.Value()) {
		case "INSERT":
			// nada
		default:
			if pDepth == blk.pDepth {
				return NewLineRequired
			}
		}
	case "INNER", "LEFT", "RIGHT", "FULL", "CROSS", "NATURAL", "ON", "USING":
		if pDepth == blk.pDepth {
			return NewLineRequired
		}
	case "JOIN":
		switch strings.ToUpper(prev.token.Value()) {
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

	return chkCommentNL(items[0], prev, nlChk)
}

func (o *dml) vertSp(items [2]wu, prev wu, nlChk int) int {

	vertSp := items[0].verticalSpace(2)
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

func (o *dml) indents(items [2]wu, prev wu, nlChk int) (i int) {

	// if is comment then want to use previous indent?

	for _, v := range o.stack {
		i++
		if v.Type != Unknown {
			i++
		}
	}

	curr := strings.ToUpper(items[0].token.Value())
	switch curr {
	case "WITH", "SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "UPSERT", "TRUNCATE":
		i -= 2
	case "UNION", "INTERSECT", "EXCEPT", "MINUS":
		i -= 2
	case "CASE":
		i -= 2
	case "WHEN", "ELSE", "END":
		i--
	case "JOIN", "INNER", "LEFT", "RIGHT", "FULL", "CROSS", "NATURAL":
		i--
	case "INTO", "FROM", "WHERE", "GROUP", "ORDER", "HAVING", "LIMIT", "OFFSET":
		i--
	}

	return maxInt(i, 0)
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
	var items [2]wu
	var prev wu

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		if q.items[i].Type != DML {
			baseIndents = q.items[i].indents
		}

		if q.items[i].Type == DML {
			lParens = items[0].newPDepth(lParens)
			//indents := 1

			_ = o.preStack(items, lParens)

			nlChk := NoNewLine
			if i > 0 {
				nlChk = o.nlCheck(items, prev, lParens)
			}

			vertSp := o.vertSp(items, prev, nlChk)

			if vertSp == 0 {
				switch {
				case items[0].token.Value() == ",":
					// nada
				case strings.HasPrefix(items[0].token.Value(), "."):
					// nada
				case strings.HasSuffix(items[1].token.Value(), "."):
					// nada
				case i == 0:
					// nada
				default:
					q.items[i].leadSp = 1
				}

			} else {
				// calc indents based on stack
				indents := o.indents(items, prev, nlChk)
				q.items[i].vertSp = vertSp
				q.items[i].indents = indents + baseIndents
			}

			q.items[i].value = items[0].formatValue()

			o.postStack(items, lParens)

		}

		prev = items[0]
		if !items[0].isComment() {
			items[1] = items[0]
		}
	}

	return err
}
