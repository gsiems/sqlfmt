package main

/* Tag and format PostgreSQL PL/PgSGL functions and procedures

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
func (o *plpgsql) isStart(items [2]wu) bool {
	switch strings.ToUpper(items[0].token.Value()) {
	case "FUNCTION", "PROCEDURE":
		switch strings.ToUpper(items[1].token.Value()) {
		case "CREATE", "REPLACE", "":
			return true
		}
	}
	return false
}

/* isEnd returns true if the current token appears to be the
valid ending token for a PL block.

*/
func (o *plpgsql) isEnd(items [2]wu) bool {
	if o.pgt != "" {
		if items[0].token.Value() == ";" && items[1].token.Value() == o.pgt {
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
					err := errors.New(fmt.Sprintf("Extra closing parens detected on line %d while tagging PL/PgSQL", lineNo))
					return err
				}

				// look for the opening/closing PL tag
				switch {
				case o.isEnd(items):
					currType = Unknown
					if lParens > 0 {
						err := errors.New(fmt.Sprintf("Extra open parens detected on line %d while tagging PL/PgSQL", lineNo))
						return err
					}
				case lParens == 0:
					if strings.ToUpper(items[1].token.Value()) == "AS" {
						if strings.HasPrefix(items[0].token.Value(), "$") && strings.HasSuffix(items[0].token.Value(), "$") {
							o.pgt = items[0].token.Value()
						}
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
