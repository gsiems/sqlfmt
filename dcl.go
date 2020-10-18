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
func (p *dcl) isStart(items [2]wu) bool {

	switch strings.ToUpper(items[0].token.Value()) {
	case "GRANT":
		// Ensure this isn't part of "WITH GRANT OPTION"
		return strings.ToUpper(items[1].token.Value()) != "WITH"
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
func (p *dcl) isEnd(items [2]wu) bool {
	return items[0].token.Value() == ";"
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to DCL statements.

*/
func (o *dcl) tag(q *queue) (err error) {

	var items [2]wu
	currType := Unknown

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]
		if q.items[i].Type == Unknown {
			switch {
			case currType == DCL:
				q.items[i].Type = DCL
				if o.isEnd(items) {
					currType = Unknown
				}
			case o.isStart(items):
				currType = DCL
				q.items[i].Type = DCL
			}
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}

	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as DCL statements.

*/
func (o *dcl) format(q *queue) (err error) {

	var items [2]wu

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		if q.items[i].Type == DCL {
			indents := 1

			// check for new line requirements
			nlChk := noNewLine
			switch {
			case i == 0:
				// nada
			case o.isStart(items):
				nlChk = newLineRequired
				indents = 0
			default:
				// check for comment
				nlChk = chkCommentNL(q.items[i], q.items[i-1], nlChk)
			}

			// vertical spaces
			vertSp := items[0].verticalSpace(2)
			switch nlChk {
			case newLineRequired:
				vertSp = maxInt(vertSp, 1)
			case newLineAllowed:
				vertSp = vertSp
			default:
				vertSp = 0
			}

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
				q.items[i].vertSp = vertSp
				q.items[i].indents = indents
			}

			q.items[i].value = items[0].token.Value()
			//q.items[i].value = formatValue()
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}
	}
	return err
}
