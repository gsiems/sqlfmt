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

			switch pTok.AsUpper() {
			case "", ";":
				canOpenBag = true
			case "THEN", "LOOP":
				// THEN and LOOP to catch DDL embedded within PL code
				canOpenBag = e.Dialect() == dialect.PostgreSQL
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
			case "IMPORT":
				openBag = e.Dialect() == dialect.PostgreSQL
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

func ddlObjType(e *env.Env, tokens []FmtToken) string {

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		switch ctVal {

		case "ACCESS", "EVENT", "FOREIGN", "LARGE", "MATERIALIZED",
			"OPERATOR", "TRANSFORM":

			if idx < idxMax {
				tVal := ctVal + " " + tokens[idx+1].AsUpper()

				switch tVal {

				case "ACCESS METHOD", "EVENT TRIGGER", "FOREIGN TABLE",
					"LARGE OBJECT", "MATERIALIZED VIEW", "OPERATOR CLASS",
					"OPERATOR FAMILY", "TRANSFORM FOR":

					return tVal

				case "FOREIGN DATA":
					return "FOREIGN DATA WRAPPER"
				}
			}

			if ctVal == "OPERATOR" {
				return ctVal
			}

		case "AGGREGATE", "CAST", "COLLATION", "COLUMN", "CONSTRAINT",
			"CONVERSION", "DATABASE", "DOMAIN", "EXTENSION", "FUNCTION",
			"INDEX", "LANGUAGE", "PACKAGE", "POLICY", "PROCEDURE",
			"PUBLICATION", "ROLE", "ROUTINE", "RULE", "SCHEMA", "SEQUENCE",
			"SERVER", "STATISTICS", "SUBSCRIPTION", "TABLE", "TABLESPACE",
			"TRIGGER", "TYPE", "VIEW":

			return ctVal
		}

	}
	return ""
}

func formatDDLKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
	// nada
	default:
		return tokens
	}

	var ret []FmtToken

	objType := ddlObjType(e, tokens)
	parensDepth := 0

	for _, cTok := range tokens {

		ctVal := cTok.AsUpper()

		switch objType {
		case "TABLE":
			switch ctVal {
			case "ADD", "ALTER", "ALWAYS", "AND", "AS", "ATTACH", "BY",
				"CASCADE", "CHECK", "COLUMN", "COMMENT", "COMMIT",
				"CONCURRENTLY", "CONSTRAINT", "CREATE", "DATA", "DEFAULT",
				"DELETE", "DETACH", "DROP", "EXCLUDE", "EXECUTE", "FOR",
				"FOREIGN", "FROM", "GENERATED", "GLOBAL", "HASH", "IDENTITY",
				"IN", "INDEX", "IS", "KEY", "LIST", "NOT", "NULL", "OF", "ON",
				"OPTIONS", "OWNER", "PARTITION", "PREPARE", "PRIMARY", "RANGE",
				"REFERENCES", "RENAME", "RESTRICT", "SELECT", "SET", "TABLE",
				"TABLESPACE", "TEMP", "TEMPORARY", "TO", "TYPE", "UNIQUE",
				"UPDATE", "USING", "VALIDATE", "VALUES", "WHERE", "WITH":

				cTok.SetUpper()
			}

		case "DATABASE":
			switch ctVal {
			case "ALTER", "CASCADE", "CREATE", "DATABASE", "DROP", "EXISTS",
				"FORCE", "IF", "ON", "OWNER", "RENAME", "SET", "TO", "WITH":
				cTok.SetUpper()
			case "ALLOW_CONNECTIONS", "BUILTIN_LOCALE",
				"COLLATION_VERSION", "ICU_LOCALE", "ICU_RULES",
				"IS_TEMPLATE", "LC_COLLATE", "LC_CTYPE", "LOCALE",
				"LOCALE_PROVIDER", "OID", "STRATEGY":
				switch e.Dialect() {
				case dialect.PostgreSQL:
					cTok.SetUpper()
				}
			}

		case "RULE":
			switch ctVal {
			case "ALSO", "INSTEAD", "NOTHING":
				switch e.Dialect() {
				case dialect.PostgreSQL:
					cTok.SetUpper()
				}
			}

		case "SERVER":
			switch ctVal {
			case "ADD", "SET", "DROP":
				switch e.Dialect() {
				case dialect.PostgreSQL:
					cTok.SetUpper()
				}
			}
		}

		switch ctVal {
		case "AND", "OR", "NOT", "NULL":
			cTok.SetUpper()
		case "IS", "DISTINCT":
			switch e.Dialect() {
			case dialect.PostgreSQL:
				cTok.SetUpper()
			}
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		default:
			switch parensDepth {
			case 0:
				if cTok.IsKeyword() && !cTok.IsDatatype() {
					cTok.SetUpper()
				}
			}
		}

		ret = append(ret, cTok)
	}

	return ret
}

func formatDDLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatDDLKeywords(e, b.tokens)

	idxMax := len(tokens) - 1

	parensDepth := 0

	ddlAction := ""
	objType := ddlObjType(e, tokens)
	isAlterOwner := false

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		if ddlAction == "" {
			switch ctVal {
			case "CREATE", "ALTER", "DROP", "IMPORT":
				ddlAction = ctVal
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := true
		ensureVSpace := false

		// TODO

		switch ctVal {
		case "OWNER":
			if ddlAction == "ALTER" {
				isAlterOwner = true
			}
		case "AS":

			switch objType {
			case "VIEW", "MATERIALIZED VIEW":
				ensureVSpace = true
			}

		case "(":
			if parensDepth == 0 {
				honorVSpace = false
			}
		case ")":
			if parensDepth < 2 {
				honorVSpace = false
			}
		}

		switch {
		case pTok.IsBag():
			bk := bagKey(pTok.typeOf, pTok.id)
			b, ok := bagMap[bk]
			if ok {
				switch {
				case b.HasTrailingComments():
					ensureVSpace = true
					//default:
					//	switch ctVal {
					//	case ")", ";":
					//			honorVSpace = true
				default:
					honorVSpace = true
					//	}
				}
			}
		case cTok.HasLeadingComments():
			ensureVSpace = true
		case pTok.HasTrailingComments():
			ensureVSpace = true
		}

		switch ctVal {
		case "CREATE", "ALTER", "DROP", "IMPORT":
			honorVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth

		switch {
		case cTok.vSpace == 0:
			// nada
		case idx == 0:
		// nada
		case ctVal == ")":
			indents = max(indents-1, 1)
		default:
			switch ddlAction {
			case "CREATE":
				switch objType {
				case "VIEW", "MATERIALIZED VIEW":
				// nada
				default:
					indents++
				}
			default:
				indents++
			}
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

		addTok := true
		if parensDepth > 0 {
			switch cTok.AsUpper() {
			case "IN", "INOUT", "OUT":
				switch e.Dialect() {
				case dialect.PostgreSQL:
					// not needed for setting ownership
					addTok = false
				}
			}
		}
		if addTok {
			tFormatted = append(tFormatted, cTok)
		}
	}

	if isAlterOwner {
		parensDepth = 0
		pTok = FmtToken{}
		for i := 1; i < len(tFormatted); i++ {

			switch tFormatted[i].value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			default:
				switch {
				case tFormatted[i].HasLeadingComments(), pTok.HasTrailingComments():
					// nada
				default:
					if tFormatted[i].vSpace > 0 {
						tFormatted[i].AdjustVSpace(false, false)
						tFormatted[i].AdjustHSpace(e, pTok)
					}
				}
			}
			pTok = tFormatted[i]
		}
		tFormatted = wrapOnCommasX(e, DDLBag, 1, tFormatted)
	}

	adjustCommentIndents(bagType, &tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, b.forObj, tFormatted)
}
