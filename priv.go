package main

/* Tag and format (GRANT/REVOKE) privilege statements

 */

import (
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
