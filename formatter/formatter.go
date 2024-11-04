package formatter

import (
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

const (
	////////////////////////////////////////////////////////////////////
	// Token Bag Categories and types
	DNFBag     = iota + 400 // A bag of Do-Not-Format tokens (Postgres Pl/Perl, PL/Python, PL/Tcl, COPY, etc.)
	DCLBag                  // A bag of DCL tokens
	DDLBag                  // A bag of DDL tokens
	DMLBag                  // A bag of DML tokens
	PLxBag                  // a bag of function/procedure/package tokens
	CommentBag              // A bag of "COMMENT ON ..." tokens
)

func tagBags(e *env.Env, m []FmtToken) (map[string]TokenBag, []FmtToken) {

	bagMap := make(map[string]TokenBag)

	remainder := tagComment(e, m, bagMap)
	remainder = tagDCL(e, remainder, bagMap)
	remainder = tagDML(e, remainder, bagMap)
	remainder = tagPLx(e, remainder, bagMap)
	remainder = tagDDL(e, remainder, bagMap)

	return bagMap, remainder
}

func cleanupParsed(e *env.Env, parsed []parser.Token) (cleaned []FmtToken) {

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
				switch identCase {
				case env.UpperCase:
					tText = strings.ToUpper(tText)
				case env.LowerCase:
					tText = strings.ToLower(tText)
				}
			}
		case parser.Datatype:
			switch dtCase {
			case env.UpperCase:
				tText = strings.ToUpper(tText)
			case env.LowerCase:
				tText = strings.ToLower(tText)
			}
		}

		vSpace := cTok.VSpace()
		if vSpace > 2 {
			vSpace = 2
		}

		cleaned = append(cleaned, FmtToken{
			id:         id,
			categoryOf: tCategory,
			typeOf:     tType,
			vSpace:     vSpace,
			hSpace:     cTok.HSpace(),
			value:      tText,
		})

		if tCategory == parser.Keyword {
			pKwVal = strings.ToUpper(tText)
		}
	}

	return cleaned
}
