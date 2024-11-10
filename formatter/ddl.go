package formatter

import (
	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

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
	ddlAction := ""
	//forObj := ""

	//pKwVal := ""      // The upper-case value of the previous keyword token
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

			/*
			 * parens depth needs to be zero
			 * if the DDL is for a PL unit then close can happen after the PLxBag
			 * if the DDL is not for a PL unit then the closing ";"
			 */

			switch ctVal {
			case "CREATE", "ALTER", "DROP":
				closeBag = true
				openBag = true
				ddlAction = ctVal

			default:
				switch ddlAction {
				case "DROP", "ALTER":

					if ctVal == ";" {
						closeBag = true
						addToBag = true
					}
				case "CREATE":
					// views, materialized views and PL units would already be bagged
					switch {
					case cTok.IsDMLBag(), cTok.IsPLBag():
						closeBag = true
						addToBag = true
					case ctVal == ";":
						closeBag = true
						addToBag = true
					}
				}

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
					}

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
				ddlAction = ctVal
			}
		}

		////////////////////////////////////////////////////////////////
		// Actually process the token
		switch {
		//case isInBag && closeBag:
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
				//forObj: forObj,
				tokens: bagTokens,
			}

			//forObj = ""

			bagId = 0
			bagTokens = nil

		case isInBag:
			bagTokens = append(bagTokens, cTok)

		case openBag:

			// Open a new bag
			isInBag = true

			bagId = cTok.id
			bagTokens = nil
			bagTokens = []FmtToken{cTok}

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
		//if cTok.IsKeyword() {
		//	pKwVal = ctVal
		//}
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
