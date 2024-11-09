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

func formatPLxBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int, baseIndents int) {

	switch e.Dialect() {
	case dialect.PostgreSQL:
		formatPgPL(e, bagMap, bagType, bagId, baseIndents)
	//case dialect.Oracle:
	//	formatOraPL(e, bagMap, bagType, bagId, baseIndents)
	}
}
