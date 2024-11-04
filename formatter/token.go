package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/parser"
)

type FmtToken struct {
	id            int    // the ID of the token
	categoryOf    int    // the category of token
	typeOf        int    // the type of token
	vSpace        int    // the count of line-feeds (vertical space) preceding the token
	commentVSpace int    // the count of line-feeds (vertical space) preceding the token due to comments
	indents       int    // the count of indentations preceding the token
	hSpace        string // the non-indentation horizontal white-space preceding the token
	value         string // the non-white-space text of the token
}

// AsUpper returns the token value as upper-case, mostly for comparison purposes
func (t *FmtToken) AsUpper() string {
	return strings.ToUpper(t.value)
}

func (t *FmtToken) IsBag() bool {
	switch t.typeOf {
	case DNFBag, DCLBag, DDLBag, DMLBag, PLxBag, CommentBag:
		return true
	}
	return false
}

func (t *FmtToken) IsCodeComment() bool {
	return t.categoryOf == parser.Comment
}

func (t *FmtToken) IsKeyword() bool {
	return t.categoryOf == parser.Keyword
}

func nameOf(i int) string {

	var names = map[int]string{
		parser.Other:            "Other",
		parser.WhiteSpace:       "WhiteSpace",
		parser.Identifier:       "Identifier",
		parser.Numeric:          "Numeric",
		parser.Comment:          "Comment",
		parser.LineComment:      "LineComment",
		parser.PoundLineComment: "PoundLineComment",
		parser.BlockComment:     "BlockComment",
		parser.SingleQuoted:     "SingleQuoted",
		parser.DoubleQuoted:     "DoubleQuoted",
		parser.BacktickQuoted:   "BacktickQuoted",
		parser.BracketQuoted:    "BracketQuoted",
		parser.Label:            "Label",
		parser.Keyword:          "Keyword",
		parser.Operator:         "Operator",
		parser.BindParameter:    "BindParameter",
		parser.OpenParen:        "OpenParen",
		parser.CloseParen:       "CloseParen",
		parser.Comma:            "Comma",
		parser.SemiColon:        "SemiColon",
		parser.Datatype:         "Datatype",
		parser.Literal:          "Literal",
		parser.String:           "String",
		parser.Punctuation:      "Punctuation",
		parser.End:              "End",

		// Token bag types/categories
		DNFBag:     "DNFBag",
		DCLBag:     "DCLBag",
		DDLBag:     "DDLBag",
		DMLBag:     "DMLBag",
		PLxBag:     "PLxBag",
		CommentBag: "CommentBag",
	}

	if tName, ok := names[i]; ok {
		return tName
	}
	return ""
}

// String returns the formatted type and value of the token
func (t *FmtToken) String() string {

	cName := nameOf(t.categoryOf)
	tName := nameOf(t.typeOf)

	return fmt.Sprintf("%6d %-12s: %-12s (%2d, %2d, %2d) [%s]",
		t.id, cName, tName, t.vSpace+t.commentVSpace, t.indents, len(t.hSpace), t.value)
}
