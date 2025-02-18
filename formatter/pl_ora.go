package formatter

import (
	"github.com/gsiems/sqlfmt/env"
)

type plObj struct {
	id         int
	objType    string
	hasIs      bool
	hasLang    bool
	beginDepth int
}

// tagOraPL ensures that the DDL for creating Oracle functions, procedures,
// packages and triggers are properly tagged
func tagOraPL(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// One issue with tagging Oracle functions, procedures, and packages is
	// that they can have sub-functions and procedures (which is pretty much the
	// definition of a package). The same might be true for triggers.

	var remainder []FmtToken
	tokMap := make(map[int][]FmtToken) // map[bagID][]FmtToken
	bagId := 0
	pKwVal := ""
	plCnt := 0
	objs := make(map[int]plObj)

	for _, cTok := range m {

		ctVal := cTok.AsUpper()
		openBag := false
		closeBag := false

		switch ctVal {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "PACKAGE", "PACKAGE BODY", "TYPE BODY":
			switch pKwVal {
			case "DROP", "ALTER":
				remainder = append(remainder, cTok)
			default:
				openBag = true
			}
		default:
			if plCnt > 0 {
				tokMap[bagId] = append(tokMap[bagId], cTok)

				switch ctVal {
				case "IS", "AS":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.hasIs = true
						objs[plCnt] = n
					}
				case "LANGUAGE":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.hasLang = true
						objs[plCnt] = n
					}
				case "DECLARE":
					if _, ok := objs[plCnt]; !ok {
						openBag = true
					}
				case "BEGIN":
					_, ok := objs[plCnt]
					if ok {
						n := objs[plCnt]
						n.beginDepth++
						objs[plCnt] = n
					} else {
						openBag = true
					}
				case "END":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.beginDepth--
						objs[plCnt] = n
					}
				case ";":
					if obj, ok := objs[plCnt]; ok {
						switch {
						case pKwVal == "END":
							if obj.beginDepth <= 0 {
								closeBag = true
							}
						case obj.hasLang:
							closeBag = true
						case !obj.hasIs:
							closeBag = true
						}
					}
				}
			} else {
				switch ctVal {
				case "DECLARE", "BEGIN":
					openBag = true
				default:
					remainder = append(remainder, cTok)
				}
			}
		}

		switch {
		case closeBag:
			if _, ok := objs[plCnt]; ok {
				delete(objs, plCnt)
			}
			plCnt--
			bagId = 0
			switch {
			case plCnt > 0:
				if obj, ok := objs[plCnt]; ok {
					bagId = obj.id
				}
			}

		case openBag:
			parentId := 0
			if plCnt > 0 {
				if obj, ok := objs[plCnt]; ok {
					parentId = obj.id
				}
			}

			hasIs := false
			bd := 0
			if ctVal == "BEGIN" {
				bd++
				hasIs = true // not really, but this is probably an anonymous pl block so...
			}

			bagId = cTok.id
			plCnt++
			objs[plCnt] = plObj{id: cTok.id, objType: ctVal, hasIs: hasIs, hasLang: false, beginDepth: bd}

			nt := FmtToken{
				id:          cTok.id,
				categoryOf:  PLxBody,
				typeOf:      PLxBody,
				vSpace:      cTok.vSpace,
				indents:     cTok.indents,
				hSpace:      cTok.hSpace,
				vSpaceOrig:  cTok.vSpaceOrig,
				hSpaceOrig:  cTok.hSpaceOrig,
				ledComments: cTok.ledComments,
				trlComments: cTok.trlComments,
			}

			tokMap[bagId] = []FmtToken{cTok}

			switch parentId {
			case 0:
				// token is added to the map, remainder gets new pointer token
				remainder = append(remainder, nt)
			default:
				// token is added to the child map, parent map gets new pointer token (to the child)
				tokMap[parentId] = append(tokMap[parentId], nt)
			}
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	////////////////////////////////////////////////////////////////////
	// If the token map is not empty (PL was found and tagged) then populate
	// the bagMap
	for id, tokens := range tokMap {
		key := bagKey(PLxBody, id)
		bagMap[key] = TokenBag{
			id:     id,
			typeOf: PLxBody,
			tokens: tokens,
		}
	}

	return remainder
}

