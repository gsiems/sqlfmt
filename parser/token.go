package parser

import (
	"errors"
	"fmt"
	"strings"
)

const (
	// NullItem indicates an undefined, or non-existent, token or other item
	NullItem = 0
	////////////////////////////////////////////////////////////////////
	// Token categories and types
	Other = iota + 300
	// WhiteSpace is a string of one or more white-space characters
	WhiteSpace
	// Identifier is a string that appears to be an identifier (or SQL keyword)
	Identifier
	// Numeric is a string that appears to represent a numeric value
	Numeric
	// Comment is any comment
	Comment
	// LineComment is an SQL end of line comment
	LineComment
	// PoundLineComment is an SQL end of line comment per MariaDB and MySQL
	PoundLineComment
	// BlockComment is an SQL block comment
	BlockComment
	// SingleQuoted is a single quoted string
	SingleQuoted
	// DoubleQuoted is a double quoted string
	DoubleQuoted
	// BacktickQuoted is a string enclosed in back-ticks '`blah blah blah`'
	BacktickQuoted
	// BracketQuoted is a string enclosed in square brackets '[blah blah blah]`
	BracketQuoted
	// Label is a string that indicates a PL label (for Oracle
	//  and PostgreSQL this means "enclosed in double greater that/less
	//  than symbols '<< blah_blah_blah >>'"). For MySQL, MS-SQL, and
	//  MariaDB this is an identifier followed by a colon 'blah_blah:'
	Label
	// Keyword is a string that matches an SQL (or PL) keyword
	Keyword
	// Operator is a string that appears to be an operator
	Operator
	// BindParameter
	BindParameter
	//
	OpenParen
	CloseParen
	OpenBracket
	CloseBracket
	OpenBrace
	CloseBrace
	Comma
	SemiColon
	//
	Datatype
	Literal
	String

	Punctuation
	End
	// Data is non-parsed data payloads (e.g. plain-text data portion of COPY from pg_dump)
	Data
)

// Token provides a single token with type information
type Token struct {
	id         int
	categoryOf int      // the category of token
	typeOf     int      // the type of token
	vSpace     int      // the count of line-feeds (vspace) preceding the token
	hSpace     string   // the horizontal white-space preceding the token
	buff       []string // the non-white-space text of the token
}

// NewToken creates and initializes a new token
func NewToken(value string, tokentype int) (Token, error) {
	var t Token

	var tt int
	// The default token type
	if tokentype < 0 {
		tt = WhiteSpace
	} else {
		tt = tokentype
	}

	tn := t.name(tt)

	if tn == "" {
		return t, errors.New("Invalid token type specified")
	}

	err := t.SetType(tt)
	if err != nil {
		return t, err
	}

	//t.SetLeadingSpace(leadingspace)
	t.buff = append(t.buff, value)

	return t, nil
}

// SetLeadingSpace sets the vertical and horizontal white space preceding the token
func (t *Token) SetLeadingSpace(s string) {
	lines := strings.Split(s, "\n")
	t.vSpace = len(lines) - 1
	t.hSpace = lines[len(lines)-1]
}

// SetHSpace sets the white-space preceding the token
func (t *Token) SetHSpace(s string) {
	t.hSpace = s
}

// HSpace returns the horizontal white space preceding the token
func (t *Token) HSpace() (s string) {
	return t.hSpace
}

// VSpace returns the amount of vertical white space (number of new lines) preceding the token
func (t *Token) VSpace() int {
	return t.vSpace
}

// SetVSpace sets the amount of vertical white space (number of new lines) preceding the token
func (t *Token) SetVSpace(vSpace int) {
	if vSpace >= 0 {
		t.vSpace = vSpace
	}
}

// Length returns the length of the token value
func (t *Token) Length() int {
	return len(t.Value())
}

// SetCategory sets the type of the token
func (t *Token) SetCategory(tc int) error {
	switch tc {
	case Comment,
		Datatype,
		Identifier,
		Keyword,
		Punctuation,
		String:

		t.categoryOf = tc
	default:
		return errors.New("Not a valid token category")
	}
	return nil
}

