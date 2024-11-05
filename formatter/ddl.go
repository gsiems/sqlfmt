package formatter

import "github.com/gsiems/sqlfmt/env"

// tagDDL ensures that DDL commands (CREATE, ALTER, DROP) are properly tagged
func tagDDL(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// TODO

    // NB that it will be necessary to also scan the PLxBody bags (for
    // PostgreSQL) in order to tag any DDL embedded in the plpgsql.

	return m
}