func formatOraPLKeywords(e *env.Env, objType string, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
		// nada
	default:
		return tokens
	}

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {
		switch tokens[idx].AsUpper() {
		case "AND", "AS", "IS", "BEGIN", "BETWEEN", "BREAK", "BULK COLLECT",
			"CASE", "CLOSE", "CONTINUE", "CONSTANT", "DECLARE", "DEFAULT",
			"DISTINCT", "ELSE", "ELSEIF", "ELSIF", "END", "END CASE", "END IF",
			"END LOOP", "EXECUTE", "EXCEPTION", "EXISTS", "EXIT", "FETCH",
			"FOR", "FORALL", "FOREACH", "FOUND", "FROM", "GET", "IF", "IN",
			"INTO", "LIKE", "LOOP", "NEXT", "NOT", "NULL", "OF", "OPEN", "OR",
			"IMMEDIATE", "RAISE", "REFRESH", "RETURN", "THEN", "VIEW", "WHEN",
			"WHILE", "FUNCTION", "PROCEDURE", "OUT", "PACKAGE", "PACKAGE BODY",
			"PRAGMA", "RECORD", "TABLE", "TYPE BODY", "VALUES", "TYPE",
			"COMMIT", "ROLLBACK", "USING":

			tokens[idx].SetUpper()
		}

		if objType == "TRIGGER" {
			switch tokens[idx].AsUpper() {
			case "AFTER", "BEFORE", "DELETE", "EACH", "INSERT", "INSTEAD OF",
				"NEW", "OLD", "ON", "REFERENCING", "ROW", "TRIGGER", "UPDATE":

				tokens[idx].SetUpper()
			}
		}
	}

	return tokens
}

