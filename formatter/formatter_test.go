package formatter

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

func TestSQLFiles(t *testing.T) {

	verbose := false

	dialects := []string{"mariadb", "mssql", "mysql", "oracle", "postgresql", "sqlite", "standard"}

	for _, d := range dialects {

		inputDir := path.Join("..", "testdata", "input", d)
		parsedDir := path.Join("..", "testdata", "parsed")
		taggedDir := path.Join("..", "testdata", "tagged")

		files, err := ioutil.ReadDir(inputDir)
		if err != nil {
			t.Errorf("%s", err)
		}

		for _, file := range files {
			// Ensure that it is a *.sql file
			if !strings.HasSuffix(file.Name(), ".sql") {
				continue
			}

			inputFile := path.Join(inputDir, file.Name())

			inBytes, err := ioutil.ReadFile(inputFile)
			if err != nil {
				t.Errorf("%s (%s)", file.Name(), err)
			}
			input := string(inBytes)

			e := env.NewEnv()

			// Extract the parsing args from the first line of the input
			// and determine which dialect to use, etc.
			l1 := strings.SplitN(input, "\n", 2)[0]
			l1 = strings.TrimLeft(l1, "-#/* \t")
			if strings.HasPrefix(l1, "sqlfmt") {
				e.SetDirectives(l1)
			}

			e.SetInputFile(inputFile)

			p := parser.NewParser(d)

			////////////////////////////////////////////////////////////////////////
			// Parse the input and compare to expected
			parsed := p.ParseStatements(input)

			err = writeParsed(parsedDir, d, file.Name(), parsed, e)
			if err != nil {
				t.Errorf("Error writing parsed for %s: %s", file.Name(), err)
				continue
			}
			if verbose {
				err = compareFiles(parsedDir, d, file.Name())
				if err != nil {
					t.Errorf("Error comparing parsed for %s: %s", file.Name(), err)
				}
			}

			////////////////////////////////////////////////////////////////////////
			// Tag the tokens and compare to expected
			cleaned := cleanupParsed(e, parsed)
			bagMap, mainTokens := tagBags(e, cleaned)

			err = writeTagged(taggedDir, d, file.Name(), mainTokens, bagMap, e, "Tagged")
			if err != nil {
				t.Errorf("Error writing tagged for %s: %s", file.Name(), err)
				continue
			}

			if verbose {
				err = compareFiles(taggedDir, d, file.Name())
				if err != nil {
					t.Errorf("Error comparing tagged for %s: %s", file.Name(), err)
				}
			}

		}
	}
}

func compareFiles(dir, d, fName string) error {

	actFile := path.Join(dir, "actual", d, fName)
	expFile := path.Join(dir, "expected", d, fName)

	actBytes, err := ioutil.ReadFile(actFile)
	if err != nil {
		return err
	}

	expBytes, err := ioutil.ReadFile(expFile)
	if err != nil {
		return err
	}

	if strings.Compare(string(actBytes), string(expBytes)) != 0 {
		return fmt.Errorf("Actual vs expected failed for %q", fName)
	}

	return err
}

func writeParsed(dir, d, fName string, tokens []parser.Token, e *env.Env) error {

	outFile := path.Join(dir, "actual", d, fName)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	var toks []string
	dn := e.DialectName()
	fn := e.InputFile()

	toks = append(toks, "Parsed")
	toks = append(toks, fmt.Sprintf("InputFile   %s", fn))
	toks = append(toks, fmt.Sprintf("Dialect     %s", dn))
	toks = append(toks, "")

	for _, t := range tokens {
		toks = append(toks, t.String())
	}

	_, err = f.Write([]byte(strings.Join(toks, "\n") + "\n"))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return err
}

func writeTagged(dir, d, fName string, m []FmtToken, bagMap map[string]TokenBag, e *env.Env, label string) error {

	outFile := path.Join(dir, "actual", d, fName)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	var toks []string

	var foldNames = map[int]string{
		dialect.FoldLower: "FoldLower",
		dialect.FoldUpper: "FoldUpper",
		dialect.NoFolding: "NoFolding",
		env.UpperCase:     "UpperCase",
		env.LowerCase:     "LowerCase",
		env.DefaultCase:   "DefaultCase",
		env.NoCase:        "NoCase",
	}

	dn := e.DialectName()

	dc := e.CaseFolding()
	dcn, _ := foldNames[dc]

	ic := e.IdentCase()
	icn, _ := foldNames[ic]

	kc := e.KeywordCase()
	kcn, _ := foldNames[kc]

	tc := e.KeywordCase()
	tcn, _ := foldNames[tc]

	fn := e.InputFile()

	toks = append(toks, label)
	toks = append(toks, fmt.Sprintf("InputFile    %s", fn))
	toks = append(toks, fmt.Sprintf("Dialect      %s", dn))
	toks = append(toks, fmt.Sprintf("FoldingCase  %s", dcn))
	toks = append(toks, fmt.Sprintf("KeywordCase  %s", kcn))
	toks = append(toks, fmt.Sprintf("IdentCase    %s", icn))
	toks = append(toks, fmt.Sprintf("DatatypeCase %s", tcn))
	toks = append(toks, "")

	for _, t := range m {
		toks = append(toks, fmt.Sprintf("                     %s", t.String()))
	}

	keys := make([]string, 0, len(bagMap))

	for key := range bagMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {

		toks = append(toks, "")

		bagId := bagMap[key].id
		bagType := nameOf(bagMap[key].typeOf)
		for _, t := range bagMap[key].tokens {
			ts := t.String()
			toks = append(toks, fmt.Sprintf("%6d %-12s: %s", bagId, bagType, ts))
		}
	}

	if toks[len(toks)-1] != "" {
		toks = append(toks, "")
	}

	_, err = f.Write([]byte(strings.Join(toks, "\n")))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return err
}
