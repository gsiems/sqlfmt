package formatter

import "strings"

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
