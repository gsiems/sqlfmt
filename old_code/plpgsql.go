package main

/* Tag and format PostgreSQL PL/PgSQL functions and procedures

[CREATE [ OR REPLACE] <PL/PgSQL type> [( ... )] [...] AS $tag$ <PL/PgSQL code> $tag$


CREATE [ OR REPLACE ] FUNCTION
    name ( [ [ argmode ] [ argname ] argtype [ { DEFAULT | = } default_expr ] [, ...] ] )
    [ RETURNS rettype
      | RETURNS TABLE ( column_name column_type [, ...] ) ]
  { LANGUAGE lang_name
    | TRANSFORM { FOR TYPE type_name } [, ... ]
    | WINDOW
    | IMMUTABLE | STABLE | VOLATILE | [ NOT ] LEAKPROOF
    | CALLED ON NULL INPUT | RETURNS NULL ON NULL INPUT | STRICT
    | [ EXTERNAL ] SECURITY INVOKER | [ EXTERNAL ] SECURITY DEFINER
    | PARALLEL { UNSAFE | RESTRICTED | SAFE }
    | COST execution_cost
    | ROWS result_rows
    | SUPPORT support_function
    | SET configuration_parameter { TO value | = value | FROM CURRENT }
    | AS 'definition'
    | AS 'obj_file', 'link_symbol'
  } ...

CREATE [ OR REPLACE ] PROCEDURE
    name ( [ [ argmode ] [ argname ] argtype [ { DEFAULT | = } default_expr ] [, ...] ] )
  { LANGUAGE lang_name
    | TRANSFORM { FOR TYPE type_name } [, ... ]
    | [ EXTERNAL ] SECURITY INVOKER | [ EXTERNAL ] SECURITY DEFINER
    | SET configuration_parameter { TO value | = value | FROM CURRENT }
    | AS 'definition'
    | AS 'obj_file', 'link_symbol'
  } ...

*/

import (
	"errors"
	"fmt"
	"strings"
)

type plpgsql struct {
	state int
	pgt   string
}

/* isStart returns true if the current token appears to be the valid
starting token for a PL block.

*/
func (o *plpgsql) isStart(q *queue, i int) bool {
	switch strings.ToUpper(q.items[i].token.Value()) {
	case "FUNCTION", "PROCEDURE":
		pnc, _ := q.prevNcWu(i)
		switch strings.ToUpper(pnc.token.Value()) {
		case "CREATE", "REPLACE", "":
			return true
		}
	}
	return false
}

/* isEnd returns true if the current token appears to be the
valid ending token for a PL block.

*/
func (o *plpgsql) isEnd(q *queue, i int) bool {
	if o.pgt != "" {
		pnc, _ := q.prevNcWu(i)
		if q.items[i].token.Value() == ";" && pnc.token.Value() == o.pgt {
			o.pgt = ""
			return true
		}
	}
	return false
}

/* tag iterates through the queue and tags the tokens that are believed
to belong to PL/PgSQL functions/procedures.

ASSERTION: The DML tokens have already been tagged

*/
func (o *plpgsql) tag(q *queue) (err error) {

	var lineNo int
	var lParens int
	var pnc wu
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
					err := errors.New(fmt.Sprintf("Extra closing parens detected on line %d while tagging PL/PgSQL", lineNo))
					return err
				}

				// look for the opening/closing PL tag
				switch {
				case o.isEnd(q, i):
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging PL/PgSQL", lineNo))
						return err
					}
				case lParens == 0:
					if strings.ToUpper(pnc.token.Value()) == "AS" {
						if strings.HasPrefix(q.items[i].token.Value(), "$") && strings.HasSuffix(q.items[i].token.Value(), "$") {
							o.pgt = q.items[i].token.Value()
						}
					}
				}
			}
		}
		if !q.items[i].isComment() {
			pnc = q.items[i]
		}
	}
	return err
}

/* format iterates through the queue and determines the formatting for the
work units that are tagged as PL/PgSQL functions/procedures.

*/
func (o *plpgsql) format(q *queue) (err error) {

	// Stub

	/*
		var lParens int

		for i := 0; i < len(q.items); i++ {
			ci = q.items[i]

			if q.items[i].Type == PL {
				lParens = ci.newPDepth(lParens)
				indents := 1

			}

			if !ci.isComment() {
				items[1] = ci
			}
		}
	*/
	return err
}
