package formatter

import "github.com/gsiems/sqlfmt/env"

// tagDCL ensures that permissions setting commands are properly tagged
func tagDCL(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	remainder := m
	kw := []string{"GRANT", "REVOKE", "REASSIGN"}

	for _, cmd := range kw {
		remainder = tagSimple(e, remainder, bagMap, cmd)
	}

	return remainder
}

func formatDCLKeywords(e *env.Env, tokens []FmtToken) []FmtToken {

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
		case "ADMIN", "ALL", "ALTER", "BY", "CASCADE", "CONNECT", "CREATE",
			"DATA", "DATABASE", "DELETE", "DOMAIN", "EXECUTE", "FOR",
			"FOREIGN", "FROM", "FUNCTION", "FUNCTIONS", "GRANT", "GRANTED",
			"IN", "INHERIT", "INSERT", "LANGUAGE", "LARGE", "MAINTAIN",
			"OBJECT", "ON", "OPTION", "PARAMETER", "PRIVILEGES", "PROCEDURE",
			"PROCEDURES", "REFERENCES", "RESTRICT", "REVOKE", "ROUTINE",
			"ROUTINES", "SCHEMA", "SELECT", "SEQUENCE", "SEQUENCES", "SERVER",
			"SET", "SYSTEM", "TABLE", "TABLES", "TABLESPACE", "TEMP",
			"TEMPORARY", "TO", "TRIGGER", "TRUNCATE", "TYPE", "UPDATE",
			"USAGE", "WITH", "WRAPPER":

			if cTok.IsKeyword() {
				cTok.SetUpper()
			}
		}

		ret = append(ret, cTok)
	}

	return ret
}

func formatDCLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.tokens) == 0 {
		return
	}

	tokens := formatDCLKeywords(e, b.tokens)

	idxMax := len(tokens) - 1
	parensDepth := 0

	var tFormatted []FmtToken

	nextIndents := 0

	var pTok FmtToken // The previous token

	for idx := 0; idx <= idxMax; idx++ {

		// current token
		cTok := tokens[idx]

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := idx == 0

		switch {
		case cTok.IsLabel():
			honorVSpace = true
		case pTok.IsLabel(), pTok.IsBag():
			honorVSpace = true
		case cTok.HasLeadingComments():
			ensureVSpace = true
		case pTok.HasTrailingComments():
			ensureVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth + nextIndents

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

		tFormatted = append(tFormatted, cTok)
	}

	wt := wrapOnCommas(e, DCLBag, 1, tFormatted)

adjustCommentIndents (bagType, &wt)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", wt)
}
