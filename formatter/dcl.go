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
