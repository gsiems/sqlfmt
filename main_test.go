package main

import (
	"fmt"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"
)

/*
for each file in test/input/*sql
    - read and parse the file
    - read the like named file in test/expected
    - the compare of the two should match
*/
func TestSQLFiles(t *testing.T) {

	inputDir := "testdata/input"
	expectedDir := "testdata/expected"

	files, err := ioutil.ReadDir(inputDir)
	if err != nil {
		t.Errorf(fmt.Sprintf("%s", err))
	}

	for _, file := range files {
		// Ensure that it is a *.sql file
		if !strings.HasSuffix(file.Name(), ".sql") {
			continue
		}

		inputFile := inputDir + "/" + file.Name()
		expectedFile := expectedDir + "/" + file.Name()

		input, err := readInput(inputFile)
		if err != nil {
			t.Errorf(fmt.Sprintf("%s", err))
		}

		// There may not be an expected file, in which case the input
		// is also the expected result
		expected, err := readInput(expectedFile)
		if err != nil {
			expected = input
		}

		// Extract the parsing args from the first line of the input
		// and determine which dialect to use
		l1 := strings.SplitN(input, "\n", 2)[0]
		args := strings.Split(strings.Replace(l1, "-", "", 2), ",")

		*indentSz = 4
		*dialectName = "none"
		*keyCase = "upper"
		*preserveCase = false
		*preserveQuotes = false
		ident = "    "

		for i := 0; i < len(args); i++ {
			kv := strings.SplitN(args[i], ":", 2)
			if len(kv) > 1 {
				key := strings.Trim(kv[0], " ")
				value := strings.Trim(kv[1], " ")

				fmt.Printf("%q:%q\n", key, value)

				switch key {
				case "indent":
					*indentSz, _ = strconv.Atoi(value)
				case "dialect", "d":
					*dialectName = value
				case "k":
					*keyCase = value
				case "p":
					*preserveCase = true
				case "q":
					*preserveQuotes = true
				}
			}
		}

		dialect = resolveDialect(*dialectName)

		// bogus code just to keep things going
		result := expected + fmt.Sprintf("\n-- dialect id: %d\n", dialect) // make the compare fail

		if strings.Compare(result, expected) != 0 {
			t.Errorf("Comparison of %s failed", inputFile)

			// Output the results so they may be investigated/compared
			target := "testdata/failed/" + file.Name()

			err = writeOutput(target, result)
			if err != nil {
				t.Errorf(fmt.Sprintf("writeOutput failed: %q", err))
			}
		}
	}
}
