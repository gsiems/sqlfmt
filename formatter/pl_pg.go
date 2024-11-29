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
func tagPgPL(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

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
				bodyType := DNFBag

				switch strings.ToLower(plLang) {
				case "sql", "plpgsql":
					bodyType = PLxBody
				default:
					if plLang == "" && isDo {
						bodyType = PLxBody
					}
				}

				typMap[bodyBagId] = bodyType

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

		key := bagKey(typ, bagId)
		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: typ,
			tokens: bagTokens,
		}
	}

	return remainder
}

func formatPgPL(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {
	switch bagType {
	case PLxBag:
		formatPgPLNonBody(e, bagMap, bagType, bagId, baseIndents)
	case PLxBody:
		formatPgPLBody(e, bagMap, bagType, bagId, baseIndents)
	}
}

func pgParamLabel(objType, paramLabel, pNcVal, nNcVal string, cTok FmtToken) string {

	ctVal := cTok.AsUpper()

	switch objType {
	case "TRIGGER":
		if pNcVal == objType {
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
			if nNcVal == "DEFERRABLE" {
				return "DEFERRABLE"
			}
		case "DEFERRABLE", "INITIALLY":
			return "DEFERRABLE"
		case "REFERENCING", "FOR", "WHEN", "EXECUTE":
			return ctVal
		}

	default:
		switch pNcVal {
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
				} else if nNcVal == "NULL" {
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

func formatPgPLBody(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	idxMax := len(b.tokens) - 1
	parensDepth := 0
	var bbStack plStack

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token
	var pNcVal string // The upper case value of the previous non-comment token
	var pKwVal string // The upper case value of the previous keyword token

	// ucKw: The list of keywords that can be set to upper-case
	var ucKw = []string{"AND", "ANY", "OR", "NOT", "IS", "NULL", "AS", "FOREACH",
		"BEGIN", "BREAK", "CASE", "CLOSE", "CONTINUE", "DECLARE",
		"DIAGNOSTICS", "DISTINCT", "ELSE", "ELSEIF", "ELSIF", "END", "EXECUTE",
		"EXCEPTION", "EXIT", "FOR", "FROM", "GET", "IF", "IN", "IS", "LIKE",
		"LOOP", "NEXT", "NOTICE", "OPEN", "QUERY", "RAISE", "RETURN", "SETOF",
		"STACKED", "THEN", "WHEN", "WHILE", "SQLSTATE", "SQLERRM", "REFRESH",
		"MATERIALIZED", "VIEW", "CONCURRENTLY", "WARNING", "ATOMIC"}

	for idx := 0; idx <= idxMax; idx++ {

		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		if cTok.IsKeyword() && !cTok.IsDatatype() {
			cTok.SetKeywordCase(e, ucKw)
		}

		switch ctVal {
		case "IS", "DISTINCT", "RAISE":
			cTok.SetKeywordCase(e, []string{ctVal})
		case "NOTICE", "WARNING", "EXCEPTION":
			if pNcVal == "RAISE" {
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		}

		////////////////////////////////////////////////////////////////
		// Update the block/branch stack
		switch ctVal {
		case "DECLARE", "BEGIN", "EXCEPTION":
			bbStack.Upsert(ctVal)
		case "IF", "LOOP", "CASE":
			// WHILE/FOR vs. LOOP???
			if pNcVal != "END" {
				bbStack.Push(ctVal)
			}
		case "END":
			_ = bbStack.Pop()
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := false

		// Determine if a new-line should be applied before specific tokens
		switch ctVal {
		case "BEGIN", "BREAK", "CALL", "CLOSE", "CONTINUE", "DECLARE",
			"ELSE", "ELSEIF", "ELSIF", "END", "EXCEPTION", "EXECUTE", "EXIT",
			"FOR", "OPEN", "RETURN", "WHILE":

			ensureVSpace = true

		case "IF", "CASE":
			if pNcVal != "END" {
				ensureVSpace = true
			}

			// save these for line wrapping
		//case "AND":
		//	if pNcVal != "BETWEEN" {
		//		ensureVSpace = true
		//	}
		//
		//case "OR":
		//	ensureVSpace = true

		case "LOOP":
			if pNcVal != "END" {
				ensureVSpace = true
			}

		case "WHEN":
			if bbStack.Last() == "CASE" {
				ensureVSpace = true
			}
			if bbStack.LastBlock() == "EXCEPTION" {
				ensureVSpace = true
			}
		}

		// Determine if a new-line should be applied after specific tokens.
		switch {
		case pTok.IsLabel(), pTok.IsCodeComment():
			// Not yet. Checked here so it doesn't need checking for each pNcVal case
		case cTok.IsLabel(), cTok.IsCodeComment():
			// Not yet. Checked here so it doesn't need checking for each pNcVal case
		default:

			switch pNcVal {
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
				honorVSpace = true

			case "BEGIN":
				if ctVal != "ATOMIC" {
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
		case cTok.IsCodeComment(), cTok.IsLabel(), cTok.IsBag():
			ensureVSpace = false
			honorVSpace = true
		case pTok.IsCodeComment(), pTok.IsLabel(), pTok.IsBag():
			ensureVSpace = false
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
				if pNcVal != "END" {
					indents--
				}
			case "CASE":
				if pNcVal != "END" {
					indents -= 2
				}
			case "WHEN":
				if bbStack.Last() == "CASE" {
					indents--
				}
				if bbStack.LastBlock() == "EXCEPTION" {
					indents--
				}
			case "ELSIF", "ELSEIF", "ELSE":
				indents--
			}

			if bbStack.LastBlock() == "EXCEPTION" {
				if pKwVal == "DIAGNOSTICS" && pNcVal == "," {
					indents++
				}
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
			if pNcVal == "IN" {
				indents++
			}
			formatBag(e, bagMap, cTok.typeOf, cTok.id, indents)
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

		// Set the various "previous token" values
		pTok = cTok
		if !cTok.IsCodeComment() {
			pNcVal = ctVal
		}
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}

		tFormatted = append(tFormatted, cTok)
	}

	//tFormatted = WrapLongLines(e, b.typeOf, tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}

func formatPgPLNonBody(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	// TODO: consider adding a check for, and emitting a warning for SECURITY
	// DEFINER functions/procedures that do not set a search path or that have
	// an insecure search path

	idxMax := len(b.tokens) - 1
	parensDepth := 0
	objType := ""

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token
	var pNcVal string // The upper case value of the previous non-comment token

	// ucKw: The list of keywords that can be set to upper-case
	var ucKw = []string{"AFTER", "AND", "AS", "BEFORE", "CALLED", "CONSTRAINT",
		"COST", "CREATE", "CURRENT", "DEFAULT", "DEFERRABLE", "DEFERRED",
		"DEFINER", "DELETE", "DO", "EACH", "EXECUTE", "EXTERNAL", "FOR",
		"FROM", "FUNCTION", "IMMEDIATE", "IMMUTABLE", "INITIALLY", "INPUT",
		"INSERT", "INSTEAD", "INVOKER", "LANGUAGE", "LEAKPROOF", "NEW", "NOT",
		"NULL", "OF", "OLD", "ON", "OR", "PARALLEL", "PROCEDURE",
		"REFERENCING", "REPLACE", "RESTRICTED", "RETURNS", "ROW", "ROWS",
		"SAFE", "SECURITY", "SET", "SETOF", "STABLE", "STATEMENT", "STRICT",
		"SUPPORT", "TABLE", "TO", "TRANSFORM", "TRIGGER", "TRUNCATE", "TYPE",
		"UNSAFE", "UPDATE", "VOLATILE", "WHEN", "WINDOW"}

	var psLabels = []string{"TYPE", "NAME", "SIGNATURE", "RETURNS", "LANGUAGE",
		"TRANSFORM", "WINDOW", "VOLATILE", "LEAKPROOF", "CALLING MODE",
		"SECURITY", "PARALLEL", "COST", "ROWS", "SUPPORT", "SET", "AS",
		"BODY", "FINAL"}

	var tsLabels = []string{"TYPE", "NAME", "EVENT", "TABLE", "FROM",
		"DEFERRABLE", "REFERENCING", "FOR", "WHEN", "EXECUTE", "FINAL"}

	var params = make(map[string][]FmtToken)
	paramLabel := ""

	for idx := 0; idx <= idxMax; idx++ {

		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		switch ctVal {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "DO":
			if objType == "" {
				objType = ctVal
			}
		}

		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		if cTok.IsKeyword() {
			cTok.SetKeywordCase(e, ucKw)
		}
		switch ctVal {
		case "DO", "SAFE", "UNSAFE", "IS", "DISTINCT":
			cTok.SetKeywordCase(e, []string{ctVal})
		}

		// check for language
		switch pNcVal {
		case "LANGUAGE":
			switch ctVal {
			case "SQL", "C":
				cTok.SetUpper()
			}
		}

		////////////////////////////////////////////////////////////////
		// Re-order the parameters of the function/procedure declaration to
		// match that found in the PostgreSQL documentation.

		// Determine the lines
		if cTok.IsCodeComment() {

			// If there is a comment then it is probably for the param
			// associated with the next non-comment token, so determine what
			// the label for the next non-comment token would be so the comment
			// can remain with the following param.

			// get the next non-comment token...
			nNcIdx := 0
			var nNcTok FmtToken

			if idx+1 < idxMax {
				for j := idx + 1; j <= idxMax; j++ {
					if !b.tokens[j].IsCodeComment() {
						nNcTok = b.tokens[j]
						nNcIdx = j
						break
					}
				}
			}

			// ...and the next non-comment value after that
			nNcVal := ""
			if nNcIdx < idxMax {
				for j := nNcIdx + 1; j <= idxMax; j++ {
					if !b.tokens[j].IsCodeComment() {
						nNcVal = b.tokens[j].AsUpper()
						break
					}
				}
			}

			paramLabel = pgParamLabel(objType, paramLabel, pNcVal, nNcVal, nNcTok)
		} else {

			// get the next non-comment value
			nNcVal := ""
			if idx < idxMax {
				for j := idx + 1; j <= idxMax; j++ {
					if !b.tokens[j].IsCodeComment() {
						nNcVal = b.tokens[j].AsUpper()
						break
					}
				}
			}

			paramLabel = pgParamLabel(objType, paramLabel, pNcVal, nNcVal, cTok)
		}

		params[paramLabel] = append(params[paramLabel], cTok)

		// Set the various "previous token" values
		pTok = cTok
		if !cTok.IsCodeComment() {
			pNcVal = ctVal
		}
	}

	var sLabels []string
	switch objType {
	case "TRIGGER":
		sLabels = tsLabels
	default:
		sLabels = psLabels
	}

	// TODO: If there is a signature then format that

	for _, sn := range sLabels {
		if toks, ok := params[sn]; ok {
			parensDepth = 0

			for idx, cTok := range toks {

				ctVal := cTok.AsUpper()

				honorVSpace := idx == 0
				ensureVSpace := false

				switch sn {
				case "TYPE", "NAME", "SIGNATURE", "SET", "AS", "FINAL":
					//nada
				case "BODY":
					ensureVSpace = cTok.IsPLBag() || pTok.IsPLBag()
				default:
					ensureVSpace = idx == 0
				}

				switch ctVal {
				case "(":
					parensDepth++
				case ")":
					parensDepth--
				}

				switch sn {
				case "SIGNATURE":
					if parensDepth == 1 {
						switch pNcVal {
						case "(", ",":
							ensureVSpace = true
						}
					}
				case "SET", "AS":
					if ctVal == sn {
						ensureVSpace = true
					}
				}

				switch {
				case cTok.IsCodeComment(), pTok.IsCodeComment():
					honorVSpace = true
				}

				cTok.AdjustVSpace(ensureVSpace, honorVSpace)

				if cTok.vSpace > 0 {

					indents := baseIndents + parensDepth

					if objType == "TRIGGER" {
						indents++
					}

					cTok.AdjustIndents(indents)
				} else {
					cTok.AdjustHSpace(e, pTok)
				}

				if cTok.IsBag() {
					formatBag(e, bagMap, cTok.typeOf, cTok.id, cTok.indents)
				}

				// Set the various "previous token" values
				pTok = cTok
				if !cTok.IsCodeComment() {
					pNcVal = ctVal
				}

				tFormatted = append(tFormatted, cTok)
			}
		}
	}

	// TODO: The following is a hack. Can't see why but sorting the non-body
	// when there are parameters (in the input) after the body can cause line
	// feeds to be added prior to the closing semi-colon. This hack deals with
	// that unwanted extra vertical space but it would be better to understand
	// why this is happening.
	pTok = FmtToken{}
	for idx, cTok := range tFormatted {
		if cTok.value == ";" {
			if !pTok.IsCodeComment() {
				tFormatted[idx].vSpace = 0
				tFormatted[idx].hSpace = " "
			}
		}
		pTok = cTok
	}

	//tFormatted = WrapLongLines(e, b.typeOf, tFormatted)

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
