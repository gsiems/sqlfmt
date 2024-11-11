
```psql -t -c "with base as ( select '2024-11-03'::date as reboot_date ) select 'Day ' || (current_date - reboot_date)::text from base ;"```

## Day 0

```
mkdir old_code
git mv *.go test* old_code/

go version

cat <<EOT> go.mod
module github.com/gsiems/sqlfmt

go 1.20
EOT

echo "" > go.sum

git add go.mod go.sum dev_notes/

git commit -m "Reboot project"
```

```
git add dialect dev_notes/JOURNAL.md

git commit -m "Lift and shift the dialect module code from github.com/gsiems/sql-parse.
Fix a few bugs and refactor to implement using an interface."
```

```
git add env dev_notes/JOURNAL.md

git commit -m "Add environment module for tracking values of configuration parameters"
```

```
git add parser dev_notes/

git commit -m "Lift and shift the parser module code from github.com/gsiems/sql-parse, refactoring as needed."
```

```
git add env/env.go

git commit -m "Fix 'duplicate case \"identcase\"' error"

git add formatter/formatter.go formatter/formatter_test.go formatter/run_tests.sh dev_notes/

git commit -m "Setup initial testing of the parser."
```

The next step is to group the tokens that go together as part of a larger
object, unit, or command into "bags" such that each bag can be formatted
separately. Sub-units such as sub-selects, CTEs, etc. should also be bagged
separately as that should make the formatting code less complex.

```
git add formatter/*.go dev_notes/

git commit -m "Added tagging for DCL commands and comments."
```

## Day 1

### Thoughts on formatting.

* Needs to be multi-pass with the first pass updating the high-level
vertical-space and indentations.

  * For DML this would result in the main clause keywords like "SELECT",
  "FROM", "WHERE" having the v-space and indentations set as well as the
  beginning of each column expression having the v-space and indentations set.

  * Would probably need to differentiate the high-level v-space from the
  comment related v-space (either the v-space prior to a comment or the v-space
  of the token following the comment (if not a main clause keyword)).

* The second pass would update the indentation of those tokens that are bag
pointers. Actually, this pass would need to update the indentations of the
tokens contained in the the bag being pointed to. Probably needs to be
iterative until the appropriate level of indentations have been propagated to
all sub-bags.

* The third pass should take care of wrapping lines (due to comments, overly
long lines, case structures, etc.)

### Regarding testing.

* Since the current directory structure for test files includes the database
dialect, the dialect does not need to be specified in the directive comment at
the top of the file.

* There should be a set of tests for parsing the directive comment to ensure
that the resulting environment matches the directives.

* For the sql test files-- multiple tests can be run from the same input file
for spaces vs. tabs, upper-case vs. lower case keywords, etc.

Back to tagging...

```
git add formatter/*.go dev_notes/

git commit -m "Added tagging for DML statements."
```

## Day 2

Woohoo!

```
git add formatter/*.go dev_notes/

git commit -m "Added tagging for PostgreSQL functions and procedures."
```

## Day 6

Having separated the PostgreSQL functions and procedures into body vs. non-body
as part of the tagging, the two can be formatted separately.

```
git add formatter/*.go dev_notes/

git commit -m "Added formatting of the non-body portion of PostgreSQL functions and procedures."
```

## Day 7

Not formatting things.

Given that formatting can be disabled for an entire input, what about inputs
that only have a portion that should not be formatted? For example the
plain-text data portion of a COPY command, or the body of a PL/perl, PL/python,
PL/tcl, etc. function/procedure?

Options:

1. Cowardly refuse to format the entire input (same as setting a do-not-format
directive). Probably safest approach. Would need to feedback warnings/errors
such as "non formattable code found". Having a feedback mechanism could also be
used to flag instances of unbalanced parens or, in the case of PL code,
unmatched begin/end blocks (or if/then, case, loop blocks).

2. Selectively format the input by separating the non-formattable portion of
the input from the formattable and only formatting the formattable portion.
This would be the nicest, provided it worked reliably and didn't mess up the
non-formattable portion of the input.
