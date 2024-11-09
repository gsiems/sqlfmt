package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
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
	vSpaceOrig    int
	hSpaceOrig    string
}

// AsUpper returns the token value as upper-case, mostly for comparison purposes
func (t *FmtToken) AsUpper() string {
	return strings.ToUpper(t.value)
}

func (t *FmtToken) IsBag() bool {
	switch t.typeOf {
	case DNFBag, DCLBag, DDLBag, DMLBag, PLxBag, PLxBody, CommentBag:
		return true
	}
	return false
}

func (t *FmtToken) IsCodeComment() bool {
	return t.categoryOf == parser.Comment
}

func (t *FmtToken) IsIdentifier() bool {
	return t.categoryOf == parser.Identifier
}

func (t *FmtToken) IsKeyword() bool {
	return t.categoryOf == parser.Keyword
}

func (t *FmtToken) IsPLBag() bool {
	switch t.typeOf {
	case PLxBag, PLxBody:
		return true
	}
	return false
}

func (t *FmtToken) IsDMLBag() bool {
	switch t.typeOf {
	case DMLBag:
		return true
	}
	return false
}

func (t *FmtToken) AdjustIndents(i int) {
	switch {
	case i <= 0:
		t.indents = 0
	default:
		if t.vSpace > 0 {
			t.indents = i
			t.hSpace = ""
		} else {
			t.indents = 0
		}
	}
}

func (t *FmtToken) AdjustHSpace(e *env.Env, pTok FmtToken) {

	if t.id == 0 {
		t.hSpace = ""
		return
	}

	if t.vSpace > 0 {
		t.hSpace = ""
		return
	}

	if t.value == "," {
		t.hSpace = ""
		return
	}

	if len(pTok.value) > 0 {

		switch e.Dialect() {
		case dialect.PostgreSQL:
			if t.value == "::" || pTok.value == "::" {
				t.hSpace = ""
				return
			}
		case dialect.MSSQL:
			if pTok.value == "::" {
				t.hSpace = ""
				return
			}
		}

		if t.IsIdentifier() && strings.HasSuffix(pTok.value, ".") {
			t.hSpace = ""
			return
		}

		if pTok.IsIdentifier() && strings.HasPrefix(t.value, ".") {
			t.hSpace = ""
			return
		}

		switch string(pTok.value[len(pTok.value)-1]) + t.value {
		case "()", "(*", "*)", ".*":
			t.hSpace = ""
			return
		}
	}

	t.hSpace = " "
}

func (t *FmtToken) AdjustVSpace(ensureVSpace, honorVSpace bool) {
	switch {
	case t.id == 0:
		// very first token
		t.vSpace = 0
	case ensureVSpace:
		t.EnsureVSpace()
	case honorVSpace:
		t.HonorVSpace()
	default:
		t.vSpace = 0
	}
}

func (t *FmtToken) EnsureVSpace() {
	switch t.vSpace {
	case 0:
		t.vSpace = 1
	case 1, 2:
	// leave as is
	default:
		// 3 or more...
		t.vSpace = 2
	}
}

func (t *FmtToken) HonorVSpace() {
	switch t.vSpace {
	case 0, 1, 2:
	// leave as is
	default:
		// 3 or more...
		t.vSpace = 2
	}
}

func (t *FmtToken) SetUpper() {
	if t.value != strings.ToUpper(t.value) {
		t.value = strings.ToUpper(t.value)
	}
}

func (t *FmtToken) SetLower() {
	if t.value != strings.ToLower(t.value) {
		t.value = strings.ToLower(t.value)
	}
}

func (t *FmtToken) SetKeywordCase(e *env.Env, kWords []string) {
	switch e.KeywordCase() {
	case env.UpperCase:
		tVal := t.AsUpper()
		for _, kw := range kWords {
			if kw == tVal {
				t.SetUpper()
				return
			}
		}
		t.SetLower()
	case env.LowerCase:
		t.SetLower()
	}
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
		PLxBody:    "PLxBody",
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

	return fmt.Sprintf("%6d %-12s: %-12s (%2d, %2d, %2d) (%2d %2d %2d) [%s]",
		t.id, cName, tName, t.vSpace+t.commentVSpace, t.indents, len(t.hSpace),
		t.vSpaceOrig, t.commentVSpace, len(t.hSpaceOrig), t.value)
}