func formatOraPL(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	const (
		preSig = iota + 1
		inSig
		postSig
	)

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	objType := b.tokens[0].AsUpper()

	tokens := formatOraPLKeywords(e, objType, b.tokens)
	idxMax := len(tokens) - 1
	parensDepth := 0
	var bbStack plStack
	bCnt := 0
	sigStat := 0

	var tFormatted []FmtToken
	pKwVal := ""

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Update the block/branch stack
		switch ctVal {
		case "TRIGGER", "PACKAGE", "PACKAGE BODY", "TYPE BODY", "FUNCTION", "PROCEDURE":
			bbStack.Push(ctVal)
		case "DECLARE", "BEGIN", "EXCEPTION":
			bbStack.Upsert(ctVal)
		case "IF", "LOOP", "CASE":
			// WHILE/FOR vs. LOOP???
			bbStack.Push(ctVal)
		case "END", "END CASE", "END IF", "END LOOP":
			_ = bbStack.Pop()
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := false

		////
		var pTok FmtToken
		var nTok FmtToken
		if idx > 0 {
			pTok = tokens[idx-1]
		}
		if idx < idxMax {
			nTok = tokens[idx+1]
		}
		ptVal := pTok.AsUpper()
		ntVal := nTok.AsUpper()

		if ctVal == "BEGIN" {
			bCnt++
		}

		// Determine if a new-line should be applied before specific tokens

		// function/procedure signature
		switch ptVal {
		case ";":
			sigStat = postSig
		case "FUNCTION", "PROCEDURE":
			sigStat = preSig
			bCnt = 0
		default:
			switch sigStat {
			case preSig:
				switch ptVal {
				case "AS", "IS":
					sigStat = postSig
				case "(":
					sigStat = inSig
					ensureVSpace = true
				}
			case inSig:
				switch ptVal {
				case ",":
					ensureVSpace = true
				case ")":
					sigStat = postSig
				}
			}
		}

		switch objType {
		case "PACKAGE", "PACKAGE BODY":

			switch parensDepth {
			case 0:
				switch ctVal {
				case "END", "SHARING", "AUTHID":
					ensureVSpace = true
				case "DEFAULT": // COLLATION
					ensureVSpace = true
				case "ACCESSIBLE": // BY
					ensureVSpace = true
				}
			case 1:
				switch ptVal {
				case "(", ",":
					ensureVSpace = true
				}
			}

		case "TRIGGER":
			switch ctVal {
			case "REFERENCING":
				ensureVSpace = true
			case "BEFORE", "AFTER", "INSTEAD OF":
				switch pTok.AsUpper() {
				case "OR":
					// nada
				default:
					ensureVSpace = true
				}
			case "FOR":
				ensureVSpace = true
			}

		case "FUNCTION", "PROCEDURE", "TYPE BODY":
			switch ctVal {
			case "TYPE", "FUNCTION", "PROCEDURE":
				switch pKwVal {
				case "IS", "AS":
					ensureVSpace = true
				}
			case "ACCESSIBLE", "AGGREGATE", "AUTHID", "DETERMINISTIC",
				"EXTERNAL", "PARALLEL_ENABLE", "PIPELINED", "RESULT_CACHE",
				"SHARING", "SQL_MACRO":
				ensureVSpace = true
			case "LANGUAGE", "LIBRARY", "NAME", "PARAMETERS":
				ensureVSpace = true
			}
		}

		switch objType {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "TYPE BODY":

			switch ctVal {
			case "BEGIN", "BREAK", "CALL", "CLOSE", "CONTINUE", "DECLARE",
				"ELSEIF", "ELSIF", "END CASE", "END IF", "END LOOP", "EXIT",
				"FORALL", "FOREACH", "IF", "OPEN", "WHILE", "EXCEPTION",
				"USING":

				ensureVSpace = true

			case "INTO":
				switch ptVal {
				case "BULK COLLECT":
					// nada
				default:
					ensureVSpace = true
				}
			case "RETURN":
				honorVSpace = true
			case "AS":
				switch ptVal {
				case "NEW", "OLD":
					// nada
				default:
					switch ntVal {
					case "NEW", "OLD":
						// nada
					default:
						ensureVSpace = parensDepth == 0
					}
				}

			case "IS":
				switch ntVal {
				case "NOT", "NULL":
					// nada
				default:
					switch pKwVal {
					case "TYPE":
						// nada
					default:
						if parensDepth == 0 {
							ensureVSpace = true
						}
					}
				}

			case "CASE":
				switch ptVal {
				case ";", "LOOP":
					ensureVSpace = true
				}
			case "ELSE":
				if bbStack.Last() == "IF" {
					ensureVSpace = true
				}
			case "END":
				switch ntVal {
				case ",", ")":
					// nada
				default:
					ensureVSpace = true
				}
			case "FOR":
				switch {
				case pKwVal == "OPEN":
					// nada
				case ptVal == "(":
					// nada
				default:
					ensureVSpace = true
				}
			case "EXECUTE":
				switch ptVal {
				case "FOR", "IN", "(":
					// nada
				default:
					ensureVSpace = true
				}
			case "WHEN":
				if bbStack.LastBlock() == "EXCEPTION" {
					ensureVSpace = true
				}
			}
		}

		// Determine if a new-line should be applied after specific tokens.
		switch ptVal {
		case ";":
			ensureVSpace = true
		case "ELSE":
			if bbStack.Last() == "IF" {
				ensureVSpace = true
			}
		case "LOOP":
			if ctVal != ";" {
				ensureVSpace = true
			}
		case "DECLARE":
			honorVSpace = true
		case "BEGIN":
			ensureVSpace = true
		case "AS":
			switch ctVal {
			case "OLD", "NEW":
				// nada
			default:
				if parensDepth == 0 {
					ensureVSpace = true
				}
			}
		case "IS":
			switch ctVal {
			case "NOT", "NULL":
				// nada
			case "RECORD", "TABLE":
				// nada
			default:
				if parensDepth == 0 {
					ensureVSpace = true
				}
			}
		case "RAISE":
			// whatever is being raised, we want no v-space
			ensureVSpace = false
			honorVSpace = false

		case "THEN":
			switch {
			case bbStack.Last() == "IF":
				ensureVSpace = true
			case bbStack.LastBlock() == "EXCEPTION":
				ensureVSpace = true
			}
		}

		// For code comments, labels, and other bags
		switch {
		case pTok.HasTrailingComments(), pTok.IsLabel():
			ensureVSpace = true
		case cTok.HasLeadingComments():
			ensureVSpace = true
		case cTok.IsLabel(), cTok.IsBag():
			honorVSpace = true
		case pTok.IsBag():
			bk := bagKey(pTok.typeOf, pTok.id)
			b, ok := bagMap[bk]
			if ok {
				switch {
				case b.HasTrailingComments():
					ensureVSpace = true
				default:
					switch ctVal {
					case ")":
						if ntVal != "LOOP" {
							honorVSpace = true
						}
					default:
						honorVSpace = true
					}
				}
			}
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth + bbStack.Indents()

		if cTok.vSpace > 0 {

			switch bbStack.Last() {
			case objType:
				indents = baseIndents + 1

				switch ctVal {
				case objType:
					indents = baseIndents
				case "IS", "AS", "END", "RETURN":
					if parensDepth == 0 {
						indents = baseIndents
					}
				case "ACCESSIBLE", "AGGREGATE", "AUTHID", "DETERMINISTIC",
					"EXTERNAL", "PARALLEL_ENABLE", "PIPELINED", "RESULT_CACHE",
					"SHARING", "SQL_MACRO":
					if parensDepth == 0 {
						indents = baseIndents
					}
				default:
					if bbStack.Last() == "PACKAGE" {
						if parensDepth == 1 {
							indents += parensDepth
						}
					}
				}
			default:
				switch ctVal {
				case "DECLARE", "BEGIN":
					indents--
				case "EXCEPTION":
					indents -= 2
				case "IF", "LOOP", "AS", "IS":
					// WHILE/FOR vs. LOOP???
					indents--
				case "CASE":
					indents -= 2
				case "WHEN":
					if bbStack.Last() == "CASE" {
						indents--
					}
					if bbStack.LastBlock() == "EXCEPTION" {
						indents--
					}
				case "ELSIF", "ELSEIF", "ELSE":
					indents--
				case "INTO":
					indents++
				}
			}
		}

		////////////////////////////////////////////////////////////////
		// Adjust the parens depth
		switch cTok.value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		}

		////////////////////////////////////////////////////////////////
		// Update the type and amount of white-space before the token
		if cTok.vSpace > 0 {
			cTok.AdjustIndents(indents)
		} else {
			cTok.AdjustHSpace(e, pTok)
		}

		switch ptVal {
		case ";":
			if bbStack.Length() > 1 {
				switch bbStack.Last() {
				case "PROCEDURE", "FUNCTION":
					_ = bbStack.Pop()
				}
			}
		}

		if cTok.IsKeyword() {
			pKwVal = cTok.AsUpper()
		}

		tFormatted = append(tFormatted, cTok)
	}

	tFormatted = wrapLines(e, bagType, tFormatted)

	adjustCommentIndents(bagType, &tFormatted)

	parensDepth = 0
	indents := 0
	for _, cTok := range tFormatted {

		switch cTok.value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		default:
			if cTok.vSpace > 0 {
				parensDepth = 0
				indents = cTok.indents
			}
			if cTok.IsBag() {
				formatBag(e, bagMap, cTok.typeOf, cTok.id, indents+parensDepth, true)
			}
		}
	}

	adjustCommentIndents(bagType, &tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}
