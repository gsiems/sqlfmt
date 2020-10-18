package main

/* Tag and format mostly DML statements

 */

import (
	"strings"
)

type dml struct {
	state int
}

/* isStart returns true if the current token appears to be the valid
starting token for a DML statement.

*/
func (p *dml) isStart(items [2]wu) bool {

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
func (p *dml) isEnd(items [2]wu) bool {
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

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]
		if q.items[i].Type == Unknown {
			switch {
			case currType == DML:
				lParens = items[0].newPDepth(lParens)
				switch {
				case lParens < 0:
					currType = Unknown
				case strings.ToUpper(items[0].token.Value()) == "LOOP":
					currType = Unknown
				default:
					q.items[i].Type = currType
					if lParens == 0 && o.isEnd(items) {
						currType = Unknown
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
