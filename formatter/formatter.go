package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

const (
	////////////////////////////////////////////////////////////////////
	// Token Bag Categories and types
	DNFBag       = iota + 400 // A bag of Do-Not-Format tokens (Postgres Pl/Perl, PL/Python, PL/Tcl, COPY, etc.)
	DCLBag                    // A bag of DCL tokens
	DDLBag                    // A bag of DDL tokens
	DMLBag                    // A bag of DML tokens
	PLxBag                    // a bag of function/procedure/package tokens
	PLxBody                   // A bag of function/procedure/package body tokens
	CommentOnBag              // A bag of "COMMENT ON ..." tokens
)

func tagBags(e *env.Env, m []FmtToken) (map[string]TokenBag, []FmtToken) {

	bagMap := make(map[string]TokenBag)

	remainder := tagCommentOn(e, m, bagMap)
	remainder = tagDCL(e, remainder, bagMap)
	remainder = tagDML(e, remainder, bagMap)

	// TODO: for now at least. need to revisit once other DBs (especially
	// Oracle) are better sorted
	switch e.Dialect() {
	case dialect.PostgreSQL:
		remainder = tagPLx(e, remainder, bagMap)
	case dialect.Oracle:
	// nada
	default:
		remainder = tagDDL(e, remainder, bagMap)
	}

	// Check for warnings and errors
	var warnings []string // list of (non-fatal) warnings found
	var errors []string   // list of (fatal) errors found

	for _, bag := range bagMap {

		if len(bag.warnings) > 0 {
			warnings = append(warnings, bag.warnings...)
		}
		if len(bag.errors) > 0 {
			errors = append(errors, bag.errors...)
		}

		parensDepth := 0

		for _, t := range bag.tokens {
			switch t.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}
		}

		if parensDepth != 0 {

			label := ""
			switch bag.typeOf {
			case CommentOnBag: // should really never, ever, ever happen unless there was some fat-fingering going on
				label = "COMMENT ON statement"
			case DCLBag:
				label = "DCL statement"
			case DDLBag:
				label = "DDL statement"
			case DMLBag:
				label = "DML statement"
			case PLxBag, PLxBody:
				label = "PL code"
			}

			if bag.forObj != "" {
				errors = append(errors, fmt.Sprintf("%d unbalanced parenthesis found while parsing %s for %s", parensDepth, label, bag.forObj))
			} else {
				errors = append(errors, fmt.Sprintf("%d unbalanced parenthesis found while parsing %s", parensDepth, label))
			}
		}
	}

	return bagMap, remainder
}

func prepParsed(e *env.Env, parsed []parser.Token) (cleaned []FmtToken) {

	dbdialect := dialect.NewDialect(e.DialectName())

	identCase := e.IdentCase()
	dtCase := e.DatatypeCase()
	foldingCase := dbdialect.CaseFolding()
	var pKwVal string // The upper case value of the previous keyword token

	// 1. Give each token a unique ID.
	// 2. Review the tokens to unquote those identifiers as may be unquoted
	// 3. Perform case folding of identifiers
	// 4. Adjust the token type as needed
	// 5. Adjust the max vertical space allowed
	for id, cTok := range parsed {

		tText := cTok.Value()
		tType := cTok.Type()
		tCategory := cTok.Category()

		switch tCategory {
		case parser.Identifier:

			if !e.PreserveQuoting() {
				tryUnquoting := false
				switch tType {
				case parser.DoubleQuoted:
					tryUnquoting = true
				case parser.BracketQuoted:
					switch e.Dialect() {
					case dialect.MSSQL, dialect.SQLite:
						tryUnquoting = true
					}
				case parser.BacktickQuoted:
					switch e.Dialect() {
					case dialect.MariaDB, dialect.MySQL, dialect.SQLite:
						tryUnquoting = true
					}
				}

				if tryUnquoting {

					// Determine if the quoted identifier can be unquoted.
					// Unquote the identifier for testing
					tTest := tText[1 : len(tText)-1]

					// IIF the unquoted token is still a valid identifier (no funky chars)
					// AND is not a reserved word
					//if dbdialect.IsIdentifier(tTest) && !dbdialect.IsReservedKeyword(tTest) {
					if dbdialect.IsIdentifier(tTest) && !dbdialect.IsKeyword(tTest) {

						// if the folding is upper AND the token is upper then the token can be unquoted.
						// if the folding is lower AND the token is lower then the token can be unquoted.
						switch foldingCase {
						case dialect.FoldUpper:
							if tTest == strings.ToUpper(tTest) {
								tText = tTest
								tType = parser.Identifier
							}
						case dialect.FoldLower:
							if tTest == strings.ToLower(tTest) {
								tText = tTest
								tType = parser.Identifier
							}
						}
					}
				}
			}
		}

		switch tType {
		case parser.Identifier:
			switch pKwVal {
			case "LANGUAGE":
			// nada
			default:
				tText = strings.ToLower(tText)
			}
		case parser.Datatype, parser.Keyword:
			tText = strings.ToLower(tText)
		}

		cleaned = append(cleaned, FmtToken{
			id:         id,
			categoryOf: tCategory,
			typeOf:     tType,
			value:      tText,
			vSpaceOrig: cTok.VSpace(),
			hSpaceOrig: cTok.HSpace(),
		})

		if tCategory == parser.Keyword {
			pKwVal = strings.ToUpper(tText)
		}
	}

	return cleaned
}

