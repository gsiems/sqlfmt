package formatter

import (
	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

type dmlFmtStat struct {
	pd   int    // parens depth
	pAct string // primary action (DML type)
	cAct string // current action
	stk  map[int]string
}

func newFmtStat() *dmlFmtStat {
	var s dmlFmtStat
	if s.stk == nil {
		s.stk = make(map[int]string)
	}
	return &s
}

func (s *dmlFmtStat) parensDepth() int {
	return s.pd
}

func (s *dmlFmtStat) incParensDepth() {
	s.pd++
}
func (s *dmlFmtStat) decParensDepth() {
	s.deleteClause()
	s.pd--
}

func (s *dmlFmtStat) primaryAction() string {
	return s.pAct
}

func (s *dmlFmtStat) currentAction() string {
	return s.cAct
}

func (s *dmlFmtStat) updateClause(c string) {
	s.stk[s.pd] = c

	// update the current action within the DML
	switch c {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "UPSERT":
		s.cAct = c
	}

	// update the primary action within the DML
	switch c {
	case "SELECT", "INSERT", "UPDATE", "DELETE", "UPSERT",
		"MERGE", "REFRESH", "REINDEX", "TRUNCATE", "WITH":

		switch s.pAct {
		case "", "WITH":
			s.pAct = c
		}

	}
}

func (s *dmlFmtStat) currentClause() string {
	for idx := s.pd; idx >= 0; idx-- {
		if v, ok := s.stk[idx]; ok {
			return v
		}
	}
	return ""
}

func (s *dmlFmtStat) deleteClause() {
	if _, ok := s.stk[s.pd]; ok {
		delete(s.stk, s.pd)
	}
}

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
			case "DELETE", "INSERT", "MERGE", "SELECT", "TRUNCATE", "UPDATE",
				"UPSERT", "WITH":
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
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
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
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
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

func formatDMLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	cat := newFmtStat()
	idxMax := len(b.tokens) - 1
	onConflict := false

	var tFormatted []FmtToken
	var pTok FmtToken  // The previous token
	var pNcVal string  // The upper case value of the previous non-comment token
	var ppNcVal string // The upper case value of the previous to the previous non-comment token
	var pKwVal string  // The upper case value of the previous keyword token

	// ucKw: The list of keywords that can be set to upper-case
	var ucKw = []string{"ALL", "AND", "ANY", "ARRAY", "AS", "ASC", "BETWEEN",
		"BY", "CASCADE", "CASE", "COLLATE", "CONCURRENTLY", "CONFLICT",
		"CONSTRAINT", "CROSS", "CURRENT", "DATA", "DEFAULT", "DELETE", "DESC",
		"DISTINCT", "DO", "ELSE", "END", "EXCEPT", "EXISTS", "FETCH", "FIRST",
		"FOR", "FROM", "FULL", "GROUP", "HAVING", "IDENTITY", "IN", "INNER",
		"INSERT", "INTERSECT", "INTO", "IS", "JOIN", "LAST", "LATERAL", "LEFT",
		"LIKE", "LIMIT", "MATCHED", "MATERIALIZED", "MERGE", "MINUS",
		"NATURAL", "NEXT", "NFC", "NFD", "NFKC", "NFKD", "NO", "NORMALIZED",
		"NOT", "NOTHING", "NOWAIT", "NULL", "NULLS", "OF", "OFFSET", "ON",
		"ONLY", "OR", "ORDER", "OUTER", "OVER", "OVERRIDING", "PARTITION",
		"RECURSIVE", "REFRESH", "REINDEX", "RESTART", "RETURNING", "RIGHT",
		"ROW", "ROWS", "SELECT", "SET", "SHARE", "SOURCE", "SYSTEM", "TABLE",
		"TARGET", "TEMP", "TEMPORARY", "THEN", "TRUNCATE", "UNION", "UNLOGGED",
		"UPDATE", "UPSERT", "USING", "VALUE", "VALUES", "VIEW", "WHEN",
		"WHERE", "WINDOW", "WITH", "WITHIN"}

	//var ucPKw = []string{"RECURSIVE", "LOCAL", "CHECK", "OPTION", "CASCADED"}

	for idx := 0; idx <= idxMax; idx++ {

		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		if cTok.IsKeyword() {
			cTok.SetKeywordCase(e, ucKw)
		}

		switch e.Dialect() {
		case dialect.PostgreSQL:
			switch ctVal {
			case "RECURSIVE", "LOCAL", "CHECK", "OPTION", "CASCADED",
				"SOURCE", "TARGET":
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		}

		////////////////////////////////////////////////////////////////
		// Track the DML type and current clause
		switch cat.parensDepth() {
		case 0:

			switch ctVal {
			case "SELECT", "INSERT", "UPDATE", "UPSERT", "DELETE", "MERGE",
				"REFRESH", "REINDEX", "TRUNCATE", "WITH":

				cat.updateClause(ctVal)

			case "FROM", "HAVING", "INTERSECT", "JOIN", "MINUS",
				"ORDER", "RETURNING", "SET", "UNION", "VALUES", "WHERE":

				cat.updateClause(ctVal)

			case "CONFLICT":
				onConflict = true

				cat.updateClause(ctVal)

			case "GROUP":
				//switch pNcVal {
				//case "WITHIN":
				//default:
				cat.updateClause(ctVal)
				//}
			}
		default:
			switch ctVal {
			case "VALUES":
				cat.updateClause(ctVal)
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := false

		// get the next non-comment token...
		//var nNcTok FmtToken
		var nNcVal string

		if idx+1 < idxMax {
			for j := idx + 1; j <= idxMax; j++ {
				if !b.tokens[j].IsCodeComment() {
					//nNcTok = b.tokens[j]
					nNcVal = b.tokens[j].AsUpper()
					break
				}
			}
		}

		switch cat.parensDepth() {
		case 0:
			switch ctVal {
			case "CROSS", "DELETE", "EXCEPT", "FULL", "HAVING", "INNER",
				"INSERT", "INTERSECT", "LEFT", "LIMIT", "MERGE", "MINUS",
				"NATURAL", "OFFSET", "ORDER", "REFRESH", "REINDEX",
				"RETURNING", "RIGHT", "SELECT", "TRUNCATE", "UNION", "UPSERT",
				"USING":

				ensureVSpace = true

			case "WHERE":
				if !onConflict || pNcVal != ")" {
					ensureVSpace = true
				}

			case "UPDATE":
				switch pNcVal {
				//case "DO":
				case "FOR":
					// nada
				default:
					ensureVSpace = true
				}

			case "FOR":
				if nNcVal == "UPDATE" {
					ensureVSpace = true
				}

			case "ON":
				switch pNcVal {
				case "CONFLICT":
					// nada
				default:
					ensureVSpace = true
				}

			case "INTO":
				switch {
				case cat.currentClause() == "INSERT":
					// nada
				case cat.currentClause() == "RETURNING":
					// nada
				case pNcVal == "MERGE":
					// nada

				default:
					ensureVSpace = true
				}

			case "WHEN":
				if cat.primaryAction() == "MERGE" {
					ensureVSpace = true
				}

			case "GROUP":
				switch pNcVal {
				case "WITHIN":
					// nada
				default:
					ensureVSpace = true
				}

			case "WITH":
				switch cat.primaryAction() {
				case "WITH", "SELECT":
					ensureVSpace = true
				}

			case "SET":
				//switch cat.primaryAction() {
				//    case "UPDATE", "CONFLICT":
				ensureVSpace = true
				//}

			case "FROM":
				switch {
				case cat.primaryAction() == "DELETE":
					// nada
				case pNcVal == "DISTINCT":
					// nada
				default:
					ensureVSpace = true
				}

			case "JOIN":
				switch pNcVal {
				case "LEFT", "RIGHT", "FULL", "CROSS", "LATERAL", "NATURAL", "INNER", "OUTER":
					// nada
				default:
					ensureVSpace = true
				}

			case "OUTER":
				switch e.Dialect() {
				case dialect.MSSQL:

					if nNcVal == "APPLY" {
						ensureVSpace = true
					}
				}
			}

			switch pNcVal {
			case ",":
				if cTok.IsCodeComment() {
					honorVSpace = true
				} else {
					ensureVSpace = true
				}
			}

		case 1:
			//switch ctVal {
			//case ")":
			//	if cat.primaryAction() == "WITH" {
			//		ensureVSpace = true
			//	}
			//}

			if cat.primaryAction() == "INSERT" {
				if cat.currentClause() == "INSERT" {
					switch pNcVal {
					case ",", "(":
						ensureVSpace = true
					}
				}
			}
		}

		switch cat.currentClause() {
		case "VALUES":
			switch ctVal {
			case "VALUES":
				if pNcVal != "DEFAULT" {
					ensureVSpace = true
				}
			case "(":
				switch {
				case cat.currentClause() == "CONFLICT":
					// nada
					//case pNcVal == "VALUES":
					// nada

				default:
					//if cat.primaryAction() != "INSERT" {
					ensureVSpace = true
					//}
				}
			case ")":
				if nNcVal == "AS" {
					ensureVSpace = true
				}
			default:
				if pNcVal == "," && ppNcVal == ")" {
					ensureVSpace = true
				}
			}
		case "WHERE", "JOIN":
			switch ctVal {
			case "OR":
				ensureVSpace = true
			case "AND":
				if pKwVal != "BETWEEN" {
					ensureVSpace = true
				}
			}
		}

		switch {
		case cTok.IsCodeComment(), cTok.IsBag():
			//ensureVSpace = false
			honorVSpace = true
		case pTok.IsCodeComment(), pTok.IsBag():
			//ensureVSpace = false
			honorVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine, and apply, the indentation level

		// TODO

		////////////////////////////////////////////////////////////////
		// Adjust the parens depth
		switch ctVal {
		case "(":
			cat.incParensDepth()
		case ")":
			cat.decParensDepth()
		}

		// Set the various "previous token" values
		pTok = cTok
		if !cTok.IsCodeComment() {
			ppNcVal = pNcVal
			pNcVal = ctVal
		}
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}

		tFormatted = append(tFormatted, cTok)
	}

	// TODO: Wrap long lines

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}
