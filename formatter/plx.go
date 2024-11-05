package formatter

import (
	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

// tagPLx ensures that blocks of PL (functions, procedures for PostgreSQL and
// functions, procedures, and packaged for Oracle) are properly tagged
func tagPLx(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {
	switch e.Dialect() {
	case dialect.PostgreSQL:
		return tagPgPL(m, bagMap)
	case dialect.Oracle:
		return tagOraPL(m, bagMap)
	}
	return m
}
