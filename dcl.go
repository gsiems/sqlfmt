package main

/* Tag and format DCL (GRANT/REVOKE) statements

GRANT { { object_privilege[, ...] | [role[, ...] } }
    [ON database_object_name]
    [TO grantee[, ...]]
    [WITH HIERARCHY OPTION] [WITH GRANT OPTION] [WITH ADMIN OPTION]
    [FROM {CURRENT_USER | CURRENT_ROLE}]

*/

import (
	"strings"

	"github.com/gsiems/sql-parse/sqlparse"
)

type dcl struct {
	state int
}

/* isStart returns true if the current token  appears to be the valid
starting token for a DCL statement.

*/
func (p *dcl) isStart(q *queue, i int) bool {

	switch strings.ToUpper(q.items[i].token.Value()) {
	case "GRANT":
		// Ensure this isn't part of "WITH GRANT OPTION"
		pnc, _ := q.prevNcWu(i)
		return strings.ToUpper(pnc.token.Value()) != "WITH"
	case "REVOKE":
		return true
	case "REASSIGN":
		return dialect == sqlparse.PostgreSQL
	}

	return false
}

/* isEnd returns true if the current token appears to be the valid
ending token for a DCL statement.

*/
func (p *dcl) isEnd(q *queue, i int) bool {
	return q.items[i].token.Value() == ";"
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to DCL statements.

*/
func (o *dcl) tag(q *queue) (err error) {

	currType := Unknown

	for i := 0; i < len(q.items); i++ {

		if q.items[i].Type == Unknown {
			switch {
			case currType == DCL:
				q.items[i].Type = DCL
				if o.isEnd(q, i) {
					currType = Unknown
				}
			case o.isStart(q, i):
				currType = DCL
				q.items[i].Type = DCL
			}
		}
	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as DCL statements.

*/
func (o *dcl) format(q *queue) (err error) {

	var pnc wu

	for i := 0; i < len(q.items); i++ {

		if q.items[i].Type == DCL {
			indents := 1

			// check for new line requirements
			nlChk := NoNewLine
			switch {
			case i == 0:
				// nada
			case o.isStart(q, i):
				nlChk = NewLineRequired
				indents = 0
			default:
				// check for comment
				nlChk = chkCommentNL(q, i, nlChk)
			}

			// vertical spaces
			vertSp := q.items[i].verticalSpace(2)
			switch nlChk {
			case NewLineRequired:
				vertSp = maxInt(vertSp, 1)
			case NewLineAllowed:
				vertSp = vertSp
			default:
				vertSp = 0
			}

			if vertSp == 0 {
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
				q.items[i].vertSp = vertSp
				q.items[i].indents = indents
			}

			q.items[i].value = q.items[i].formatValue()
		}

		if !q.items[i].isComment() {
			pnc = q.items[i]
		}
	}
	return err
}
