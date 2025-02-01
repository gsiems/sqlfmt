package env

import (
	"strconv"
	"strings"

	"github.com/gsiems/db-dialect/dialect"
)

const (
	////////////////////////////////////////////////////////////////////
	// Case folding
	UpperCase = iota + 100
	LowerCase
	DefaultCase
	NoCase
	WrapNone
	WrapAll
	WrapLong
)

type Env struct {
	keywordCase     int    // Indicates whether to upper-case, lower-case, or leave keywords
	indentString    string // The character string used for indentation
	inputFile       string // The file to read from
	outputFile      string // The file to write to
	formatCode      bool   // Indicate if there should be any formatting performed or not
	preserveQuoting bool   // Preserve quoted identifiers (default is to unquote identifiers when possible)
	wrapMultiTuples int    // Indicates how values with multiple tuples should be wrapped
	maxLineLength   int    // The suggested maximum line length after which line-wrapping is triggered
	dbdialect       dialect.DbDialect
}

func NewEnv() *Env {
	var e Env

	e.keywordCase = UpperCase
	e.indentString = "    " // 4 spaces
	e.inputFile = "-"
	e.outputFile = "-"
	e.formatCode = true
	e.preserveQuoting = false
	e.wrapMultiTuples = WrapNone
	e.maxLineLength = 120

	return &e
}

func (e *Env) SetString(k, v string) {
	switch strings.ToLower(k) {
	case "dialect", "d":
		e.SetDialect(v)
	case "keywordcase", "kwc":
		e.SetKeywordCase(v)
	case "input", "if":
		e.SetInputFile(v)
	case "output", "of":
		e.SetOutputFile(v)
	}
}

func (e *Env) SetInt(k string, v int) {
	switch strings.ToLower(k) {
	case "indentsize", "indent":
		e.SetIndent(v)
	case "maxlinelength":
		e.SetMaxLineLength(v)
	}
}

func (e *Env) SetBool(k string, v bool) {
	switch strings.ToLower(k) {
	case "preservequoting":
		e.preserveQuoting = v
	case "disableformatting":
		e.formatCode = false
	case "enableformatting":
		e.formatCode = true
	}
}

func (e *Env) FormatCode() bool {
	return e.formatCode
}

// Preserve Quoting ////////////////////////////////////////////////////

func (e *Env) PreserveQuoting() bool {
	return e.preserveQuoting
}

func (e *Env) SetPreserveQuoting(v bool) {
	e.preserveQuoting = v
}

// Database Dialect ////////////////////////////////////////////////////

func (e *Env) Dialect() int {
	if e.dbdialect == nil {
		e.dbdialect = dialect.NewDialect("StandardSQL")
	}
	return e.dbdialect.Dialect()
}

func (e *Env) DialectName() string {
	if e.dbdialect == nil {
		e.dbdialect = dialect.NewDialect("StandardSQL")
	}
	return e.dbdialect.DialectName()
}

func (e *Env) SetDialect(v string) {
	e.dbdialect = dialect.NewDialect(v)
}

// Files ///////////////////////////////////////////////////////////////

func (e *Env) SetInputFile(v string) {
	e.inputFile = v
}

func (e *Env) SetOutputFile(v string) {
	e.outputFile = v
}

func (e *Env) InputFile() string {
	return e.inputFile
}

func (e *Env) OutputFile() string {
	return e.outputFile
}

// Indentation /////////////////////////////////////////////////////////

func (e *Env) Indent() string {
	return e.indentString
}

func (e *Env) SetIndent(v int) {
	if v > 0 {
		e.indentString = strings.Repeat(" ", v)
	} else {
		e.indentString = "\t"
	}
}

// Casing //////////////////////////////////////////////////////////////

//// Identifiers

func (e *Env) CaseFolding() int {
	if e.dbdialect == nil {
		e.dbdialect = dialect.NewDialect("StandardSQL")
	}
	return e.dbdialect.CaseFolding()
}

//// Keyword Case

func (e *Env) KeywordCase() int {
	switch e.keywordCase {
	case DefaultCase:
		switch e.CaseFolding() {
		case dialect.FoldLower, dialect.FoldUpper:
			return UpperCase
		default:
			return NoCase
		}
	}
	return e.keywordCase
}

func (e *Env) SetKeywordCase(v string) {
	switch strings.ToLower(v) {
	case "foldupper", "uppercase", "upper":
		e.keywordCase = UpperCase
	case "foldlower", "lowercase", "lower":
		e.keywordCase = LowerCase
	default:
		switch e.CaseFolding() {
		case dialect.FoldLower, dialect.FoldUpper:
			e.keywordCase = UpperCase
		default:
			e.keywordCase = NoCase
		}
	}
}

// Maximum Line Length /////////////////////////////////////////////////

func (e *Env) MaxLineLength() int {
	return e.maxLineLength
}

func (e *Env) SetMaxLineLength(v int) {
	switch {
	case v < 72:
	// let's not go there
	default:
		e.maxLineLength = v
	}
}

// Multi-tuple Wrapping ////////////////////////////////////////////////

func (e *Env) WrapMultiTuples() int {
	return e.wrapMultiTuples
}

func (e *Env) SetMultiTupleWrapping(v string) {

	switch strings.ToLower(v) {
	case "all", "wrapall":
		e.wrapMultiTuples = WrapAll
	case "long", "wraplong":
		e.wrapMultiTuples = WrapLong
	default:
		e.wrapMultiTuples = WrapNone
	}
}

// File Directives /////////////////////////////////////////////////////

func (e *Env) SetDirectives(v string) {

	l1 := strings.TrimLeft(v, "-#/* \t")
	if !strings.HasPrefix(l1, "sqlfmt") {
		return
	}

	l1 = strings.Replace(l1, "sqlfmt", "", 1)

	args := strings.Split(l1, ";")

	for i := 0; i < len(args); i++ {
		kv := strings.SplitN(args[i], ":", 2)

		switch len(kv) {
		case 1:
			k := strings.Trim(kv[0], " \t")

			switch strings.ToLower(k) {
			case "preservequoting":
				e.preserveQuoting = true
			case "noformat":
				e.formatCode = false
			}

		case 2:
			k := strings.Trim(kv[0], " \t")
			v := strings.Trim(kv[1], " \t")

			switch strings.ToLower(k) {
			case "dialect", "d":
				e.SetDialect(v)
			case "keywordcase", "kwc":
				e.SetKeywordCase(v)
			case "input", "if":
				e.SetInputFile(v)
			case "output", "of":
				e.SetOutputFile(v)
			case "indentsize", "indent":
				if s, err := strconv.Atoi(v); err == nil {
					e.SetIndent(s)
				}
			case "maxlinelength", "xl":
				if s, err := strconv.Atoi(v); err == nil {
					e.SetMaxLineLength(s)
				}
			case "preservequoting":
				switch strings.ToLower(v) {
				case "on", "true", "t":
					e.preserveQuoting = true
				default:
					e.preserveQuoting = false
				}
			case "noformat":
				switch strings.ToLower(v) {
				case "off", "false", "f":
					e.formatCode = true
				default:
					e.formatCode = false
				}
			case "wrapMultiTuples":
				e.SetMultiTupleWrapping(v)
			}
		}
	}
}
