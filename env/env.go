package env

import (
	"strconv"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
)

const (
	////////////////////////////////////////////////////////////////////
	// Case folding
	UpperCase = iota + 100
	LowerCase
	DefaultCase
	NoCase
)

type Env struct {
	keywordCase     int    // Indicates whether to upper-case, lower-case, or leave keywords
	indentString    string // The character string used for indentation
	inputFile       string // The file to read from
	outputFile      string // The file to write to
	formatCode      bool   // Indicate if there should be any formatting performed or not
	preserveQuoting bool   // Preserve quoted identifiers (default is to unquote identifiers when possible)
	wrapLongLines   bool   // Indicates if line-wrapping should be performed on long lines
	minLineLength   int    // The suggested minimum line length before line-wrapping is triggered
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
	e.wrapLongLines = true
	e.minLineLength = 40
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
	case "minlinelength":
		e.SetMinLineLength(v)
	case "maxlinelength":
		e.SetMaxLineLength(v)
	}
}

func (e *Env) SetBool(k string, v bool) {
	switch strings.ToLower(k) {
	case "preservequoting":
		e.preserveQuoting = v
	case "linewrapping", "wraplonglines", "wrapping":
		e.SetWrapLongLines(v)
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

func StrToCase(v string) int {
	switch strings.ToLower(v) {
	case "foldupper", "uppercase", "upper":
		return UpperCase
	case "foldlower", "lowercase", "lower":
		return LowerCase
	case "nofolding", "nocase", "none":
		return NoCase
	}
	// Default to DefaultCase
	return DefaultCase
}

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
	c := StrToCase(v)
	switch c {
	case DefaultCase:
		switch e.CaseFolding() {
		case dialect.FoldLower, dialect.FoldUpper:
			e.keywordCase = UpperCase
		default:
			e.keywordCase = NoCase
		}
	default:
		e.keywordCase = c
	}
}

// Line Length /////////////////////////////////////////////////////////

//// Minimum Line Length

func (e *Env) MinLineLength() int {
	if e.wrapLongLines {
		return e.minLineLength
	}
	return 0
}

func (e *Env) SetMinLineLength(v int) {
	switch {
	case v+e.MinLineLength() >= e.maxLineLength:
		e.maxLineLength = v + e.MinLineLength()
		e.minLineLength = v
		e.wrapLongLines = true
	case v < 40:
	// let's not go there either
	default:
		e.minLineLength = v
		e.wrapLongLines = true
	}
}

//// Maximum Line Length

func (e *Env) MaxLineLength() int {
	if e.wrapLongLines {
		return e.maxLineLength
	}
	return 0
}

func (e *Env) SetMaxLineLength(v int) {
	switch {
	case v-e.MinLineLength() <= e.minLineLength:
	// let's not go there
	case v < 72:
	// let's not go there either
	default:
		e.maxLineLength = v
		e.wrapLongLines = true
	}
}

//// No wrapping

func (e *Env) WrapLongLines() bool {
	return e.wrapLongLines
}

func (e *Env) SetWrapLongLines(v bool) {
	e.wrapLongLines = v
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
			case "linewrapping", "wraplonglines", "wrapping":
				e.SetWrapLongLines(true)
			case "disableformatting", "donotformat", "noformatting":
				e.formatCode = false
			case "enableformatting":
				e.formatCode = true
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
			case "minlinelength", "ml":
				if s, err := strconv.Atoi(v); err == nil {
					e.SetMinLineLength(s)
				}
			case "maxlinelength", "xl":
				if s, err := strconv.Atoi(v); err == nil {
					e.SetMaxLineLength(s)
				}
			case "preservequoting":
				switch strings.ToLower(v) {
				case "off", "false", "f":
					e.preserveQuoting = false
				default:
					e.preserveQuoting = true
				}
			case "linewrapping", "wraplonglines", "wrapping":
				switch strings.ToLower(v) {
				case "off", "false", "f":
					e.SetWrapLongLines(false)
				default:
					e.SetWrapLongLines(true)
				}
			}
		}
	}
}
