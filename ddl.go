package main

import (
	"strings"
)

type ddl struct {
	state int
}

/* isStart returns true if the current token appears to be the valid
starting token for a DDL statement.

*/
func (p *ddl) isStart(items [2]wu) bool {

	switch strings.ToUpper(items[0].token.Value()) {
	case "CREATE", "ALTER", "DROP", "COMMENT":
		return true
	case "SET", "SHOW":
		return true
	}
	return false
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to DDL statements.

ASSERTION: tagging DDL is the final tagging operation, therefore
anything not otherwise tagged is considered to be DDL

*/
func (o *ddl) tag(q *queue) (err error) {
	for i := 0; i < len(q.items); i++ {
		if q.items[i].Type == Unknown {
			q.items[i].Type = DDL
		}
	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as DDL statements.

*/
func (o *ddl) format(q *queue) (err error) {

	var lParens int
	inDDL := false
	var items [2]wu

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		if q.items[i].Type == DDL {
			lParens = items[0].newPDepth(lParens)
			indents := 1

			// check for new line requirements
			nlChk := noNewLine
			switch {
			case i == 0:
				// nada
			case o.isStart(items):
				nlChk = newLineRequired
				inDDL = true
				indents = 0
			case items[1].token.Value() == ",":
				if lParens == 0 {
					nlChk = newLineRequired
				}
			default:
				// check for comment

				/* Need to determine the difference between comments
				   that are stand-alone and comments that are embedded
				   within a DDL statement as the first should have no
				   indent and the latter should have an indent */
				if items[0].isComment() && !inDDL {
					indents = 0
				}

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

		/* If the code is creating a PostgreSQL PL/PgSQL, or
		   Oracle PL/SQL, object then there will be no [available]
		   trailing semi-colon to use for identifying the end of the DDL.
		*/
		switch {
		case items[0].token.Value() == ";":
			inDDL = false
		case items[1].Type == PL:
			switch items[1].token.Value() {
			case "/", ";":
				inDDL = false
			}
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}
	}
	return err
}
