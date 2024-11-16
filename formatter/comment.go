package formatter

import (
	"github.com/gsiems/sqlfmt/env"
)

// tagCommentOn ensures that "COMMENT ON ... IS ..." commands are properly tagged
func tagCommentOn(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	remainder := tagSimple(e, m, bagMap, "COMMENT")

	return remainder
}

func formatCodeComment(e *env.Env, cTok FmtToken, baseIndents int) FmtToken {

	rt := FmtToken{
		id:            cTok.id,
		categoryOf:    cTok.categoryOf,
		typeOf:        cTok.typeOf,
		vSpace:        cTok.vSpace,
		indents:       cTok.indents,
		hSpace:        cTok.hSpace,
		value:         cTok.value,
		vSpaceOrig:    cTok.vSpaceOrig,
		hSpaceOrig:    cTok.hSpaceOrig,
	}

	rt.HonorVSpace()
	rt.AdjustHSpace(e, FmtToken{})

	if rt.vSpace > 0 {
		rt.indents = baseIndents
	}

	switch rt.categoryOf {
	case LineComment, PoundLineComment:
		// nada
	case BlockComment:

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
