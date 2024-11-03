
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
