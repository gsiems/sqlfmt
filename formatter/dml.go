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
	case "SELECT", "INSERT", "UPDATE", "DELETE", "UPSERT", "MERGE":
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

				if pTok.AsUpper() == "(" {
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
			switch pTok.AsUpper() {
			case "", "(", ";":
				canOpenBag = true
			case "BEGIN", "LOOP", "THEN", "ELSE", "IN", "AS":
				canOpenBag = true
			case "ATOMIC", "QUERY":
				canOpenBag = e.Dialect() == dialect.PostgreSQL
			case "/":
				canOpenBag = e.Dialect() == dialect.Oracle
			default:
				if e.Dialect() == dialect.PostgreSQL && isPgBodyBoundary(pTok.value) {
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
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	/*
		////////////////////////////////////////////////////////////////////
		// Tag case structures
		var ids []int
	*/
	typMap := make(map[int]int) // map[bagID]BagType
	/*
		for id, _ := range tokMap {
			ids = append(ids, id)
			typMap[id] = DMLBag
		}

		for _, baseBagId := range ids {

			//bagTokens := tokMap[baseBagId]
			caseIds := make(map[int]int) // map[caseDepth]bagId
			caseDepth := 0
			caseBagId := 0
			var newBagTokens []FmtToken

			for _, cTok := range tokMap[baseBagId] {

				switch cTok.AsUpper() {
				case "CASE":

					nt := FmtToken{
						id:         cTok.id,
						categoryOf: DMLBag,
						typeOf:     DMLCaseBag,
						vSpace:     cTok.vSpace,
						indents:    cTok.indents,
						hSpace:     cTok.hSpace,
						vSpaceOrig: cTok.vSpaceOrig,
						hSpaceOrig: cTok.hSpaceOrig,
					}

					caseDepth++
					switch {
					case caseDepth == 1:
						newBagTokens = append(newBagTokens, nt)
					case caseDepth > 1:
						tokMap[caseBagId] = append(tokMap[caseBagId], nt)
					}

					caseBagId = cTok.id
					caseIds[caseDepth] = caseBagId
					typMap[caseBagId] = DMLCaseBag
					tokMap[caseBagId] = append(tokMap[caseBagId], cTok)

				case "END":

					tokMap[caseBagId] = append(tokMap[caseBagId], cTok)
					if _, ok := bagIds[caseDepth]; ok {
						delete(bagIds, caseDepth)
					}
					caseDepth--
					caseBagId = caseIds[caseDepth]
					if caseDepth <= 0 {
						// Reset the bag IDs in case there are more CASE statements to tag
						caseBagId = 0
						caseDepth = 0
						caseIds = nil
						caseIds = make(map[int]int)
					}

				default:
					if caseDepth > 0 {
						tokMap[caseBagId] = append(tokMap[caseBagId], cTok)
					} else {
						newBagTokens = append(newBagTokens, cTok)
					}
				}
			}
			tokMap[baseBagId] = newBagTokens
		}
	*/
	////////////////////////////////////////////////////////////////////
	// If the token map is not empty (DML was found and tagged) then populate
	// the bagMap
	for bagId, bagTokens := range tokMap {

		typ := DMLBag

		if t, ok := typMap[bagId]; ok {
			typ = t
		}

		key := bagKey(typ, bagId)

		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: typ,
			tokens: bagTokens,
		}
	}

	return remainder
}

func formatDMLKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
	// nada
	default:
		return tokens
	}

	var ret []FmtToken

	for _, cTok := range tokens {

		ctVal := cTok.AsUpper()

		switch ctVal {
		case "ALL", "AND", "ANY", "AS", "ASC", "BETWEEN", "BY", "CASCADE",
			"CASE", "COLLATE", "CONCURRENTLY", "CONFLICT", "CONSTRAINT",
			"CROSS", "CURRENT", "DATA", "DELETE", "DESC", "DISTINCT", "DO",
			"ELSE", "END", "EXCEPT", "EXISTS", "FETCH", "FIRST", "FOR",
			"FOR UPDATE", "FROM", "FULL", "GROUP", "HAVING", "IDENTITY", "IN",
			"INNER", "INSERT", "INTERSECT", "INTO", "IS", "JOIN", "LAST",
			"LATERAL", "LEFT", "LIKE", "LIMIT", "MATCHED", "MATERIALIZED",
			"MERGE", "MINUS", "NATURAL", "NEXT", "NFC", "NFD", "NFKC", "NFKD",
			"NO", "NORMALIZED", "NOT", "NOTHING", "NOWAIT", "NULL", "NULLS",
			"OF", "OFFSET", "ON", "ON CONFLICT", "ONLY", "OR", "ORDER",
			"ORDER BY", "OUTER", "OVER", "OVERRIDING", "PARTITION",
			"PARTITION BY", "RECURSIVE", "REFRESH", "REINDEX", "RESTART",
			"RETURNING", "RIGHT", "ROW", "ROWS", "SELECT", "SET", "SHARE",
			"SOURCE", "SYSTEM", "TABLE", "TARGET", "TEMP", "TEMPORARY", "THEN",
			"TRUNCATE", "UNION", "UNLOGGED", "UPDATE", "UPSERT", "USING",
			"VALUE", "VALUES", "VIEW", "WHEN", "WHERE", "WINDOW", "WITH",
			"WITHIN", "GROUP BY":

			if cTok.IsKeyword() {
				cTok.SetUpper()
			}
		}

		switch e.Dialect() {
		case dialect.PostgreSQL:
			switch ctVal {
			case "RECURSIVE", "LOCAL", "CHECK", "OPTION", "CASCADED",
				"SOURCE", "TARGET":
				cTok.SetUpper()
			}
		case dialect.SQLite:
			switch ctVal {
			case "REPLACE":
				cTok.SetUpper()
			}
		}

		ret = append(ret, cTok)
	}

	return ret
}

func formatDMLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatDMLKeywords(e, b.tokens)

	cat := newFmtStat()
	idxMax := len(tokens) - 1
	indents := baseIndents
	onConflict := false

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token
	var pKwVal string // The upper case value of the previous keyword token

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Track the DML type and current clause
		switch cat.parensDepth() {
		case 0:

			switch ctVal {
			case "DELETE", "FROM", "GROUP BY", "HAVING", "INSERT", "INTERSECT",
				"JOIN", "MERGE", "MINUS", "ORDER BY", "REFRESH", "REINDEX",
				"RETURNING", "SELECT", "SET", "TRUNCATE", "UNION", "UPDATE",
				"UPSERT", "VALUES", "WHERE", "WITH":

				cat.updateClause(ctVal)

			case "ON CONFLICT":
				onConflict = true
				cat.updateClause(ctVal)

			}
		default:
			switch ctVal {
			case "VALUES":
				cat.updateClause(ctVal)
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := false
		ensureVSpace := false

		if idx == 0 {
			honorVSpace = true
			ensureVSpace = forceInitVSpace
		}

		nNcVal := ""
		if idx < idxMax {
			nNcVal = tokens[idx+1].AsUpper()
		}

		switch cat.parensDepth() {
		case 0:
			switch ctVal {
			case "":
			// nada
			case cat.primaryAction():
				ensureVSpace = true
			case cat.currentAction():
				ensureVSpace = true
			case "CROSS", "DELETE", "EXCEPT", "FULL", "HAVING", "INNER",
				"INSERT", "INTERSECT", "LEFT", "LIMIT", "MERGE", "MINUS",
				"NATURAL", "OFFSET", "ORDER", "REFRESH", "REINDEX",
				"RETURNING", "RIGHT", "SELECT", "TRUNCATE", "UNION", "UPSERT",
				"USING", "GROUP BY", "ORDER BY":

				ensureVSpace = true

			case "REPLACE":
				switch e.Dialect() {
				case dialect.SQLite:
					ensureVSpace = true
				}

			case "WHERE":
				if !onConflict || pTok.value != ")" {
					ensureVSpace = true
				}

			case "UPDATE":
				ensureVSpace = true

			case "FOR UPDATE", "ON CONFLICT":
				ensureVSpace = true

			case "ON":
				switch nNcVal {
				case "CONSTRAINT":
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
				case pTok.AsUpper() == "MERGE":
					// nada
				default:
					switch e.Dialect() {
					case dialect.SQLite:
						switch ctVal {
						case "REPLACE":
						// nada
						default:
							ensureVSpace = true
						}
					default:
						ensureVSpace = true
					}
				}
			case "WHEN":
				if cat.primaryAction() == "MERGE" {
					ensureVSpace = true
				}

			case "GROUP":
				switch pTok.AsUpper() {
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
				case pTok.AsUpper() == "DISTINCT":
					// nada
				default:
					ensureVSpace = true
				}

			case "JOIN":
				switch pTok.AsUpper() {
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
		}

		switch cat.currentClause() {
		case "WITH":
			if ctVal == ")" {
				switch pKwVal {
				case "RECURSIVE":
					// nada
				default:
					ensureVSpace = true
				}
			}
			switch pTok.value {
			case ",":
				if cat.parensDepth() == 0 {
					ensureVSpace = true
				}
			}

		case "VALUES":
			switch ctVal {
			case cat.currentClause():
				ensureVSpace = true
			case ")":
				if nNcVal == "AS" {
					ensureVSpace = true
				}
			}

		case "SELECT", "SET", "GROUP BY", "ORDER BY":
			switch pTok.value {
			case ",":
				if cat.parensDepth() == 0 {
					ensureVSpace = true
				}
			}

		case "WHERE", "JOIN":
			if isLogical(pKwVal, cTok) {
				ensureVSpace = true
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
					//	default:
					//		honorVSpace = true
					//	}
				}
			}
		case cTok.HasLeadingComments(), pTok.HasTrailingComments():
			honorVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine, and apply, the indentation level
		localIndents := 0

		if cTok.vSpace > 0 {
			indents = baseIndents + cat.parensDepth()

			switch cat.primaryAction() {

			case "WITH":
				switch ctVal {
				case "WITH":
					// nada
				case ")":
					localIndents = -1
				}
			case "SELECT":
				switch ctVal {
				case "SELECT", "ALL", "EXCEPT", "INTERSECT", "MINUS", "UNION":
					// nada
				case "INTO", "FROM", "WHERE", "GROUP", "HAVING", "WINDOW", "ORDER",
					"OFFSET", "LIMIT", "FETCH", "FOR", "WITH", "FOR UPDATE":
					localIndents = 1
				case "GROUP BY", "ORDER BY", "PARTITION BY":
					localIndents = 1
				case "CROSS", "FULL", "INNER", "JOIN", "LATERAL", "LEFT",
					"NATURAL", "OUTER", "RIGHT":
					localIndents = 1
				case "VALUES":
					localIndents = 1
				case ")":
					localIndents = 1
				default:
					switch {
					case cat.currentAction() == "SELECT":
						localIndents = 2
					case cTok.IsBag():
						localIndents = 1
					default:
						localIndents = 2
					}
				}
			case "INSERT":
				switch cat.currentAction() {
				case "SELECT":
					switch ctVal {
					case "SELECT", "ALL", "EXCEPT", "INTERSECT", "MINUS", "UNION":
						localIndents = 1
					case "FROM", "WHERE", "GROUP", "HAVING", "WINDOW", "ORDER",
						"OFFSET", "LIMIT", "FETCH", "FOR", "WITH", "RETURNING":
						localIndents = 2
					case "GROUP BY", "ORDER BY":
						localIndents = 2
					case "CROSS", "FULL", "INNER", "JOIN", "LATERAL", "LEFT",
						"NATURAL", "OUTER", "RIGHT":
						localIndents = 2
					case "VALUES", "ON CONFLICT":
						localIndents = 2
					case "ON":
						localIndents = 3
					case ")":
						localIndents = 2
					default:
						switch {
						case cTok.IsBag():
							localIndents = 2
						default:
							localIndents = 3
						}
					}

				default:
					switch {
					case onConflict:
						switch ctVal {
						case "ON CONFLICT":
							localIndents = 1
						case "UPDATE", "DELETE":
							localIndents = 2
						case "SET", "WHERE":
							localIndents = 3
						default:
							localIndents = 4
						}
					default:
						switch ctVal {
						case "INSERT":
							// nada
						case "(":
							localIndents = 2
						default:
							localIndents = 1
						}
					}
				}
			case "MERGE":
				switch ctVal {
				case "MERGE":
					localIndents = 0
				case "USING", "WHEN":
					localIndents = 1
				case "ON":
					localIndents = 2
				case "INSERT", "UPDATE", "DELETE":
					localIndents = 2
				case "SET", "VALUES", "RETURNING":
					localIndents = 3
				default:
					localIndents = 4
				}
			case "DELETE":
				switch ctVal {
				case "DELETE":
					localIndents = 0
				case "USING", "WHERE", "RETURNING":
					localIndents = 1
				case "CROSS", "FULL", "INNER", "JOIN", "LATERAL", "LEFT",
					"NATURAL", "OUTER", "RIGHT":
					localIndents = 1
				default:
					localIndents = 2
				}
			case "UPDATE":
				switch ctVal {
				case "UPDATE":
					localIndents = 0
				case "SET", "FROM", "WHERE", "RETURNING":
					localIndents = 1
				case "CROSS", "FULL", "INNER", "JOIN", "LEFT", "NATURAL",
					"OUTER", "RIGHT":
					localIndents = 1
				default:
					localIndents = 2
				}
			case "REFRESH", "REINDEX", "TRUNCATE":
				switch ctVal {
				case "":
					localIndents = 1
				case cat.primaryAction():
					localIndents = 0
				default:
					localIndents = 1
				}

			}

			indents += localIndents

		} // end cTok.vSpace > 0

		////////////////////////////////////////////////////////////////
		// Update the type and amount of white-space before the token
		if cTok.vSpace > 0 {
			cTok.AdjustIndents(indents)
		} else {
			cTok.AdjustHSpace(e, pTok)
		}

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
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}

		tFormatted = append(tFormatted, cTok)
	}

	wt := wrapLines(e, DMLBag, tFormatted)

	parensDepth := 0
	indents = 0
	for _, cTok := range wt {

		switch cTok.value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		default:
			if cTok.vSpace > 0 {
				parensDepth = 0
				indents = cTok.indents
				switch cTok.AsUpper() {
				case "SELECT":
					indents += 2
				case "WHERE":
					indents++
				}
			}
			if cTok.IsBag() {
				formatBag(e, bagMap, cTok.typeOf, cTok.id, indents+parensDepth, true)
			}
		}
	}

	adjustCommentIndents(bagType, &wt)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", wt)
}
