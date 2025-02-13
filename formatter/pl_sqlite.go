package formatter

import "github.com/gsiems/sqlfmt/env"

/*
CREATE TRIGGER cust_addr_chng
INSTEAD OF UPDATE OF cust_addr ON customer_address
BEGIN
  <dml statement>;
  [<dml statement>;]
END;
*/

func tagSQLiteTrigger(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	var remainder []FmtToken
	var bagTokens []FmtToken
	isInBag := false
	bagId := 0
	pKwVal := ""

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		switch isInBag {
		case true:
			bagTokens = append(bagTokens, cTok)
			switch ctVal {
			case ";":
				if pKwVal == "END" {

					key := bagKey(PLxBag, bagId)
					bagMap[key] = TokenBag{
						id:     bagId,
						typeOf: PLxBag,
						tokens: bagTokens,
					}

					isInBag = false
					bagTokens = nil
					pKwVal = ""
				}
			}

		case false:

			// check for the beginning of the PL object
			switch ctVal {
			case "TRIGGER":
				// Open a new bag
				isInBag = true
				bagId = cTok.id
				bagTokens = append(bagTokens, cTok)

				// Add a token that has the pointer to the new bag
				remainder = append(remainder, FmtToken{
					id:          bagId,
					categoryOf:  PLxBag,
					typeOf:      PLxBag,
					vSpace:      cTok.vSpace,
					indents:     cTok.indents,
					hSpace:      cTok.hSpace,
					vSpaceOrig:  cTok.vSpaceOrig,
					hSpaceOrig:  cTok.hSpaceOrig,
					ledComments: cTok.ledComments,
					trlComments: cTok.trlComments,
				})

			default:
				// Not in any PL object
				remainder = append(remainder, cTok)
			}
		}

		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	// On the off chance that the bag wasn't closed properly (incomplete or
	// incorrect statement submitted?), ensure that no tokens are lost.
	if len(bagTokens) > 0 {
		key := bagKey(PLxBag, bagId)

		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: PLxBag,
			tokens: bagTokens,
		}
	}

	return remainder
}

func formatSQLiteTriggerKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

	switch e.KeywordCase() {
	case env.UpperCase:
	// nada
	default:
		return tokens
	}

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {
		switch tokens[idx].AsUpper() {
		case "AFTER", "BEFORE", "BEGIN", "DELETE", "EACH", "END", "EXISTS",
			"FOR", "IF", "INSERT", "INSTEAD", "INSTEAD OF", "NOT", "OF", "ON",
			"ROW", "TRIGGER", "UPDATE", "WHEN":

			tokens[idx].SetUpper()

		}
	}

	return tokens
}

func formatSQLiteTrigger(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatSQLiteTriggerKeywords(e, b.tokens)
	idxMax := len(tokens) - 1

	var tFormatted []FmtToken
	ptVal := ""

	for idx := 0; idx <= idxMax; idx++ {

		ensureVSpace := false

		cTok := tokens[idx]
		switch cTok.AsUpper() {
		case "BEFORE", "AFTER", "INSTEAD":
			ensureVSpace = true
		case "DELETE", "INSERT", "UPDATE":
			switch ptVal {
			case "BEFORE", "AFTER", "INSTEAD OF", "OF":
				// nada
			default:
				ensureVSpace = true
			}
		case "ON", "FOR", "BEGIN", "END":
			ensureVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, false)

		switch cTok.AsUpper() {
		case "BEFORE", "AFTER", "INSTEAD", "INSTEAD OF", "DELETE", "INSERT",
			"UPDATE", "FOR", "ON":
			if cTok.vSpace > 0 {
				cTok.AdjustIndents(1)
			} else {
				cTok.hSpace = " "
			}
		}

		// Set the various "previous token" values
		ptVal = cTok.AsUpper()

		tFormatted = append(tFormatted, cTok)
	}

	for _, cTok := range tFormatted {
		if cTok.IsBag() {
			formatBag(e, bagMap, cTok.typeOf, cTok.id, 1, true)
		}
	}

	adjustCommentIndents(bagType, &tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}
