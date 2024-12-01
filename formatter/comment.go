package formatter

import (
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

// tagCommentOn ensures that "COMMENT ON ... IS ..." commands are properly tagged
func tagCommentOn(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	remainder := tagSimple(e, m, bagMap, "COMMENT")

	return remainder
}

func formatCommentOn(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.lines) == 0 {
		return
	}

	var newLines [][]FmtToken

	line := b.lines[0]

	idxMax := len(line) - 1
	parensDepth := 0

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token

	for idx := 0; idx <= idxMax; idx++ {

		cTok := line[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Update keyword capitalization as needed
		// Identifiers should have been properly cased in cleanupParsed
		switch parensDepth {
		case 0:
			if cTok.IsKeyword() && !cTok.IsDatatype() {
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		default:
			switch ctVal {
			case "AS":
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := idx == 0

		switch {
		case cTok.IsCodeComment(), pTok.IsCodeComment():
			honorVSpace = true
		}

		cTok.AdjustVSpace(ensureVSpace, honorVSpace)

		////////////////////////////////////////////////////////////////
		// Determine the indentation level
		indents := baseIndents + parensDepth

		if idx > 0 && parensDepth == 0 {
			indents++
		}

		if cTok.vSpace > 0 {
			cTok.AdjustIndents(indents)
		} else {
			cTok.AdjustHSpace(e, pTok)
		}

		////////////////////////////////////////////////////////////////
		switch {
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

		tFormatted = append(tFormatted, cTok)
	}

	if len(tFormatted) > 0 {
		newLines = append(newLines, tFormatted)
	}

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, b.forObj, newLines)
}

func formatCodeComment(e *env.Env, cTok FmtToken, baseIndents int) FmtToken {

	rt := FmtToken{
		id:         cTok.id,
		categoryOf: cTok.categoryOf,
		typeOf:     cTok.typeOf,
		vSpace:     cTok.vSpace,
		indents:    cTok.indents,
		hSpace:     cTok.hSpace,
		value:      cTok.value,
		vSpaceOrig: cTok.vSpaceOrig,
		hSpaceOrig: cTok.hSpaceOrig,
	}

	rt.HonorVSpace()
	rt.AdjustHSpace(e, FmtToken{})

	if rt.vSpace > 0 {
		rt.indents = baseIndents
	}

	switch rt.categoryOf {
	case parser.LineComment, parser.PoundLineComment:
		// nada
	case parser.BlockComment:

		/* TODO:
		 * For comments that have leading vertical space, determine the initial indent and compare it to the indents value.
		 * If the indent values differ then adjust the indentation of the comment.
		 * For multiple line comments this will require removing the "initial indent" from each line and prepending the new initial indent
		 */

		//		lines := strings.Split(cTok.Value(), "\n")
		//ihSpace := cTok.hSpace

		//switch rt.vSpace {
		//case 0:

		// use indents to set leading space for each line

		//default:

		// use ihSpace to trim, then set, leading space for each line

		//}

		//for idx, line := range lines {

		/*
		   4-spaces per indent:

		   space space space space => 1 indent
		   space space space tab => 1 indent
		   space space tab => 1 indent
		   space tab => 1 indent
		   tab => 1 indent

		*/

		//}

	}
	return rt
}
