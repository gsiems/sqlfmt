package formatter

import (
	"github.com/gsiems/db-dialect/dialect"
	"github.com/gsiems/sqlfmt/env"
)

// tagPLx ensures that blocks of PL (functions, procedures for PostgreSQL and
// functions, procedures, and packaged for Oracle) are properly tagged
func tagPLx(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {
	switch e.Dialect() {
	case dialect.PostgreSQL:
		return tagPgPL(e, m, bagMap)
	case dialect.SQLite:
		return tagSQLiteTrigger(e, m, bagMap)
	case dialect.Oracle:
		return tagOraPL(m, bagMap)
	}
	return m
}

func formatPLxBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	switch e.Dialect() {
	case dialect.PostgreSQL:
		formatPgPL(e, bagMap, bagType, bagId, baseIndents, forceInitVSpace)
	case dialect.SQLite:
		formatSQLiteTrigger(e, bagMap, bagType, bagId, baseIndents, forceInitVSpace)
	case dialect.Oracle:
		formatOraPL(e, bagMap, bagType, bagId, baseIndents, forceInitVSpace)
	}
}
