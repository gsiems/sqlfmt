package formatter

import (
	"fmt"
	"io/ioutil"
	"log"
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

	baseDir := path.Join("..", "testdata")

	//dialects := []string{"mariadb", "mssql", "mysql", "oracle", "postgresql", "sqlite", "standard"}

	dataDir := path.Join(baseDir, "input")

	rd, err := os.ReadDir(dataDir)
	if err != nil {
		t.Error(err)
		return
	}

	for _, f := range rd {

		if !f.IsDir() {
			continue
		}

		d := f.Name()

		inputDir := path.Join(dataDir, d)
		cleanedDir := path.Join(baseDir, "cleaned")
		taggedDir := path.Join(baseDir, "tagged")
		untaggedDir := path.Join(baseDir, "untagged")
		formattedDir := path.Join(baseDir, "formatted")
		outputDir := path.Join(baseDir, "output")

		files, err := ioutil.ReadDir(inputDir)
		if err != nil {
			t.Errorf("%s", err)
		}

		for _, file := range files {
			// Ensure that it is a *.sql file
			if !strings.HasSuffix(file.Name(), ".sql") {
				continue
			}

			if false {
				switch file.Name() {
				case "foo.sql":
				//case "pg_wrapping.sql":
				default:
					continue
				}
			}

			if false {
				log.Printf("\n\n%q\n", file.Name())
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

			e.SetDialect(d)

			if !e.FormatCode() {
				continue
			}

			e.SetInputFile(inputFile)

			p := parser.NewParser(d)

			////////////////////////////////////////////////////////////////////////
			// Parse the input and compare to expected
			var parsed []parser.Token
			parsed, err = p.ParseStatements(input)
			if err != nil {
				t.Errorf("Error parsing input for %s (%s)", file.Name(), err)
				continue
			}

			////////////////////////////////////////////////////////////////////////
			// "Clean" the tokens and compare to expected
			cleaned := prepParsed(e, parsed)

			err = writeCleaned(cleanedDir, d, file.Name(), cleaned, e)
			if err != nil {
				t.Errorf("Error writing cleaned for %s: %s", file.Name(), err)
				continue
			}

			if verbose {
				err = compareFiles(taggedDir, d, file.Name())
				if err != nil {
					t.Errorf("Error comparing cleaned for %s: %s", file.Name(), err)
				}
			}

			////////////////////////////////////////////////////////////////////////
			// Tag the tokens and compare to expected
			bagMap, mainTokens, _, _ := tagBags(e, cleaned)

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

			////////////////////////////////////////////////////////////////////////
			// Format the tokens and compare to expected
			fmtTokens := formatBags(e, mainTokens, bagMap)

			err = writeTagged(formattedDir, d, file.Name(), fmtTokens, bagMap, e, "Formatted")
			if err != nil {
				t.Errorf("Error writing formatted for %s: %s", file.Name(), err)
				continue
			}

			if verbose {
				err = compareFiles(formattedDir, d, file.Name())
				if err != nil {
					t.Errorf("Error comparing formatted for %s: %s", file.Name(), err)
				}
			}

			////////////////////////////////////////////////////////////////////////
			// Untag the tokens and compare to expected
			untagged := untagBags(fmtTokens, bagMap)
			unstashed := unstashComments(e, untagged)

			err = writeTagged(untaggedDir, d, file.Name(), unstashed, bagMap, e, "Untagged")
			if err != nil {
				t.Errorf("Error writing formatted for %s: %s", file.Name(), err)
				continue
			}

			if verbose {
				err = compareFiles(untaggedDir, d, file.Name())
				if err != nil {
					t.Errorf("Error comparing formatted for %s: %s", file.Name(), err)
				}
			}

			////////////////////////////////////////////////////////////////////////
			// Recombine the tokens and write the final output
			fmtStatement := combineTokens(e, unstashed)

			err = writeOutput(outputDir, d, file.Name(), fmtStatement)
			if err != nil {
				t.Errorf("Error writing final for %s: %s", file.Name(), err)
				continue
			}

			err = compareFiles(outputDir, d, file.Name())
			if err != nil {
				t.Errorf("Error comparing final for %s: %s", file.Name(), err)
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

func writeCleaned(dir, d, fName string, tokens []FmtToken, e *env.Env) error {

	outFile := path.Join(dir, "actual", d, fName)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	var toks []string
	dn := e.DialectName()
	fn := e.InputFile()

	toks = append(toks, "Cleaned")
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

	kc := e.KeywordCase()
	kcn, _ := foldNames[kc]

	fn := e.InputFile()

	toks = append(toks, label)
	toks = append(toks, fmt.Sprintf("InputFile    %s", fn))
	toks = append(toks, fmt.Sprintf("Dialect      %s", dn))
	toks = append(toks, fmt.Sprintf("FoldingCase  %s", dcn))
	toks = append(toks, fmt.Sprintf("KeywordCase  %s", kcn))

	for _, t := range m {
		toks = append(toks, fmt.Sprintf("                     %s", t.String()))
	}

	keys := make([]string, 0, len(bagMap))

	for key := range bagMap {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if len(bagMap[key].errors) > 0 {
			toks = append(toks, "ERRORS:")
			for _, t := range bagMap[key].errors {
				toks = append(toks, "    "+t)
			}
		}
	}
	toks = append(toks, "")

	for _, key := range keys {

		toks = append(toks, "")

		bagId := bagMap[key].id
		bagType := nameOf(bagMap[key].typeOf)
		for _, t := range bagMap[key].tokens {
			ts := t.String()
			lct := len(t.ledComments)
			tct := len(t.trlComments)

			toks = append(toks, fmt.Sprintf("%6d %-12s (%d, %d): %s", bagId, bagType, lct, tct, ts))
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

func writeOutput(dir, d, fName string, statement string) error {

	outFile := path.Join(dir, "actual", d, fName)
	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.Write([]byte(statement))
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	return err
}
