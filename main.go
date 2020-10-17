package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/gsiems/sql-parse/sqlparse"
)

//
var (
	showVersion    = flag.Bool("version", false, "")
	version        = ""
	indentSz       = flag.Int("indent", 4, "")
	dialectName    = flag.String("dialect", "standard", "")
	inputFile      = flag.String("i", "", "")
	outputFile     = flag.String("o", "", "")
	keyCase        = flag.String("k", "upper", "")
	preserveCase   = flag.Bool("p", false, "")
	preserveQuotes = flag.Bool("q", false, "")
	debug          = flag.Bool("debug", false, "")
	ident          = "    "
	dialect        = sqlparse.StandardSQL

// flags for ???
//  - translate (some) things to standard SQL (Oracle decode, nvl, etc.)
//  - max line length
//  - leading commas
)

func main() {
	rc, err := runapp()
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("%s", err))
	}
	os.Exit(rc)
}

func runapp() (rc int, err error) {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: sqlfmt [flags]

  -dialect  the SQL dialect of the input (default is standard) (standard,postgres,sqlite,mariadb,mssql,mysql,oracle)
  -i        the file to read (defaults to stdin)
  -o        the file to write to (defaults to stdout)
  -indent   number of spaces to indent (default is 4), use 0 for tabs
  -k        keywords case (default is upper) (upper,lower,nochange)
  -p        preserve case of non-keywords (default is to lower case non-keywords)
  -q        preserve quoted identifiers (default is to unquote identifiers when possible)
`)
	}
	flag.Parse()

	dialect = resolveDialect(*dialectName)

	var input string
	input, err = readInput(*inputFile)
	if err != nil {
		return 1, err
	}

	var formatted string
	formatted, err = runFormatter(input, dialect)
	if err != nil {
		return 1, err
	}

	err = writeOutput(*outputFile, formatted)
	if err != nil {
		return 1, err
	}

	return 0, err
}

func resolveDialect(s string) int {
	var dialects = map[string]int{
		"standard": sqlparse.StandardSQL,
		"postgres": sqlparse.PostgreSQL,
		"sqlite":   sqlparse.SQLite,
		"mysql":    sqlparse.MySQL,
		"oracle":   sqlparse.Oracle,
		"mssql":    sqlparse.MSSQL,
		"mariadb":  sqlparse.MariaDB,
	}

	d, ok := dialects[s]
	if !ok {
		return sqlparse.StandardSQL
	}
	return d
}

func readInput(f string) (input string, err error) {

	var inBytes []byte

	switch f {
	case "", "-":
		reader := bufio.NewReader(os.Stdin)
		inBytes, err = ioutil.ReadAll(reader)
	default:
		inBytes, err = ioutil.ReadFile(f)
	}

	return string(inBytes), err
}

func writeOutput(f, output string) (err error) {

	switch f {
	case "", "-":
		fmt.Print(output)
	default:
		err = ioutil.WriteFile(f, []byte(output), 0644)
	}

	return err
}

func runFormatter(input string, dialect int) (formatted string, err error) {

	var q queue
	tokens := sqlparse.ParseStatements(input, dialect)
	q, err = initialzeQueue(tokens)
	if err != nil {
		return formatted, err
	}

	//
	var Priv priv
	var DML dml
	var PLPgSQL plpgsql
	err = Priv.tag(&q)
	if err != nil {
		return formatted, err
	}
	err = DML.tag(&q)
	if err != nil {
		return formatted, err
	}

	switch dialect {
	case sqlparse.PostgreSQL:
		err = PLPgSQL.tag(&q)
		if err != nil {
			return formatted, err
		}
	}

	// temp for validating tagging to this point
	var s []string
	for _, v := range q.items {
		s = append(s, fmt.Sprintf("%v: %q", v.Type, v.token.Value()))
	}

	formatted = strings.Join(s, "\n")

	return formatted, err
}