func formatBags(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// remember that bagMap is a pointer to the map, not a copy of the map

	var mainTokens []FmtToken
	parensDepth := 0
	var pTok FmtToken // The previous token

	for _, cTok := range m {

		switch {
		case cTok.IsBag():
			formatBag(e, bagMap, cTok.typeOf, cTok.id, parensDepth)
		case cTok.IsCodeComment():
			cTok = formatCodeComment(e, cTok, parensDepth)
		default:
			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}
			cTok.AdjustVSpace(false, true)
			if cTok.vSpace == 0 {
				cTok.AdjustHSpace(e, pTok)
			}
		}

		pTok = cTok
		mainTokens = append(mainTokens, cTok)
	}

	return mainTokens
}

func formatBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	// remember that bagMap is a pointer to the map, not a copy of the map

	key := bagKey(bagType, bagId)

	if b, ok := bagMap[key]; ok {

		switch b.typeOf {
		case DCLBag:
			formatDCLBag(e, bagMap, b.typeOf, b.id, baseIndents)
		case DDLBag, CommentOnBag:
			formatDDLBag(e, bagMap, b.typeOf, b.id, baseIndents)
		//case DMLBag:
		//	formatDMLBag(e, bagMap, b.typeOf, b.id, baseIndents)
		case PLxBag, PLxBody:
			formatPLxBag(e, bagMap, b.typeOf, b.id, baseIndents)
		case DNFBag:
			// nada
		}
	}
}

func untagBags(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	tl1 := m
	guard := 0

	// Iterate for as long as there are mapped bags to accommodate nested DML and PL
	found := true
	for found {
		found = false

		tl2 := make([]FmtToken, 0)

		for _, cTok := range tl1 {

			if cTok.IsBag() {

				// get the key
				// look up the key in the map
				// copy the tokens from the map
				key := bagKey(cTok.typeOf, cTok.id)

				tb, ok := bagMap[key]
				if ok {
					tl2 = append(tl2, tb.tokens...)
					found = true
				} else {
					tl2 = append(tl2, cTok)
				}

			} else {
				tl2 = append(tl2, cTok)
			}
		}

		guard++
		if guard > 20 {
			// 1. Do we want to retain the guard?
			// 2. Is 20 a reasonable guard limit?
			// 3. If the guard is exceeded then should we log this?
			found = false
		}

		tl1 = tl2

	}
	return tl1
}

func combineTokens(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) string {

	tokens := untagBags(m, bagMap)

	var z []string

	for _, tc := range tokens {
		if tc.vSpace > 0 {
			z = append(z, strings.Repeat("\n", tc.vSpace))
			if tc.indents > 0 {
				z = append(z, strings.Repeat(e.Indent(), tc.indents))
			}
		} else if tc.hSpace != "" {
			z = append(z, tc.hSpace)
		}

		z = append(z, tc.value)
	}

	z = append(z, "\n")

	return strings.Join(z, "")
}
