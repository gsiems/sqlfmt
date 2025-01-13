package formatter

import (
	"strings"

	"github.com/gsiems/sqlfmt/env"
)

// isPgBodyBoundary determines if the supplied string is a boundary marker for
// a PostgreSQL function or procedure
func isPgBodyBoundary(s string) bool {
	if !strings.HasPrefix(s, "$") {
		return false
	}
	if !strings.HasSuffix(s, "$") {
		return false
	}
	if len(s) < 2 {
		return false
	}
	return true
}

// tagPgPL ensures that the DDL for creating PostgreSQL functions and
// procedures are properly tagged
func tagPgPL(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// One issue with tagging PostgreSQL functions and procedures is the
	// relationship of the body to everything around it in that many/most of
	// the directives outside the body (such as LANGUAGE, SECURITY, etc.) can
	// appear in pretty much any order and can appear either before or after
	// the body of the function/procedure.
	// One "fun" wrinkle is that the language might not be known until after
	// the tokens have been mostly bagged such that it may be necessary to
	// revisit the bags to update the bag types.

	// There are also differences in how the beginning and end of plpgsql,
	// old-style sql, atomic sql functions/procedures are specified and also DO
	// blocks as may be found in psql scripts. Oh, and triggers... wouldn't do
	// to forget about triggers-- strictly speaking, they aren't really PL in
	// PostgreSQL (unlike Oracle) (though they DO have an EXECUTE ... bit) and
	// it's the trigger function that the contains the real PL.

	/*
	   CREATE [ OR REPLACE ] { PROCEDURE | FUNCTION }
	       name ( [ [ argmode ] [ argname ] argtype [ { DEFAULT | = } default_expr ] [, ...] ] )
	      ...

	   CREATE [ OR REPLACE ]        } DDL
	   { FUNCTION | PROCEDURE } ... } bag
	   LANGUAGE SQL                 } bag
	   IMMUTABLE                    } bag
	   RETURNS NULL ON NULL INPUT   } bag
	   RETURN a + b;                } bag

	   CREATE [ OR REPLACE ]        } DDL
	   { FUNCTION | PROCEDURE } ... } bag
	   LANGUAGE SQL                 } bag
	   AS $$                        } bag
	       <DML>                    } body (pointer to DML bag)
	   $$ <possible other stuff> ;  } bag

	   CREATE [ OR REPLACE ]        } DDL
	   { FUNCTION | PROCEDURE } ... } bag
	   LANGUAGE SQL                 } bag
	   BEGIN ATOMIC                 } body
	       <DML>                    } body (pointer to DML bag)
	   END ;                        } body

	   CREATE [ OR REPLACE ]        } DDL
	   { FUNCTION | PROCEDURE } ... } bag
	   LANGUAGE plpgsql             } bag
	   AS $$                        } bag
	   [ DECLARE ... ]              } body
	   BEGIN                        } body
	      <Procedural stuff>        } body
	   END ;                        } body
	   $$ <possible other stuff> ;  } bag

	   CREATE [ OR REPLACE ]        } DDL
	   TRIGGER ...                  } bag
	   ";"                          } bag

	*/

	// If in declaration and not in body and see ";" then done
	// If found "body boundary" and see ";" after seeing second matching body boundary then done
	// If see ";" after seeing "end" after seeing "begin atomic" then done

	var remainder []FmtToken

	tokMap := make(map[int][]FmtToken) // map[bagID][]FmtToken
	typMap := make(map[int]int)        // map[bagID]BagType

	canCloseBody := false
	canOpenBody := false
	isAtomic := false
	isInBag := false
	isInBody := false
	isDo := false

	bagId := 0
	bodyBagId := 0
	bodyBoundary := ""
	plLang := ""

	pKwVal := "" // The upper-case value of the previous keyword token

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		addToBag := false
		addToBody := false
		closeBag := false
		closeBody := false
		openBag := false
		openBody := false

		switch isInBag {
		case true:
			// Consider whether the bag should be closed or if the body bag can
			// be opened.

			if plLang == "" && pKwVal == "LANGUAGE" {
				plLang = cTok.value
			}

			switch isInBody {
			case true:
				switch {
				case ctVal == "ATOMIC":
					isAtomic = true
				case isAtomic && canCloseBody:
					if ctVal == ";" {
						canCloseBody = false
						closeBag = true
						addToBody = true
					}
				case isAtomic:
					// ASSERTION: DML has already been bagged so we don't
					// need to worry about "CASE .... END"
					canCloseBody = ctVal == "END"
					addToBody = true
				default:
					// ASSERT bodyBoundary != ""
					if cTok.value == bodyBoundary {
						closeBody = true
					}
					addToBody = !closeBody
				}

				// Set the body type as early as possible just in case the
				// input doesn't result in a properly closed bag thereby
				// disconnecting the body from the rest of the PLx definition.
				if bodyBagId > 0 && typMap[bodyBagId] == DNFBag {
					switch strings.ToLower(plLang) {
					case "sql", "plpgsql":
						typMap[bodyBagId] = PLxBody
					default:
						if plLang == "" && isDo {
							typMap[bodyBagId] = PLxBody
						}
					}
				}

			case false:
				// in bag, not in body
				switch ctVal {
				case "DECLARE", "BEGIN":
					openBody = true
					addToBody = true
				case ";":
					closeBag = true
					addToBag = true

				default:
					if canOpenBody {
						// ASSERT: the previous token was the body boundary
						openBody = true
						addToBody = true
					} else if isPgBodyBoundary(ctVal) {
						bodyBoundary = cTok.value
						canOpenBody = true
					}
				}
			}
		case false:
			// not in bag
			switch ctVal {
			case "FUNCTION", "PROCEDURE", "TRIGGER":
				switch pKwVal {
				case "CREATE", "REPLACE":
					openBag = true
				}
			case "DO":
				openBag = true
			}
		}

		switch {
		case openBag:
			// Open the new bag
			isInBag = true
			bagId = cTok.id

			// Add a token that has the pointer to the new bag...
			remainder = append(remainder, FmtToken{
				id:         bagId,
				categoryOf: PLxBag,
				typeOf:     PLxBag,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
			})

			// ...and start the new bag
			tokMap[bagId] = []FmtToken{cTok}
			typMap[bagId] = PLxBag

			if ctVal == "DO" {
				isDo = true
			}

		case openBody:
			// Add a pointer to the parent bag...
			tokMap[bagId] = append(tokMap[bagId], FmtToken{
				id:         cTok.id,
				categoryOf: PLxBag,
				typeOf:     PLxBody,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
			})

			// ...and start the body bag
			isInBody = true
			bodyBagId = cTok.id
			tokMap[bodyBagId] = append(tokMap[bodyBagId], cTok)
			typMap[bodyBagId] = DNFBag

			canOpenBody = false

		case closeBody:
			if addToBody {
				tokMap[bodyBagId] = append(tokMap[bodyBagId], cTok)
			} else {
				tokMap[bagId] = append(tokMap[bagId], cTok)
			}
			isInBody = false

		case closeBag:
			if addToBody {
				tokMap[bodyBagId] = append(tokMap[bodyBagId], cTok)
			} else if addToBag {
				tokMap[bagId] = append(tokMap[bagId], cTok)
			} else {
				remainder = append(remainder, cTok)
			}

			// Ensure that the bag types for the body bag and body bag pointer
			// tokens are properly set to match the appropriate type for the
			// PL language
			if bodyBagId > 0 {
				bodyType := typMap[bodyBagId]

				if bodyType == DNFBag {
					switch strings.ToLower(plLang) {
					case "sql", "plpgsql":
						bodyType = PLxBody
					default:
						if plLang == "" && isDo {
							bodyType = PLxBody
						}
					}
					typMap[bodyBagId] = bodyType
				}

				// Check the bagType of the pointer token for the body.
				// If the bagType differs from the bagType that corresponds to
				// the language then update the bagType of the pointer
				toks := tokMap[bagId]
				idxMax := len(toks) - 1
				updateMap := false
				for idx := 0; idx <= idxMax; idx++ {
					t := toks[idx]
					if t.id == bodyBagId && t.typeOf != bodyType {
						t.typeOf = bodyType
						toks[idx] = t
						updateMap = true
					}
				}
				if updateMap {
					tokMap[bagId] = toks
				}
			}

			bagId = 0
			bodyBagId = 0
			isDo = false
			isInBag = false
			isInBody = false
			plLang = ""

		case isInBody:
			tokMap[bodyBagId] = append(tokMap[bodyBagId], cTok)
		case isInBag:
			tokMap[bagId] = append(tokMap[bagId], cTok)
		default:
			remainder = append(remainder, cTok)
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	// If the token map is not empty (PL was found and tagged) then populate
	// the bagMap
	for bagId, bagTokens := range tokMap {

		typ := DNFBag
		if t, ok := typMap[bagId]; ok {
			typ = t
		}

		if typ == PLxBody {
			bagTokens = tagDDL(e, bagTokens, bagMap)
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

func formatPgPL(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {
	switch bagType {
	case PLxBag:
		formatPgPLNonBody(e, bagMap, bagType, bagId, baseIndents, forceInitVSpace)
	case PLxBody:
		formatPgPLBody(e, bagMap, bagType, bagId, baseIndents, forceInitVSpace)
	}
}

func pgParamLabel(objType, paramLabel string, pTok, cTok, nTok FmtToken) string {

	ctVal := cTok.AsUpper()
	ptVal := pTok.AsUpper()
	ntVal := nTok.AsUpper()

	switch objType {
	case "TRIGGER":
		if ptVal == objType {
			return "NAME"
		}

		switch ctVal {
		case "TRIGGER":
			return "TYPE"
		case "BEFORE", "AFTER", "INSTEAD":
			return "EVENT"
		case "ON":
			return "TABLE"
		case "NOT":
			if ntVal == "DEFERRABLE" {
				return "DEFERRABLE"
			}
		case "DEFERRABLE", "INITIALLY":
			return "DEFERRABLE"
		case "REFERENCING", "FOR", "WHEN", "EXECUTE":
			return ctVal
		}

	default:
		switch ptVal {
		case objType:
			return "NAME"
		default:
			switch ctVal {
			case "FUNCTION", "PROCEDURE", "DO":
				return "TYPE"
			case "(":
				if paramLabel == "NAME" {
					return "SIGNATURE"
				}
			case "RETURNS":
				if paramLabel == "SIGNATURE" {
					return "RETURNS"
				} else if ntVal == "NULL" {
					return "CALLING MODE"
				}
			case "LANGUAGE", "TRANSFORM", "PARALLEL", "COST", "ROWS", "SET", "AS":
				return ctVal
			case "IMMUTABLE", "STABLE", "VOLATILE":
				return "VOLATILE"
			case "NOT", "LEAKPROOF":
				return "LEAKPROOF"
			case "CALLED", "STRICT":
				return "CALLING MODE"
			case "EXTERNAL", "SECURITY":
				return "SECURITY"
			case ";":
				return "FINAL"
			default:
				switch {
				case isPgBodyBoundary(ctVal):
					return "BODY"
				case cTok.typeOf == PLxBody:
					return "BODY"
				}
			}
		}
	}

	return paramLabel
}

func formatPgPLBodyKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
	// nada
	default:
		return tokens
	}

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		switch tokens[idx].AsUpper() {
		case "AND", "ANY", "AS", "ATOMIC", "BEGIN", "BETWEEN", "BREAK", "CASE",
			"CLOSE", "CONCURRENTLY", "CONTINUE", "DECLARE", "DISTINCT", "ELSE",
			"ELSEIF", "ELSIF", "END", "END CASE", "END IF", "END LOOP",
			"EXECUTE", "EXCEPTION", "EXISTS", "EXIT", "FETCH", "FOR",
			"FOREACH", "FOUND", "FROM", "GET", "IF", "IN", "INTO", "IS",
			"LIKE", "LOOP", "MATERIALIZED", "NEXT", "NOT", "NULL", "OPEN",
			"OR", "QUERY", "RAISE", "REFRESH", "RETURN", "SETOF", "THEN",
			"VIEW", "WHEN", "WHILE":

			//"SQLERRM", "SQLSTATE", "STACKED", "DIAGNOSTICS",

			tokens[idx].SetUpper()

		case "NOTICE", "WARNING":
			if idx > 0 {
				switch tokens[idx-1].AsUpper() {
				case "RAISE":
					tokens[idx].SetUpper()
				}
			}
		}

	}

	return tokens
}

func formatPgPLNonBodyKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
	// nada
	default:
		return tokens
	}

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		switch tokens[idx].AsUpper() {
		case "AFTER", "AND", "AS", "BEFORE", "CALLED", "CONSTRAINT", "COST",
			"CREATE", "CURRENT", "DEFAULT", "DEFERRABLE", "DEFERRED",
			"DEFINER", "DELETE", "DISTINCT", "DO", "EACH", "EXECUTE",
			"EXTERNAL", "FOR", "FROM", "FUNCTION", "IMMEDIATE", "IMMUTABLE",
			"INITIALLY", "INPUT", "INSERT", "INSTEAD", "INVOKER", "IS",
			"LANGUAGE", "LEAKPROOF", "NEW", "NOT", "NULL", "OF", "OLD", "ON",
			"OR", "PARALLEL", "PROCEDURE", "REFERENCING", "REPLACE",
			"RESTRICTED", "RETURNS", "ROW", "ROWS", "SAFE", "SECURITY", "SET",
			"SETOF", "STABLE", "STATEMENT", "STRICT", "SUPPORT", "TABLE", "TO",
			"TRANSFORM", "TRIGGER", "TRUNCATE", "TYPE", "UNSAFE", "UPDATE",
			"VOLATILE", "WHEN", "WINDOW":

			tokens[idx].SetUpper()

		case "SQL", "C":
			// check for language
			if idx > 0 {
				switch tokens[idx-1].AsUpper() {
				case "LANGUAGE":
					tokens[idx].SetUpper()
				}
			}
		}
	}

	return tokens
}

