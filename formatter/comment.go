package formatter

import "github.com/gsiems/sqlfmt/env"

// tagComment ensures that "COMMENT ON ... IS ..." commands are properly tagged
func tagComment(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	remainder := tagSimple(e, m, bagMap, "COMMENT")

	return remainder
}
