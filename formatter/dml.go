package formatter

import (
	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

func getBId(bagIds map[int]int, parensDepth int) int {

	// Get the most current bag ID, if needed/available
	// Problem: There won't be a valid bag ID for all parensDepths and
	// increasing the parensDepth doesn't signify that a new bag is needed.
	// So what IS needed is to backtrack up from the parensDepth until a valid
	// bagId is found.
	// This requires that the bagId entries be cleared up as the parensDepth is
	// decreased or when a bag is closed.

	pd := parensDepth
	testId := 0
	for pd >= 0 && testId == 0 {
		if bi, ok := bagIds[pd]; ok {
			testId = bi
		}
		pd--
	}
	return testId
}

// tagDML ensures that DML commands (SELECT, INSERT, etc.) are properly tagged
func tagDML(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// One issue with tagging DML is that the terminator might not be a ";"
	// For sub-queries it could be a closing parens and for PL code the DML
	// could be part of a loop.

	// Another issue with tagging DML, compared to tagging DCL or "COMMENT ON..."
	// statements is that the DML can have sub-units (sub-queries or CTEs) and
	// those sub-units each need to get their own bag. Giving sub-queries their
	// own bag should make determining the appropriate amount on indentation
	// during formatting much easier as each DML bag can be initially formatted
	// without needing to worry about how deeply nested the bag may be.

	var remainder []FmtToken

	tokMap := make(map[int][]FmtToken) // map[bagID][]FmtToken
	bagIds := make(map[int]int)        // map[parensDepth]bagID

	isInBag := false
	bagId := 0

	parensDepth := 0
	pKwVal := ""      // The upper-case value of the previous keyword token
	pNcVal := ""      // The upper-case value of the previous non-comment token
	var pTok FmtToken // The previous token

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		addToMap := false
		canOpenBag := false
		canOpenChildBag := false
		closeBag := false
		openBag := false
		openChildBag := false

		switch isInBag {
		case true:
			// Consider whether the bag should be closed or if a child bag can
			// be opened.

			if bagId == 0 {
				bagId = getBId(bagIds, parensDepth)
			}

			switch ctVal {
			case ";":
				closeBag = true
				addToMap = true
			case "(":
				// NB we only care about the parens depth if we are in a bag
				// so that when the parens depth goes negative then we know
				// to exit the bag
				parensDepth++
				addToMap = true
			case ")":
				if _, ok := bagIds[parensDepth]; ok {
					delete(bagIds, parensDepth)
				}

				parensDepth--

				if parensDepth < 0 {
					closeBag = true
				} else {
					// Restore the bagId to the ID of the parent, if there is
					// one, and append the token to the parent bag
					bagId = getBId(bagIds, parensDepth)
					addToMap = true
				}
			case "LOOP":
				closeBag = true

			default:

				if pNcVal == "(" {
					// ASSERTION: all sub-queries are wrapped in parens
					canOpenChildBag = true
				}

				if pTok.IsBag() {
					closeBag = true
				} else {
					switch e.Dialect() {
					case dialect.PostgreSQL:
						if isPgBodyBoundary(ctVal) {
							closeBag = true
						} else {
							addToMap = true
						}
					default:
						addToMap = true
					}
				}
			}

		case false:
			// Consider the previous token data to determine if a bag could be opened
			switch pNcVal {
			case "", "(", ";":
				canOpenBag = true
			case "BEGIN", "LOOP", "THEN", "ELSE", "IN", "AS":
				canOpenBag = true
			case "ATOMIC":
				canOpenBag = e.Dialect() == dialect.PostgreSQL
			case "/":
				canOpenBag = e.Dialect() == dialect.Oracle
			default:
				if e.Dialect() == dialect.PostgreSQL && isPgBodyBoundary(pNcVal) {
					canOpenBag = true
				} else {
					canOpenBag = pTok.IsBag()
				}
			}
		}

		////////////////////////////////////////////////////////////////
		// If it is possible to maybe open a bag for either a DML query or for
		// a sub-query, determine if a bag should be opened
		switch {
		case canOpenBag:
			switch ctVal {
			case "DELETE", "INSERT", "MERGE", "SELECT", "UPDATE", "UPSERT",
				"WITH":
				openBag = true
			case "REFRESH":
				// materialized view
				openBag = true
			case "REPLACE":
				if e.Dialect() == dialect.SQLite {
					openBag = true
				}
			case "VALUES":
				// PostgreSQL doesn't appear to necessarily need the
				// "SELECT FROM" for VALUES statements.
				switch pKwVal {
				case "AS":
					openBag = true
				}
			}

		case canOpenChildBag:
			switch ctVal {
			case "DELETE", "INSERT", "MERGE", "SELECT", "UPDATE", "UPSERT",
				"WITH":
				openChildBag = true
			}
		}

		////////////////////////////////////////////////////////////////
		// Actually process the token
		switch {
		case openBag:
			// Open the initial new bag
			isInBag = true
			bagId = cTok.id

			// Add a token that has the pointer to the new bag...
			remainder = append(remainder, FmtToken{
				id:         bagId,
				categoryOf: DMLBag,
				typeOf:     DMLBag,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
			})

			// ...and start the new bag
			bagIds[0] = bagId
			tokMap[bagId] = []FmtToken{cTok}
			parensDepth = 0

		case openChildBag:

			// Add a pointer to the parent bag...
			tokMap[bagId] = append(tokMap[bagId], FmtToken{
				id:         cTok.id,
				categoryOf: DMLBag,
				typeOf:     DMLBag,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
			})

			// ...and start the child bag
			bagId = cTok.id
			bagIds[parensDepth] = bagId
			tokMap[bagId] = append(tokMap[bagId], cTok)

		case addToMap:
			tokMap[bagId] = append(tokMap[bagId], cTok)
		default:
			remainder = append(remainder, cTok)
		}

		if closeBag {
			isInBag = false

			// Reset the bag IDs in case there are more DML blocks to tag
			bagId = 0
			for id, _ := range bagIds {
				delete(bagIds, id)
			}
			parensDepth = 0
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		pTok = cTok
		if !cTok.IsCodeComment() {
			pNcVal = ctVal
		}
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	// If the token map is not empty (DML was found and tagged) then populate
	// the bagMap
	for bagId, bagTokens := range tokMap {

		key := bagKey(DMLBag, bagId)
		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: DMLBag,
			tokens: bagTokens,
		}
	}

	return remainder
}
