package main

import (
	"bufio"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"

	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/formatter"
)

var (
	indentSz       = flag.Int("indent", 4, "")
	maxLineLen     = flag.Int("l", 120, "")
	configFile     = flag.String("c", "", "")
	dialectName    = flag.String("d", "standard", "")
	inputFile      = flag.String("i", "", "")
	outputFile     = flag.String("o", "", "")
	keyCase        = flag.String("k", "upper", "")
	tupleWrapping  = flag.String("t", "none", "")
	preserveQuotes = flag.Bool("q", false, "")
	version        = flag.Bool("version", false, "")
)

func main() {
	rc := runapp()
	os.Exit(rc)
}

func runapp() (rc int) {
	flag.Usage = func() {
		fmt.Fprint(os.Stderr, `usage: sqlfmt [flags]

  -c        the configuration file to read
  -d        the SQL dialect of the input (default is standard) (standard, postgres, sqlite, mariadb, mssql, mysql, oracle)
  -indent   number of spaces to indent (default is 4), set to 0 to use tabs
  -i        the file to read (defaults to stdin)
  -k        keywords case (default is upper) (upper,lower)
  -l        max line length (defaut is 120)
  -o        the file to write to (defaults to stdout)
  -q        preserve quoted identifiers (default is to unquote identifiers when possible)
  -t        multi-tuple wrapping for values statements (default is none) (all, long, none)
  -version  display the version information
`)
	}
	flag.Parse()

	if *version {
		fmt.Println("Version 2025.01.30")
		return 0
	}

	input, err := readInput(*inputFile)
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("%s while reading input %s", err, *inputFile))
		return 1
	}

	e := env.NewEnv()

	////////////////////////////////////////////////////////////////////
	// Read the config file if specified/found
	if *configFile != "" {
		cfg, err := readInput(*configFile)
		if err != nil {
			fmt.Fprint(os.Stderr, fmt.Sprintf("%s while reading config %s", err, *configFile))
			return 1
		}

		lines := strings.Split(cfg, "\n")

		for _, line := range lines {
			line = strings.TrimSpace(line)
			if len(line) == 0 {
				continue
			}
			if strings.HasPrefix(line, "#") {
				continue
			}

			p := strings.Split(line, "=")
			if len(p) != 2 {
				continue
			}

			k := strings.ToLower(strings.TrimSpace(p[0]))
			v := strings.TrimSpace(p[1])

			switch k {
			case "dialect":
				*dialectName = v

			case "indentsize":
				if s, err := strconv.Atoi(v); err == nil {
					*indentSz = s
				}

			case "keywordcase":
				*keyCase = v

			case "maxlinelength":
				if s, err := strconv.Atoi(v); err == nil {
					*maxLineLen = s
				}

			case "preservequoting":
				switch strings.ToLower(v) {
				case "on", "true", "t":
					*preserveQuotes = true
				default:
					*preserveQuotes = false
				}

			case "wrapmultituples":
				*tupleWrapping = strings.TrimSpace(p[1])
			}
		}
	}

	////////////////////////////////////////////////////////////////////
	e.SetMaxLineLength(*maxLineLen)
	e.SetKeywordCase(*keyCase)
	e.SetIndent(*indentSz)
	e.SetOutputFile(*outputFile)
	e.SetInputFile(*inputFile)
	e.SetDialect(*dialectName)
	e.SetPreserveQuoting(*preserveQuotes)
	e.SetMultiTupleWrapping(*tupleWrapping)

	////////////////////////////////////////////////////////////////////
	// Read the file directive if specified/found (extract the parsing args
	// from the first line of the input and determine which dialect to use, etc.)
	l1 := strings.SplitN(input, "\n", 2)[0]
	l1 = strings.TrimLeft(l1, "-#/* \t")
	if strings.HasPrefix(l1, "sqlfmt") {
		e.SetDirectives(l1)
	}

	if !e.FormatCode() {
		return 0
	}

	////////////////////////////////////////////////////////////////////
	formatted, warnStrings, errStrings := formatter.FormatInput(e, input)

	logStderr("WARNING", *inputFile, warnStrings)

	if len(errStrings) > 0 {
		logStderr("ERROR", *inputFile, errStrings)
		return 1
	}

	////////////////////////////////////////////////////////////////////
	err = writeOutput(*outputFile, formatted)
	if err != nil {
		fmt.Fprint(os.Stderr, fmt.Sprintf("%s while writing output %s", err, *outputFile))
		return 1
	}

	return 0
}

func dedupe(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

func logStderr(label, fileName string, messages []string) {

	if len(messages) > 0 {
		smp := make(map[string]bool)
		var msgs []string
		for _, s := range messages {
			if _, ok := smp[s]; !ok {
				msgs = append(msgs, s)
				smp[s] = true
			}
		}

		for _, msg := range msgs {
			fmt.Fprint(os.Stderr, fmt.Sprintf("%s: %s (%s)\n", label, msg, fileName))
		}
	}
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
