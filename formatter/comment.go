package formatter

import (
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
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

	if len(b.tokens) == 0 {
		return
	}

	idxMax := len(b.tokens) - 1
	parensDepth := 0
	hasParens := false

	var tFormatted []FmtToken
	var pTok FmtToken // The previous token

	for idx := 0; idx <= idxMax; idx++ {

		cTok := b.tokens[idx]
		ctVal := cTok.AsUpper()

		////////////////////////////////////////////////////////////////
		// Update keyword capitalization as needed
		switch parensDepth {
		case 0:
			if cTok.IsKeyword() && !cTok.IsDatatype() {
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		default:
			switch ctVal {
			case "IN", "INOUT", "OUT":
				switch e.Dialect() {
				case dialect.PostgreSQL:
					// not needed for commenting
					continue
				}
			case "AS":
				cTok.SetKeywordCase(e, []string{ctVal})
			}
		}

		////////////////////////////////////////////////////////////////
		// Determine the preceding vertical spacing (if any)
		honorVSpace := idx == 0
		ensureVSpace := idx == 0

		switch {
		case cTok.HasLeadingComments(), pTok.HasTrailingComments():
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
		// Adjust the parens depth
		switch cTok.value {
		case "(":
			hasParens = true
			parensDepth++
		case ")":
			parensDepth--
		}

		// Set the various "previous token" values
		pTok = cTok

		tFormatted = append(tFormatted, cTok)
	}

	if hasParens {
		tFormatted = wrapOnCommasX(e, bagType, 1, tFormatted)
	}

	adjustCommentIndents(bagType, &tFormatted)

	// Replace the mapped tokens with the newly formatted tokens
	UpsertMappedBag(bagMap, b.typeOf, b.id, b.forObj, tFormatted)
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

func adjustCommentIndents(bagType int, tokens *[]FmtToken) {

	ledIndents := 0
	trlIndents := 0
	idxMax := len((*tokens))
	for idx := 0; idx < idxMax; idx++ {
		if (*tokens)[idx].vSpace > 0 {

			// Determine the amount of indentation for the comments. By default
			// the indentation should be controlled by the token that it is
			// attached to. This isn't perfect, and there is no guarantee that
			// the comment is attached to the most appropriate token but it's
			// currently the best guess...
			ledIndents = (*tokens)[idx].indents
			trlIndents = calcIndent(bagType, (*tokens)[idx])

			switch (*tokens)[idx].value {
			case ")":
				//if idx > 0 {
				//ledIndents = (*tokens)[idx-1].indents
				ledIndents++
				trlIndents--
				//}
			default:
				switch bagType {
				case CommentOnBag:
					trlIndents++
				case DMLBag:
					switch (*tokens)[idx].AsUpper() {
					case "FROM":
						ledIndents++
					}
				case PLxBody:
					switch (*tokens)[idx].AsUpper() {
					case "ELSE", "END IF", "END CASE", "END LOOP", "END":
						ledIndents++
					case "EXCEPTION", "BEGIN":
						ledIndents++
					}
				}
			}
		}

		// If the comment is a block comment then check the last line of
		// the comment to see if the indentation/leading whitespace is zero.
		// If it is then leave the indents at zero.
		if (*tokens)[idx].HasLeadingComments() {
			for j, ct := range (*tokens)[idx].ledComments {
				ti := ledIndents
				switch ct.typeOf {
				case parser.BlockComment:
					z := strings.Split(ct.value, "\n")
					if len(z) > 1 {
						if z[len(z)-1] == strings.TrimLeft(z[len(z)-1], " \t") {
							ti = 0
						}
					}
				}
				(*tokens)[idx].ledComments[j].AdjustIndents(ti)
			}
		}
		if (*tokens)[idx].HasTrailingComments() {
			for j, ct := range (*tokens)[idx].trlComments {
				ti := trlIndents
				switch ct.typeOf {
				case parser.BlockComment:
					z := strings.Split(ct.value, "\n")
					if len(z) > 1 {
						if z[len(z)-1] == strings.TrimLeft(z[len(z)-1], " \t") {
							ti = 0
						}
					}
				}
				(*tokens)[idx].trlComments[j].AdjustIndents(ti)
			}
		}
	}
}
