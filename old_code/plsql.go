package main

/* Tag and format Oracle PL/SQL paackages, functions, and procedures

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
func (o *plsql) isStart(q *queue, i int) bool {
	switch strings.ToUpper(q.items[i].token.Value()) {
	case "FUNCTION", "PROCEDURE", "PACKAGE":
		pnc, _ := q.prevNcWu(i)
		switch strings.ToUpper(pnc.token.Value()) {
		case "CREATE", "REPLACE", "FORCE", "":
			return true
		}
	}
	return false
}

/* isEnd returns true if the current token appears to be the
valid ending token for a PL block.

*/
func (o *plsql) isEnd(q *queue, i int) bool {
	pnc, _ := q.prevNcWu(i)

	if q.items[i].token.Value() == "/" && pnc.token.Value() == ";" {
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
	currType := Unknown

	for i := 0; i < len(q.items); i++ {

		lineNo += strings.Count(q.items[i].token.WhiteSpace(), "\n")

		if q.items[i].Type == Unknown {

			switch currType {
			case Unknown:
				if o.isStart(q, i) {
					currType = PL
					q.items[i].Type = PL
				}
			case PL:
				q.items[i].Type = PL

				lParens = q.items[i].newPDepth(lParens)
				if lParens < 0 {
					err := errors.New(fmt.Sprintf("Extra closing parens detected on line %d while tagging PL/SQL", lineNo))
					return err
				}

				switch {
				case o.isEnd(q, i):
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging PL/SQL", lineNo))
						return err
					}
				}
			}
		}
	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as PL/SQL functions/procedures.

*/
func (o *plsql) format(q *queue) (err error) {

	// Stub

	/*
		var lParens int


		for i := 0; i < len(q.items); i++ {

			if q.items[i].Type == PL {
				lParens = items[0].newPDepth(lParens)
				indents := 1

			}

			if !items[0].isComment() {
				items[1] = items[0]
			}
		}
	*/
	return err
}
