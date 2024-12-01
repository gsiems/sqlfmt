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

func formatDCLBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	idxMax := len(b.tokens) - 1
	nextIndents := 0
	parensDepth := 0

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token
	var pNcVal string // The upper case value of the previous non-comment token

	// ucKw: The list of keywords that can be set to upper-case
	var ucKw = []string{"ADMIN", "ALL", "ALTER", "BY", "CASCADE", "CONNECT",
		"CREATE", "DATA", "DATABASE", "DELETE", "DOMAIN", "EXECUTE", "FOR",
		"FOREIGN", "FROM", "FUNCTION", "FUNCTIONS", "GRANT", "GRANTED", "IN",
		"INHERIT", "INSERT", "LANGUAGE", "LARGE", "MAINTAIN", "OBJECT", "ON",
		"OPTION", "PARAMETER", "PRIVILEGES", "PROCEDURE", "PROCEDURES",
		"REFERENCES", "RESTRICT", "REVOKE", "ROUTINE", "ROUTINES", "SCHEMA",
		"SELECT", "SEQUENCE", "SEQUENCES", "SERVER", "SET", "SYSTEM", "TABLE",
		"TABLES", "TABLESPACE", "TEMP", "TEMPORARY", "TO", "TRIGGER",
		"TRUNCATE", "TYPE", "UPDATE", "USAGE", "WITH", "WRAPPER"}

	for idx := 0; idx <= idxMax; idx++ {

		// current token
		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		if cTok.IsKeyword() {
			cTok.SetKeywordCase(e, ucKw)
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := idx == 0

		switch {
		case cTok.IsCodeComment(), cTok.IsLabel():
			honorVSpace = true
		case pTok.IsCodeComment(), pTok.IsLabel(), pTok.IsBag():
			honorVSpace = true
		}

		// ensure v-space for non-comment tokens that follow a semi-colon
		if !cTok.IsCodeComment() && pNcVal == ";" {
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
			formatBag(e, bagMap, cTok.typeOf, cTok.id, indents)
		case cTok.IsCodeComment():
			cTok = formatCodeComment(e, cTok, indents)
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

		tFormatted = append(tFormatted, cTok)
	}

	// TODO: Wrap long lines

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, "", tFormatted)
}