func formatPgPLBody(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatPgPLBodyKeywords(e, b.tokens)

	idxMax := len(tokens) - 1

	isAtomic := false
	parensDepth := 0
	var bbStack plStack

	var tFormatted []FmtToken
	var pKwVal string // The upper case value of the previous keyword token

	declareCnt := 0
	for idx := 0; idx <= idxMax; idx++ {
		if tokens[idx].AsUpper() == "DECLARE" {
			declareCnt++
		}
	}

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Update the block/branch stack
		switch ctVal {
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

		// Determine if a new-line should be applied before specific tokens
		switch ctVal {
		case "BEGIN", "BREAK", "CALL", "CASE", "CLOSE", "CONTINUE", "DECLARE",
			"ELSE", "ELSEIF", "ELSIF", "END", "END CASE", "END IF", "END LOOP",
			"EXCEPTION", "EXIT", "FOREACH", "IF", "INTO", "OPEN", "RETURN",
			"WHILE":

			ensureVSpace = true

		case "FOR":
			switch pKwVal {
			case "OPEN":
				// nada
			default:
				ensureVSpace = true
			}
		case "EXECUTE":
			switch ptVal {
			case "FOR", "IN":
				// nada
			default:
				ensureVSpace = true
			}

		case "WHEN":
			if bbStack.Last() == "CASE" {
				ensureVSpace = true
			}
			if bbStack.LastBlock() == "EXCEPTION" {
				ensureVSpace = true
			}

		case ")":
			switch {
			//case pTok.IsLabel(), pTok.HasTrailingComments():
			case pTok.IsLabel():
				ensureVSpace = true
			case pTok.IsBag():
				if ntVal != "LOOP" {
					honorVSpace = true
				}
			}
		}

		// Determine if a new-line should be applied after specific tokens.
		switch {
		case pTok.IsLabel(), pTok.HasTrailingComments():
			// Not yet. Checked here so it doesn't need checking for each ntVal case
		case cTok.IsLabel(), cTok.HasLeadingComments():
			// Not yet. Checked here so it doesn't need checking for each ntVal case
		default:

			switch ptVal {
			case ";", "ELSE":
				ensureVSpace = true

			case "LOOP":
				if ctVal != ";" {
					ensureVSpace = true
				}

			case "DECLARE":
				// it would be nice to always have a new-line after DECLARE, but...
				// since some code uses DECLARE before each individual variable
				// (ESRI comes to mind) it can't be assumed that there will be a
				// new-line after
				switch declareCnt {
				case 1:
					ensureVSpace = true
				default:
					honorVSpace = true
				}

			case "BEGIN":
				switch ctVal {
				case "ATOMIC":
					isAtomic = true
				default:
					ensureVSpace = true
				}

			case "RAISE":
				// whatever is being raised, we want no v-space
				ensureVSpace = false
				honorVSpace = false

			case "THEN":
				switch {
				case bbStack.Last() == "IF":
					ensureVSpace = true
				case bbStack.Last() == "CASE":
					ensureVSpace = true
				case bbStack.LastBlock() == "EXCEPTION":
					ensureVSpace = true
				}
			}
		}

		// For code comments, labels, and other (DML) bags, defer to the
		// original white-space.
		switch {
		//case cTok.HasLeadingComments(), cTok.IsLabel(), cTok.IsBag():
		//	honorVSpace = true
		//case pTok.HasTrailingComments(), pTok.IsLabel():
		//	honorVSpace = true

		case cTok.IsLabel(), cTok.IsBag():
			honorVSpace = true
		case pTok.IsLabel():
			honorVSpace = true
		case pTok.IsBag():

			switch ctVal {
			case ")":
				if ntVal != "LOOP" {
					honorVSpace = true
				}
			default:
				honorVSpace = true
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
		case pTok.HasTrailingComments():
			ensureVSpace = true
		case pTok.IsLabel():
			ensureVSpace = true
		case cTok.HasLeadingComments():
			ensureVSpace = true
		case cTok.IsLabel(), cTok.IsBag():
			honorVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth + bbStack.Indents()

		if cTok.vSpace > 0 {
			switch ctVal {
			case "DECLARE", "BEGIN":
				indents--
			case "EXCEPTION":
				indents -= 2
			case "IF", "LOOP":
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

			if bbStack.LastBlock() == "EXCEPTION" {
				if pKwVal == "DIAGNOSTICS" && ptVal == "," {
					indents++
				}
			}

			if isAtomic && indents > 0 {
				// Even though ATOMIC SQL functions have a BEGIN and END,
				// ISTM that the indentation should match the non-atomic
				// SQL functions... if only so that a function can be flipped
				// between the two without changing the indentation of the
				// whole thing.
				indents--
			}
		}

		////////////////////////////////////////////////////////////////
		// Update the type and amount of white-space before the token
		if cTok.vSpace > 0 {
			cTok.AdjustIndents(indents)
		} else {
			cTok.AdjustHSpace(e, pTok)
		}

		////////////////////////////////////////////////////////////////
		switch {
		case cTok.IsBag():
			if ptVal == "IN" {
				indents++
			}
			formatBag(e, bagMap, cTok.typeOf, cTok.id, indents, ensureVSpace)
			//case cTok.IsComment():
			//	cTok = formatComment(e, cTok, indents)
		}

		////////////////////////////////////////////////////////////////
		// Adjust the parens depth
		switch cTok.value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		}

		// set the line wrapping break points
		switch {
		case cTok.vSpace == 0:
			// nada
		case cTok.IsKeyword():
			cTok.fbp = true
		case ptVal == ";":
			cTok.fbp = true
		}

		// Set the previous keyword value
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}

		tFormatted = append(tFormatted, cTok)
	}

	wt := wrapLines(e, PLxBody, tFormatted)

	adjustCommentIndents(bagType, &wt)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", wt)
}

func formatPgPLNonBody(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	// TODO: consider adding a check for, and emitting a warning for SECURITY
	// DEFINER functions/procedures that do not set a search path or that have
	// an insecure search path

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatPgPLNonBodyKeywords(e, b.tokens)
	idxMax := len(tokens) - 1
	objType := ""

	var tFormatted []FmtToken

	// procedure/function labels
	var psLabels = []string{"TYPE", "NAME", "SIGNATURE", "RETURNS", "LANGUAGE",
		"TRANSFORM", "WINDOW", "VOLATILE", "LEAKPROOF", "CALLING MODE",
		"SECURITY", "PARALLEL", "COST", "ROWS", "SUPPORT", "SET", "AS",
		"BODY", "FINAL"}

	// trigger labels
	var tsLabels = []string{"TYPE", "NAME", "EVENT", "TABLE", "FROM",
		"DEFERRABLE", "REFERENCING", "FOR", "WHEN", "EXECUTE", "FINAL"}

	var params = make(map[string][]FmtToken)
	paramLabel := ""

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		switch cTok.AsUpper() {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "DO":
			if objType == "" {
				objType = cTok.AsUpper()
			}
		}

		////////////////////////////////////////////////////////////////
		// Re-order the parameters of the function/procedure declaration to
		// match that found in the PostgreSQL documentation.
		var pTok FmtToken
		var nTok FmtToken
		if idx > 0 {
			pTok = tokens[idx-1]
		}
		if idx < idxMax {
			nTok = tokens[idx+1]
		}

		paramLabel = pgParamLabel(objType, paramLabel, pTok, cTok, nTok)
		params[paramLabel] = append(params[paramLabel], cTok)
	}

	var sLabels []string
	switch objType {
	case "TRIGGER":
		sLabels = tsLabels
	default:
		sLabels = psLabels
	}

	// TODO: If there is a signature then format that
	if true {
		parensDepth := 0
		var pTok FmtToken

		for _, sn := range sLabels {
			if toks, ok := params[sn]; ok {

				for idx, cTok := range toks {

					honorVSpace := idx == 0
					ensureVSpace := false

					ctVal := cTok.AsUpper()

					switch ctVal {
					case "(":
						parensDepth++
					case ")":
						parensDepth--
					}

					switch sn {
					case "SIGNATURE", "RETURNS":
						if parensDepth == 1 {
							switch pTok.value {
							case "(", ",":
								ensureVSpace = true
							}
						}
					}

					switch sn {
					case "TYPE", "NAME", "SIGNATURE", "FINAL":
						//nada
					case "SET", "AS":
						if ctVal == sn {
							ensureVSpace = true
						}
					case "BODY":
						if cTok.IsPLBag() || pTok.IsPLBag() {
							ensureVSpace = true
						}
					default:
						if idx == 0 {
							ensureVSpace = true
						}
					}

					switch {
					case cTok.HasLeadingComments(), pTok.HasTrailingComments():
						honorVSpace = true
					}

					cTok.AdjustVSpace(ensureVSpace, honorVSpace)

					if cTok.vSpace > 0 {
						if objType == "TRIGGER" {
							cTok.AdjustIndents(baseIndents + parensDepth + 1)
						} else {
							cTok.AdjustIndents(baseIndents + parensDepth)
						}
					} else {
						cTok.AdjustHSpace(e, pTok)
					}

					if cTok.IsBag() {
						formatBag(e, bagMap, cTok.typeOf, cTok.id, cTok.indents, ensureVSpace)
					}

					// Set the previous token
					pTok = cTok

					tFormatted = append(tFormatted, cTok)
				}
			}
		}
	}

	// Cleanup extraneous vertical spacing
	if true {
		var pTok FmtToken
		for idx, cTok := range tFormatted {

			if isPgBodyBoundary(cTok.value) {
				if pTok.AsUpper() == "AS" {
					tFormatted[idx].vSpace = 0
					tFormatted[idx].hSpace = " "
				}
			} else if cTok.vSpace > 1 {
				tFormatted[idx].vSpace = 1
			}

			// TODO: The following is a hack. Can't see why but sorting the
			// non-body when there are parameters (in the input) after the
			// body can cause vertical space to be added prior to the closing
			// semi-colon. This removes extra vertical space but it would be
			// better to understand why this is happening.
			if cTok.value == ";" {
				if !pTok.HasTrailingComments() {
					tFormatted[idx].vSpace = 0
					tFormatted[idx].hSpace = " "
				}
			}
			// Set the previous token
			pTok = cTok
		}
	}

	adjustCommentIndents(bagType, &tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}

/*

formatting...






CREATE [ OR REPLACE ] FUNCTION name                                                         | name          | nl after function
     ( [ [ argmode ] [ argname ] argtype [ { DEFAULT | = } default_expr ] [, ...] ] )       | signature     | after name
    [ RETURNS rettype  | RETURNS TABLE ( column_name column_type [, ...] ) ]                | returns       | after signature
  { LANGUAGE lang_name                                                                      | language      |
    | TRANSFORM { FOR TYPE type_name } [, ... ]                                             | transform     |
    | WINDOW                                                                                | window        |
    | { IMMUTABLE | STABLE | VOLATILE }                                                     | volatile      |
    | [ NOT ] LEAKPROOF                                                                     | leakproof     | check next token
    | { CALLED ON NULL INPUT | RETURNS NULL ON NULL INPUT | STRICT }                        | calling mode  |
    | { [ EXTERNAL ] SECURITY INVOKER | [ EXTERNAL ] SECURITY DEFINER }                     | security      |
    | PARALLEL { UNSAFE | RESTRICTED | SAFE }                                               | parallel      |
    | COST execution_cost                                                                   | cost          |
    | ROWS result_rows                                                                      | rows          |
    | SUPPORT support_function                                                              | support       |
    | SET configuration_parameter { TO value | = value | FROM CURRENT }                     | set config    | can be list of SETs
    | AS 'definition'                                                                       | as            |
    | AS 'obj_file', 'link_symbol'                                                          | as            |
    | sql_body                                                                              | body          |
  } ...

CREATE [ OR REPLACE ] PROCEDURE name                                                        | name          |
    ( [ [ argmode ] [ argname ] argtype [ { DEFAULT | = } default_expr ] [, ...] ] )        | signature     |
  { LANGUAGE lang_name                                                                      | language      |
    | TRANSFORM { FOR TYPE type_name } [, ... ]                                             | transform     |
    | [ EXTERNAL ] SECURITY INVOKER | [ EXTERNAL ] SECURITY DEFINER                         | security      |
    | SET configuration_parameter { TO value | = value | FROM CURRENT }                     | set config    |
    | AS 'definition'                                                                       | as            |
    | AS 'obj_file', 'link_symbol'                                                          | as            |
    | sql_body                                                                              | body          |
  } ...

CREATE [ OR REPLACE ] [ CONSTRAINT ] TRIGGER name                                           | name          |
    { BEFORE | AFTER | INSTEAD OF } { event [ OR ... ] }                                    | event         |
    ON table_name                                                                           | table         |
    [ FROM referenced_table_name ]                                                          | from          |
    [ NOT DEFERRABLE | [ DEFERRABLE ] [ INITIALLY IMMEDIATE | INITIALLY DEFERRED ] ]        | deferrable    | check next token
    [ REFERENCING { { OLD | NEW } TABLE [ AS ] transition_relation_name } [ ... ] ]         | referencing   |
    [ FOR [ EACH ] { ROW | STATEMENT } ]                                                    | for           |
    [ WHEN ( condition ) ]                                                                  | when          |
    EXECUTE { FUNCTION | PROCEDURE } function_name ( arguments )                            | execute       |

where event can be one of:

    INSERT
    UPDATE [ OF column_name [, ... ] ]
    DELETE
    TRUNCATE

*/
