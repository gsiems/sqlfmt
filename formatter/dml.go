package formatter

import "github.com/gsiems/sqlfmt/env"

// tagDML ensures that DML commands (SELECT, INSERT, etc.) are properly tagged
func tagDML(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// TODO
	// The issue with tagging DML is that the terminator might not be a ";"
	// For sub-queries it could be a closing parens and for PL code the DML
    // could be part of a loop.
    //
    // Question is, do we need to check value of last non-comment token, or if
    // the last non-comment token is an extracted bag, then the last non-comment
    // token of the extracted bag, or can we just look for ";", "END ...", or
    // for the parens count to go negative?

	return m
}
