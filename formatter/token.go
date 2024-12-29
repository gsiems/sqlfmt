package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

type CmtToken struct {
	id         int    // the ID of the token
	categoryOf int    // the category of token
	typeOf     int    // the type of token
	vSpace     int    // the count of line-feeds (vertical space) preceding the token
	indents    int    // the count of indentations preceding the token
	hSpace     string // the non-indentation horizontal white-space preceding the token
	value      string // the non-white-space text of the token
}

func (t *CmtToken) AdjustIndents(i int) {
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

type FmtToken struct {
	id          int        // the ID of the token
	categoryOf  int        // the category of token
	typeOf      int        // the type of token
	vSpace      int        // the count of line-feeds (vertical space) preceding the token
	indents     int        // the count of indentations preceding the token
	hSpace      string     // the non-indentation horizontal white-space preceding the token
	value       string     // the non-white-space text of the token
	vSpaceOrig  int        // the original preceding vertical white-space value as parsed
	hSpaceOrig  string     // the original preceding horizontal white-space value as parsed
	trlComments []CmtToken // Trailing (end of line) comment(s)
	ledComments []CmtToken // Leading comments
}

// AsUpper returns the token value as upper-case, mostly for comparison purposes
func (t *FmtToken) AsUpper() string {
	return strings.ToUpper(t.value)
}

func (t *FmtToken) IsBag() bool {
	switch t.typeOf {
	case DNFBag, DCLBag, DDLBag, DMLBag, DMLCaseBag, PLxBag, PLxBody, CommentOnBag:
		return true
	}
	return false
}
func (t *FmtToken) IsUnpackedBag() bool {
	switch t.typeOf {
	case UnpackedBag:
		return true
	}
	return false
}

func (t *FmtToken) HasLeadingComments() bool {
	return len(t.ledComments) > 0
}

func (t *FmtToken) HasTrailingComments() bool {
	return len(t.trlComments) > 0
}

func (t *FmtToken) AddLeadingComment(toks ...CmtToken) {
	for _, tok := range toks {
		t.ledComments = append(t.ledComments, tok)
	}
}

func (t *FmtToken) AddTrailingComment(toks ...CmtToken) {
	for _, tok := range toks {
		t.trlComments = append(t.trlComments, tok)
	}
}

func (t *FmtToken) IsCodeComment() bool {
	return t.categoryOf == parser.Comment
}

func (t *FmtToken) IsDatatype() bool {
	return t.categoryOf == parser.Datatype
}

func (t *FmtToken) IsIdentifier() bool {
	return t.categoryOf == parser.Identifier
}

func (t *FmtToken) IsKeyword() bool {
	return t.categoryOf == parser.Keyword
}

func (t *FmtToken) IsLabel() bool {
	return t.categoryOf == parser.Label
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

func (t *FmtToken) IsDMLCaseBag() bool {
	return t.typeOf == DMLCaseBag
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

	for idx, _ := range t.trlComments {
		t.trlComments[idx].AdjustIndents(i)
	}
	for idx, _ := range t.ledComments {
		t.ledComments[idx].AdjustIndents(i)
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

	switch t.value {
	case ",", "..":
		t.hSpace = ""
		return
	}
	switch pTok.value {
	case "..":
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
	//case t.id == 0:
	//	// very first token
	//	t.vSpace = 0
	case ensureVSpace:
		t.EnsureVSpace()
	case honorVSpace:
		t.HonorVSpace()
	default:
		t.vSpace = 0
	}
}

func (t *FmtToken) EnsureVSpace() {
	switch t.vSpaceOrig {
	case 0:
		t.vSpace = 1
	case 1, 2:
		t.vSpace = t.vSpaceOrig
	default:
		// 3 or more...
		t.vSpace = 2
	}
}

func (t *FmtToken) SetVSpace(i int) {
	t.vSpace = i
}

func (t *FmtToken) SetIndents(i int) {
	t.indents = i
}

func (t *FmtToken) HonorVSpace() {
	switch t.vSpaceOrig {
	case 0, 1, 2:
		t.vSpace = t.vSpaceOrig
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
	}
	t.SetLower()
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
		parser.Data:             "Data",

		// Token bag types/categories
		DNFBag:       "DNFBag",
		DCLBag:       "DCLBag",
		DDLBag:       "DDLBag",
		DMLBag:       "DMLBag",
		DMLCaseBag:   "DMLCaseBag",
		PLxBag:       "PLxBag",
		PLxBody:      "PLxBody",
		CommentOnBag: "CommentOnBag",
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

	return fmt.Sprintf("%6d %-12s: %-14s (%2d, %2d, %2d) (%2d %2d) [%s]",
		t.id, cName, tName,
		t.vSpace, t.indents, len(t.hSpace),
		t.vSpaceOrig, len(t.hSpaceOrig),
		t.value)
}
