package parser

import (
	"strconv"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

/*

parser.go provides the actual parsing logic

*/

type Parser struct {
	dbdialect dialect.DbDialect
}

func NewParser(e *env.Env) *Parser {
	var p Parser

	p.dbdialect = dialect.NewDialect(e.DialectName())

	return &p
}

// ParseStatements takes a string of one or more SQL-ish looking
// statements and/or procedural SQL blocks and splits them into a list
// of word, symbol, comment, quoted string, etc. tokens. The dialect of
// the SQL being submitted is used to better tokenize the submitted string.
func (p *Parser) ParseStatements(stmts string) []Token {

	return p.tokenizeStatement(stmts)
}

// tokenizeStatement is primarily about resolving those tokens that are either
// delimited by standard start/end character strings (like comments and
// comment blocks), are white-space, or are stand-alone punctuation.
func (p *Parser) tokenizeStatement(stmts string) []Token {

	var tlRe []Token

	if stmts == "" {
		return tlRe
	}

	qiMax := len(stmts) - 1
	qi := -1
	iStart := 0
	tType := NullItem

	for qi < qiMax {
		qi++

		chr := string(stmts[qi])
		chrNext := ""
		if qi < qiMax {
			chrNext = string(stmts[qi+1])
		}

		// Dealing with an escape char?
		if chr == "\\" {
			qi++
			continue
		}

		// if we are in an *enclosed* token (has a defined start and end
		// character), check for the ending
		switch tType {
		case DoubleQuoted, SingleQuoted, BacktickQuoted, BracketQuoted:

			if p.chkTokenEnd(chr, tType) {
				if qi+1 > iStart {
					nt, _ := NewToken(string(stmts[iStart:qi+1]), tType)
					tlRe = append(tlRe, nt)
				}
				iStart = qi + 1

				if iStart < qiMax && p.isWhiteSpaceChar(string(stmts[iStart])) {
					tType = WhiteSpace
				} else {
					tType = Other
				}

			}
			continue

		case BlockComment:
			if p.chkTokenEnd(chr+chrNext, tType) {
				if qi+2 > iStart {
					nt, _ := NewToken(string(stmts[iStart:qi+2]), tType)
					tlRe = append(tlRe, nt)
				}
				iStart = qi + 2
				qi++

				if iStart < qiMax && p.isWhiteSpaceChar(string(stmts[iStart])) {
					tType = WhiteSpace
				} else {
					tType = Other
				}
			}
			continue

		case LineComment, PoundLineComment:
			if chr == "\n" {
				nt, _ := NewToken(string(stmts[iStart:qi]), tType)
				iStart = qi
				tlRe = append(tlRe, nt)
				tType = WhiteSpace
			}
			continue
		}

		// check for the beginning of an *enclosed* token
		tt2 := p.chkTokenStart(chr, chrNext)
		switch tt2 {

		case DoubleQuoted, SingleQuoted, BacktickQuoted,
			BracketQuoted, LineComment, BlockComment,
			PoundLineComment:

			if qi > iStart {
				nt, _ := NewToken(string(stmts[iStart:qi]), tType)
				tlRe = append(tlRe, nt)
				iStart = qi
			}
			tType = tt2
			continue
		}

		// Special punctuation
		switch chr {
		case "(", ")", ",", ";":

			if qi > iStart {
				nt, _ := NewToken(string(stmts[iStart:qi]), tType)
				tlRe = append(tlRe, nt)
			}

			switch chr {
			case "(":
				nt, _ := NewToken(chr, OpenParen)
				tlRe = append(tlRe, nt)

			case ")":
				nt, _ := NewToken(chr, CloseParen)
				tlRe = append(tlRe, nt)

			case ",":
				nt, _ := NewToken(chr, Comma)
				tlRe = append(tlRe, nt)

			case ";":
				nt, _ := NewToken(chr, SemiColon)
				tlRe = append(tlRe, nt)
			}

			iStart = qi + 1
			if iStart < qiMax && p.isWhiteSpaceChar(string(stmts[iStart])) {
				tType = WhiteSpace
			} else {
				tType = Other
			}
			continue
		}

		// White-space
		switch tType {
		case WhiteSpace:
			if !p.isWhiteSpaceChar(chr) {
				if qi > iStart {
					nt, _ := NewToken(string(stmts[iStart:qi]), tType)
					iStart = qi
					tlRe = append(tlRe, nt)
				}
				tType = Other
			}
			continue
		}

		if p.isWhiteSpaceChar(chr) {
			if tType != WhiteSpace {
				if qi > iStart {
					nt, _ := NewToken(string(stmts[iStart:qi]), tType)
					tlRe = append(tlRe, nt)
				}
				iStart = qi
			}
			tType = WhiteSpace
			continue
		}

		// Don't know (yet) what to do with it
		if tType != Other {
			if qi > iStart {
				nt, _ := NewToken(string(stmts[iStart:qi]), tType)
				tlRe = append(tlRe, nt)
			}
			iStart = qi
			tType = Other
		}
	}

	return p.updateTokenTypes(tlRe)
}

// updateTokenTypes takes the output of tokenizeStatement and attempts to
// resolve those tokens not resolved in the first pass by inspecting, and
// splitting, the contents of the token into multiple tokens. This also
// attempts to resolve if the *enclosed* tokens are for strings or identifiers.
func (p *Parser) updateTokenTypes(tlIn []Token) []Token {

	var tlRe []Token

	for _, tc := range tlIn {

		if tc.typeOf != Other {
			// If the token was resolved in the first pass then
			// resolve whether this is a "quoted" identifier or a string
			// literal. The standard is to use single quotes for literals
			// and double quotes for identifiers. Not all DBMS vendors got
			// the memo, or paid attention to it if they did. (Or else they
			// were trying to be compatible with the vendors that didn't
			// get the memo...)

			switch tc.typeOf {
			case DoubleQuoted:
				switch p.dbdialect.Dialect() {
				case dialect.MySQL, dialect.MariaDB:
					tc.categoryOf = String
				case dialect.PostgreSQL:
					if tc.Value() == "\"char\"" {
						tc.categoryOf = Datatype
					} else {
						tc.categoryOf = Identifier
					}
				default:
					tc.categoryOf = Identifier
				}
			case SingleQuoted:
				tc.categoryOf = String
			case BacktickQuoted, BracketQuoted:
				tc.categoryOf = Identifier
			}

			tlRe = append(tlRe, tc)
			continue
		}

		// If the token type can be determined based on the token text then set
		// the token type, push it to the queue and move on
		tt2 := p.chkTokenString(tc.Value())
		switch tt2 {
		case BindParameter, Datatype, Identifier, Keyword, Label, Numeric, Operator:
			tc.SetType(tt2)
			tlRe = append(tlRe, tc)
			continue
		}

		// Check for cases where the current token consists of a longer string
		// that would be multiple tokens were the white-space that should be was.
		// This currently consists of strings that have operators and that don't
		// have white-space on one, or both, sides
		remainder := tc.Value()
		s2 := ""

		for remainder != "" {

			s2, remainder = p.splitOnOperator(remainder)
			if s2 != "" {
				tt2 := p.chkTokenString(s2)
				switch tt2 {
				case BindParameter, Datatype, Identifier, Keyword, Numeric, Operator:
					nt, _ := NewToken(s2, tt2)
					tlRe = append(tlRe, nt)
				default:
					nt, _ := NewToken(s2, Other)
					tlRe = append(tlRe, nt)
				}
			}
		}
	}
	return p.consolidateStrings(tlRe)
}

// consolidateStrings deals with combining, as needed, adjacent
// SingleQuotedTokens or SingleQuotedTokens that have a preceding decoration
// (Identifier)
//
// Some DBs have the option of preceding a string literal with a bit of text
// that indicates how to interpret the string literal
//
//   - N'string literal',
//   - E'string literal',
//   - _utf8'string literal',
//   - etc.
//
// Also, if a single-quoted token is followed by another single-quoted token
// (with no white-space separating) then they should be combined. This could go
// on for several tokens.
//
//   - ”, 'something'   -> ”'something'   -> "'something"
//
//   - 'something', 's'  -> 'something”s'  -> "something's"
//
//   - ”, ”            -> ””            -> "'"
//
//   - MySQL can use either single or double quotes for string literals. This might
//     be a problem.
//
//   - MySQL also concatenates string literals:
//     "'to' ' ' 'do'" is the same as "'to do'" -- I can only assume that this
//     requires there to be white space between tokens?
func (p *Parser) consolidateStrings(tlIn []Token) []Token {

	var tlRe []Token

	idxMax := len(tlIn) - 1
	idx := -1

	for idx < idxMax {
		idx++

		tc := tlIn[idx]

		if idx >= idxMax {
			tlRe = append(tlRe, tc)
			break
		}

		if tc.categoryOf == String || tc.typeOf == Identifier {

			// As long as the next token is a string literal
			// with no leading white-space then we want to join them
			doContinue := true

			for doContinue && idx < idxMax {

				tcNext := tlIn[idx+1]

				if tcNext.categoryOf == String {

					tc.WriteString(tcNext.Value())

					tc.typeOf = tcNext.typeOf
					tc.categoryOf = tcNext.categoryOf
					idx++
					continue
				}
				doContinue = false
			}
		}

		if tc.typeOf == Identifier {
			tc.categoryOf = Identifier
		}
		tlRe = append(tlRe, tc)
	}

	return p.consolidateWhitespace(tlRe)
}

// consolidateWhitespace consolidates white-space tokens by folding them into
// the following token (if any).
func (p *Parser) consolidateWhitespace(tlIn []Token) []Token {

	var tlRe []Token

	leadingSpace := ""

	idxMax := len(tlIn) - 1

	for idx := 0; idx <= idxMax; idx++ {

		tc := tlIn[idx]

		if idx == idxMax && tc.typeOf == WhiteSpace {
			// Preserve any trailing whitespace
			tlRe = append(tlRe, tc)
		} else {
			if leadingSpace != "" {
				tc.SetLeadingSpace(leadingSpace)
				leadingSpace = ""
			} else if tc.typeOf == WhiteSpace {
				leadingSpace = tc.Value()
				continue
			}
			tlRe = append(tlRe, tc)
		}
	}
	return tlRe
}

// splitOnOperator takes a string and searches it for, and splits on, operators.
// The issue being that a query could have strings such as "x+y" or "a>b" which
// should be split into the proper tokens
func (p *Parser) splitOnOperator(s string) (string, string) {

	// Search for operators starting with the longest possible operator for
	// the dialect being parsed.
	// If the max operator length is greater than the length of the string to
	// parse (unlikely) then we set the max length to the string length.
	maxOperatorLen := p.dbdialect.MaxOperatorLength()
	maxLen := maxOperatorLen
	if maxLen > len(s) {
		maxLen = len(s)
	}

	pre := s
	remainder := ""

	for i := maxLen; i > 0; i-- {
		if len(s)-i < 0 {
			continue
		}

		for j := 0; j <= len(s)-i; j++ {

			var tstOp string
			if i == 1 {
				tstOp = string(s[j])
			} else {
				tstOp = s[j : j+i]
			}

			if p.dbdialect.IsOperator(tstOp) {
				if j == 0 {
					pre = tstOp
					remainder = s[len(pre):]
				} else {
					pre = s[0:j]
					remainder = s[j:]
				}
				return pre, remainder
			}
		}
	}
	return pre, remainder
}

// chkTokenStart checks the character(s) provided to determine if they are
// the start of an *enclosed* token such as a quoted string, line comment, etc.
func (p *Parser) chkTokenStart(s, s2 string) (d int) {

	switch s {
	case "\"":
		return DoubleQuoted
	case "'":
		return SingleQuoted
	case "#":
		switch p.dbdialect.Dialect() {
		case dialect.MySQL, dialect.MariaDB:
			return PoundLineComment
		}
	case "`":
		switch p.dbdialect.Dialect() {
		case dialect.MySQL, dialect.MariaDB, dialect.SQLite:
			// SQLite in compatibility mode
			return BacktickQuoted
		}
	case "[":
		switch p.dbdialect.Dialect() {
		case dialect.MSSQL, dialect.SQLite:
			// SQLite in compatibility mode
			return BracketQuoted
		}
	case "/":
		if s2 == "*" {
			return BlockComment
		}
	case "-":
		if s2 == "-" {
			return LineComment
		}
	}

	return NullItem
}

// chkTokenEnd checks the string provided to determine if it is the end of
// an *enclosed* token such as a quoted string, line comment, etc.
func (p *Parser) chkTokenEnd(s string, typeOf int) bool {

	var tokenEnd = map[int]string{
		BacktickQuoted:   "`",
		BlockComment:     "*/",
		BracketQuoted:    "]",
		DoubleQuoted:     "\"",
		LineComment:      "\n",
		PoundLineComment: "\n",
		SingleQuoted:     "'",
	}

	if te, ok := tokenEnd[typeOf]; ok {
		if s == te {
			return true
		}
	}

	return false
}

// chkTokenString checks the string provided to determine what kind of token
// it represents. Note that order of evaluation is important.
func (p *Parser) chkTokenString(s string) (d int) {

	switch true {
	case p.dbdialect.IsDatatype(s):
		return Datatype
	case p.dbdialect.IsKeyword(s):
		return Keyword
	case p.dbdialect.IsLabel(s):
		return Label
	case p.isNumericString(s):
		return Numeric
	case p.dbdialect.IsIdentifier(s):
		return Identifier
	case p.dbdialect.IsOperator(s):
		return Operator
	case p.isBindVar(s):
		return BindParameter
	}

	return NullItem
}

// isWhiteSpaceChar determines whether or not the supplied character is
// considered to be a white space character
func (p *Parser) isWhiteSpaceChar(s string) bool {
	const wsChars = " \n\r\t"
	return strings.Contains(wsChars, s)
}

// isNumericString determines whether or not the supplied string is
// considered to be a valid number
func (p *Parser) isNumericString(s string) bool {

	if s == "" {
		return false
	}

	// Numbers in scientific notation are numbers. But even scientific notation
	// should only have one E.
	if strings.Count(strings.ToUpper(s), "E") > 1 {
		return false
	}

	for _, s := range strings.Split(strings.ToUpper(s), "E") {
		_, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return false
		}
	}
	return true
}

func (p *Parser) isBindVar(s string) bool {
	// bind variables?
	// :x
	// ?
	// $x
	// other?

	if s == "?" {
		return true
	}
	if len(s) > 1 {
		if string(s[0]) == ":" && strings.Count(s, ":") == 1 {
			return true
		}
		if string(s[0]) == "$" && strings.Count(s, "$") == 1 {
			return true
		}
	}
	return false
}
