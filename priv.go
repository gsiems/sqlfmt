package main

/* Tag and format (GRANT/REVOKE) privilege statements

 */

import (
	"fmt"
	"strings"

	"github.com/gsiems/sql-parse/sqlparse"
)

type priv struct {
	state int
}

/* isStart returns true if the current token  appears to be the valid
starting token for a privilege statement.

*/
func (p *priv) isStart(items [2]wu) bool {

	switch dialect {
	case sqlparse.Oracle:
		switch strings.ToUpper(items[0].token.Value()) {
		case "GRANT":
			return strings.ToUpper(items[1].token.Value()) != "WITH"
		case "REVOKE":
			return true
		}
	default:
		switch strings.ToUpper(items[0].token.Value()) {
		case "GRANT", "REVOKE":
			return true
		}
	}

	return false
}

/* isEnd returns true if the current token appears to be the valid
ending token for a privilege statement.

*/
func (p *priv) isEnd(items [2]wu) bool {
	switch items[0].token.Value() {
	case ";":
		return true
	}
	return false
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to privilege statements.

*/
func (o *priv) tag(q *queue) (err error) {

	var items [2]wu
	currType := Unknown

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]
		if q.items[i].Type == Unknown {
			switch {
			case currType == Privilege:
				q.items[i].Type = Privilege
				if o.isEnd(items) {
					currType = Unknown
				}
			case o.isStart(items):
				currType = Privilege
				q.items[i].Type = Privilege
			}
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}

	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as privilege statements.

*/
func (o *priv) format(q *queue) (err error) {

	var items [2]wu

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		if q.items[i].Type == Privilege {
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
