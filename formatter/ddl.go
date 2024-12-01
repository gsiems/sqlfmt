package formatter

import (
	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

// Adding features does not necessarily increase functionality -- it just makes the manuals thicker.

// tagDDL ensures that DDL commands (CREATE, ALTER, DROP) are properly tagged
func tagDDL(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// TODO

	// ASSERTION: DCL, DML, and PL have already been tagged and bagged.

	// NB that it will be necessary to also scan the PLxBody bags (for
	// PostgreSQL) in order to tag any DDL embedded in the plpgsql.

	// Note that Oracle "CREATE TYPE BODY" may cause problems.
	// TODO: Revisit this once Oracle (esp PL/SQL bagging) is better sorted.

	return tagDDLV0(e, m, bagMap)
}

func tagDDLV0(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// remember that bagMap is a pointer to the map, not a copy of the map

	var remainder []FmtToken
	var bagTokens []FmtToken

	isInBag := false
	bagId := 0
	parensDepth := 0

	pNcVal := ""      // The upper-case value of the previous non-comment token
	var pTok FmtToken // The previous token

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		addToBag := false
		canOpenBag := false
		closeBag := false
		openBag := false

		switch isInBag {
		case true:
			// Consider whether the bag should be closed.

			// * parens depth needs to be zero
			// * if the DDL is for a PL unit then close can happen after the PLxBag
			// * if the DDL is not for a PL unit then the closing ";"
			addToBag = true
			switch ctVal {
			case "(":
				// NB we only care about the parens depth if we are in a bag
				// so that when the parens depth goes negative then we know
				// to exit the bag
				parensDepth++
			case ")":
				parensDepth--
				if parensDepth < 0 {
					closeBag = true
					addToBag = false
				}
			default:
				switch {
				case cTok.IsDMLBag(), cTok.IsPLBag():
					closeBag = true
				case ctVal == ";":
					closeBag = true
				}
			}

		case false:
			// Consider the previous token data to determine if a bag could be opened
			switch pNcVal {
			case "", ";":
				canOpenBag = true
			case "/":
				canOpenBag = e.Dialect() == dialect.Oracle
			default:
				canOpenBag = pTok.IsBag()
			}
		}

		////////////////////////////////////////////////////////////////
		// If it is possible to maybe open a bag, determine if a bag
		// should be opened
		switch {
		case canOpenBag:
			switch ctVal {
			case "CREATE", "ALTER", "DROP":
				openBag = true
			}
		}

		////////////////////////////////////////////////////////////////
		// Actually process the token
		switch {
		case isInBag && closeBag:

			if addToBag {
				bagTokens = append(bagTokens, cTok)
			}

			// Close the bag
			isInBag = false

			key := bagKey(DDLBag, bagId)
			bagMap[key] = TokenBag{
				id:     bagId,
				typeOf: DDLBag,
				tokens: bagTokens,
			}

			bagId = 0
			bagTokens = nil

		case isInBag:
			bagTokens = append(bagTokens, cTok)

		case openBag:

			// Open a new bag
			isInBag = true

			bagId = cTok.id
			bagTokens = nil
			//bagTokens = []FmtToken{cTok}
			bagTokens = append(bagTokens, cTok)

			// Add a token that has the pointer to the new bag
			remainder = append(remainder, FmtToken{
				id:         cTok.id,
				categoryOf: DDLBag,
				typeOf:     DDLBag,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
			})

		default:
			// We are not currently in a bag and we aren't opening one either
			remainder = append(remainder, cTok)
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		pTok = cTok
		if !cTok.IsCodeComment() {
			pNcVal = ctVal
		}
	}

	// On the off chance that the bag wasn't closed properly (incomplete or
	// incorrect statement submitted?), ensure that no tokens are lost.
	if len(bagTokens) > 0 {
		key := bagKey(DDLBag, bagId)
		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: DDLBag,
			//forObj: forObj,
			tokens: bagTokens,
		}
	}

	return remainder
}

func formatDDLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	idxMax := len(b.tokens) - 1
	parensDepth := 0

	ddlAction := ""
	objType := ""
	lineCount := 0

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token

	//var ucKw = []string{
	//}

	for idx := 0; idx <= idxMax; idx++ {

		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		if ddlAction == "" {
			switch ctVal {
			case "CREATE", "ALTER", "DROP":
				ddlAction = ctVal
			}
		}

		switch objType {
		case "":
			switch ctVal {
			case "AGGREGATE", "CAST", "COLLATION", "COLUMN", "CONSTRAINT",
				"CONVERSION", "DATABASE", "DOMAIN", "EXTENSION", "FUNCTION",
				"INDEX", "LANGUAGE", "POLICY", "PROCEDURE",
				"PUBLICATION", "ROLE", "ROUTINE", "RULE", "SCHEMA", "SEQUENCE",
				"SERVER", "STATISTICS", "SUBSCRIPTION", "TABLE", "TABLESPACE",
				"TRIGGER", "TYPE", "VIEW":

				objType = ctVal

			case "ACCESS", "EVENT", "FOREIGN", "LARGE", "MATERIALIZED",
				"OPERATOR", "TRANSFORM":

				if idx < idxMax {
					nTok := b.tokens[idx+1]
					switch ctVal + " " + nTok.AsUpper() {

					case "ACCESS METHOD", "EVENT TRIGGER", "FOREIGN TABLE",
						"LARGE OBJECT", "MATERIALIZED VIEW", "OPERATOR CLASS",
						"OPERATOR FAMILY", "TRANSFORM FOR":
						objType = ctVal + " " + nTok.AsUpper()

					case "FOREIGN DATA":
						objType = "FOREIGN DATA WRAPPER"
					}
				}

				if objType == "" && ctVal == "OPERATOR" {
					objType = ctVal
				}
			}
		}

		////////////////////////////////////////////////////////////////
		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		//		if cTok.IsKeyword() && !cTok.IsDatatype() {
		//			cTok.SetKeywordCase(e, []string{ctVal})
		//		}

		switch objType {
		case "TABLE":
			switch ctVal {
			case "ALTER", "CASCADE", "CONSTRAINT", "CREATE",
				"DEFAULT", "DROP", "FOREIGN", "INDEX", "KEY", "NULL", "ON",
				"OWNER", "PRIMARY", "REFERENCES", "RESTRICT", "SET", "TABLE",
				"TO", "UNIQUE", "UPDATE":
				cTok.SetKeywordCase(e, []string{ctVal})
			}

		default:
			switch parensDepth {
			case 0:
				if cTok.IsKeyword() && !cTok.IsDatatype() {
					cTok.SetKeywordCase(e, []string{ctVal})
				}
			}

			switch objType {
			case "DATABASE":
				switch ctVal {
				case "ALTER", "CASCADE", "CREATE", "DATABASE", "DROP", "EXISTS",
					"FORCE", "IF", "ON", "OWNER", "RENAME", "SET", "TO", "WITH":
					cTok.SetKeywordCase(e, []string{ctVal})
				}

				switch e.Dialect() {
				case dialect.PostgreSQL:
					switch ctVal {
					case "ALLOW_CONNECTIONS", "BUILTIN_LOCALE",
						"COLLATION_VERSION", "ICU_LOCALE", "ICU_RULES",
						"IS_TEMPLATE", "LC_COLLATE", "LC_CTYPE", "LOCALE",
						"LOCALE_PROVIDER", "OID", "STRATEGY":
						cTok.SetKeywordCase(e, []string{ctVal})
					}
				}
			}
		}

		switch ctVal {
		case "AND", "OR", "NOT":
			cTok.SetKeywordCase(e, []string{ctVal})
		}

		switch e.Dialect() {
		case dialect.PostgreSQL:
			switch ctVal {
			case "IS", "DISTINCT":
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := true
		ensureVSpace := false

		// TODO

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth

		if idx > 0 && cTok.vSpace > 0 {
			lineCount++
		}

		if lineCount > 0 {
			indents++
		}

		if cTok.vSpace > 0 {
			cTok.AdjustIndents(indents)
		} else {
			cTok.AdjustHSpace(e, pTok)
		}

		////////////////////////////////////////////////////////////////
		switch {
		case cTok.IsBag():
			formatBag(e, bagMap, cTok.typeOf, cTok.id, indents, ensureVSpace)
		case cTok.IsCodeComment():
			cTok = formatCodeComment(e, cTok, indents)
		}

		////////////////////////////////////////////////////////////////
		// Adjust the parens depth
		switch cTok.value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		}

		// Set the various "previous token" values
		pTok = cTok

		tFormatted = append(tFormatted, cTok)
	}

	//tFormatted = WrapLongLines(e, b.typeOf, tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, b.forObj, tFormatted)
}