// SetType sets the type of the token
func (t *Token) SetType(tt int) error {
	switch tt {
	case BindParameter,
		Label,
		Numeric,
		Operator,
		Other,
		NullItem,
		WhiteSpace:
		t.typeOf = tt
		t.categoryOf = Other
	case BlockComment,
		LineComment,
		PoundLineComment:
		t.typeOf = tt
		t.categoryOf = Comment
	case BacktickQuoted,
		BracketQuoted,
		DoubleQuoted,
		SingleQuoted:
		t.typeOf = tt
		t.categoryOf = String
	case Datatype:
		t.typeOf = tt
		t.categoryOf = tt
	case OpenParen,
		CloseParen,
		OpenBracket,
		CloseBracket,
		OpenBrace,
		CloseBrace,
		Comma,
		SemiColon:
		t.typeOf = tt
		t.categoryOf = Punctuation
	case Identifier:
		t.typeOf = tt
		t.categoryOf = Identifier
	case Keyword,
		End:
		t.typeOf = tt
		t.categoryOf = Keyword
	case Data:
		t.typeOf = tt
		t.categoryOf = Data
	default:
		return errors.New("Not a valid token type")
	}
	return nil
}

// String returns the formatted type and value of the token
func (t *Token) String() string {
	return fmt.Sprintf("%6d %-12s: %-12s (%2d, %2d) [%s]",
		t.id, t.CategoryName(), t.TypeName(), t.VSpace(), len(t.hSpace), t.Value())
}

func (t *Token) name(i int) string {

	var names = map[int]string{
		//NullItem:         "NullItem",
		Other:            "Other",
		WhiteSpace:       "WhiteSpace",
		Identifier:       "Identifier",
		Numeric:          "Numeric",
		Comment:          "Comment",
		LineComment:      "LineComment",
		PoundLineComment: "PoundLineComment",
		BlockComment:     "BlockComment",
		SingleQuoted:     "SingleQuoted",
		DoubleQuoted:     "DoubleQuoted",
		BacktickQuoted:   "BacktickQuoted",
		BracketQuoted:    "BracketQuoted",
		Label:            "Label",
		Keyword:          "Keyword",
		Operator:         "Operator",
		BindParameter:    "BindParameter",
		OpenParen:        "OpenParen",
		CloseParen:       "CloseParen",
		OpenBracket:      "OpenBracket",
		CloseBracket:     "CloseBracket",
		OpenBrace:        "OpenBrace",
		CloseBrace:       "CloseBrace",
		Comma:            "Comma",
		SemiColon:        "SemiColon",
		Datatype:         "Datatype",
		Literal:          "Literal",
		String:           "String",
		Punctuation:      "Punctuation",
		End:              "End",
		Data:             "Data",
	}

	if name, ok := names[i]; ok {
		return name
	}
	return ""
}

// Category returns the category of the token
func (t *Token) Category() int {
	return t.categoryOf
}

// CategoryName returns the name of the category of the token
func (t *Token) CategoryName() string {
	return t.name(t.categoryOf)
}

// Type returns the type of the token
func (t *Token) Type() int {
	return t.typeOf
}

// TypeName returns the name of the type of the token
func (t *Token) TypeName() string {
	return t.name(t.typeOf)
}

// Id returns the id of the token
func (t *Token) Id() int {
	return t.id
}

// SetId sets the id of the token
func (t *Token) SetId(id int) {
	t.id = id
}

// Value returns the value of the token
func (t *Token) Value() string {
	return strings.Join(t.buff, "")
}

// SetValue sets the value of the token
func (t *Token) SetValue(s string) {
	t.buff = []string{s}
}

// WriteString appends a string to the tokens value and returns the number of bytes written
func (t *Token) WriteString(s string) (int, error) {
	t.buff = append(t.buff, s)
	return len(s), nil
}
