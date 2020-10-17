package main

/* Tag and format Oracle PL/SQL functions and procedures

[CREATE [ OR REPLACE] <PL/SQL type> [( ... )] [...] [IS|AS] <PL/SQL code> /

*/

import (
	"errors"
	"fmt"
	"strings"
)

type plsql struct {
	state int
}

/* isStart returns true if the current token appears to be the valid
starting token for a PL block.

*/
func (o *plsql) isStart(items [2]wu) bool {
	switch strings.ToUpper(items[0].token.Value()) {
	case "FUNCTION", "PROCEDURE", "PACKAGE":
		switch strings.ToUpper(items[1].token.Value()) {
		case "CREATE", "REPLACE", "FORCE", "":
			return true
		}
	}
	return false
}

/* isEnd returns true if the current token appears to be the
valid ending token for a PL block.

*/
func (o *plsql) isEnd(items [2]wu) bool {

	if items[0].token.Value() == "/" && items[1].token.Value() == ";" {
		return true
	}
	return false
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to PL/SQL functions/procedures.

ASSERTION: The DML tokens have already been tagged

*/
func (o *plsql) tag(q *queue) (err error) {

	var lineNo int
	var lParens int
	var items [2]wu
	currType := Unknown

	for i := 0; i < len(q.items); i++ {
		items[0] = q.items[i]

		lineNo += strings.Count(items[0].token.WhiteSpace(), "\n")

		if q.items[i].Type == Unknown {

			switch currType {
			case Unknown:
				if o.isStart(items) {
					currType = PL
					q.items[i].Type = PL
				}
			case PL:
				q.items[i].Type = PL

				lParens = items[0].newPDepth(lParens)
				if lParens < 0 {
					err := errors.New(fmt.Sprintf("Extra closing parens detected on line %d while tagging PL/SQL", lineNo))
					return err
				}

				switch {
				case o.isEnd(items):
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging PL/SQL", lineNo))
						return err
					}
				}
			}
		}

		if !items[0].isComment() {
			items[1] = items[0]
		}

	}
	return err
}
