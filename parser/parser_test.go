package parser

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestSQLFiles(t *testing.T) {

	dialects := []string{"mariadb", "mssql", "mysql", "oracle", "postgresql", "sqlite", "standard"}

	for _, d := range dialects {

		inputDir := path.Join("..", "testdata", "input", d)
		parsedDir := path.Join("..", "testdata", "parsed")

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

			p := NewParser(d)

			////////////////////////////////////////////////////////////////////////
			parsed := p.ParseStatements(input)

			err = writeParsed(parsedDir, d, file.Name(), parsed)
			if err != nil {
				t.Errorf("Error writing parsed for %s: %s", file.Name(), err)
				continue
			}

			var z []string

			for _, tc := range parsed {
				if tc.vSpace > 0 {
					z = append(z, strings.Repeat("\n", tc.vSpace))
				}
				if tc.hSpace != "" {
					z = append(z, tc.hSpace)
				}
				z = append(z, tc.Value())
			}

			if string(inBytes) != strings.Join(z, "") {
				t.Errorf("Error comparing original to re-constructed for %s", file.Name())
			}
		}
	}
}

func writeParsed(dir, d, fName string, parsed []Token) error {

	outFile := path.Join(dir, "actual", d, fName)

	f, err := os.OpenFile(outFile, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	defer f.Close()

	var toks []string
	toks = append(toks, "Parsed")
	toks = append(toks, fmt.Sprintf("InputFile   %s", fName))
	toks = append(toks, fmt.Sprintf("Dialect     %s", d))
	toks = append(toks, "")

	for _, t := range parsed {
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
