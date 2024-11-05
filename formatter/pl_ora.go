package formatter

// tagOraPL ensures that the DDL for creating Oracle functions, procedures,
// packages and triggers are properly tagged
func tagOraPL(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// TODO

	// One issue with tagging Oracle functions, procedures, and packages is
	// that they can have sub-functions and procedures (that is pretty much the
	// definition of a package). The same might be true for triggers.

	return m
}
