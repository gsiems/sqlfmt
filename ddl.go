package main

/* Tag and format mostly DDL statements. Since DDL is the last thing
tagged it also includes some things that aren't really DDL but that
don't belong in the previous categories either.

*/

import (
	"strings"
)

type ddl struct {
	state int
}

/* isStart returns true if the current token appears to be the valid
starting token for a DDL statement.

*/
func (p *ddl) isStart(q *queue, i int) bool {

	switch strings.ToUpper(q.items[i].token.Value()) {
	case "CREATE", "ALTER", "DROP", "COMMENT":
		return true
	case "SET", "SHOW": // not really DDL
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
	var pnc wu
	inDDL := false

	for i := 0; i < len(q.items); i++ {

		if q.items[i].Type == DDL {
			lParens = q.items[i].newPDepth(lParens)
			indents := 1

			// check for new line requirements
			nlChk := NoNewLine
			switch {
			case i == 0:
				// nada
			case o.isStart(q, i):
				nlChk = NewLineRequired
				inDDL = true
				indents = 0
			case pnc.token.Value() == ",":
				if lParens == 0 {
					nlChk = NewLineRequired
				}
			default:
				// check for comment

				/* Need to determine the difference between comments
				   that are stand-alone and comments that are embedded
				   within a DDL statement as the first should have no
				   indent and the latter should have an indent */
				if q.items[i].isComment() && !inDDL {
					indents = 0
				}

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

		/* If the code is creating a PostgreSQL PL/PgSQL, or
		Oracle PL/SQL, object then there will be no [available]
		trailing semi-colon to use for identifying the end of the DDL.

		Also, with Oracle SQL-Plus scripts the SET command may not
		have a trailing semi-colon either. It appears that the new
		line is the terminator. SHOW behaves similarly?

		*/
		switch {
		case q.items[i].token.Value() == ";":
			inDDL = false
		case pnc.Type == PL:
			switch pnc.token.Value() {
			case "/", ";":
				inDDL = false
			}
		}

		if !q.items[i].isComment() {
			pnc = q.items[i]
		}
	}
	return err
}
