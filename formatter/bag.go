package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

type TokenBag struct {
	id       int          // the ID for the bag
	typeOf   int          // the type of token bag
	forObj   string       // the name of the kind of object that the bag is for (not all bag types care)
	lines    [][]FmtToken // the lines of token lists that make up the bag
	warnings []string     // list of (non-fatal) warnings found
	errors   []string     // list of (fatal) errors found
}

func bagKey(bagType, bagId int) string {

	/* One might hope that the ID of a token would be sufficient, however, in
		the case of Pg functions/procedures where the language is sql AND the first
		token of the body is also the first token of a DML block then using just
		the ID will result in one mapped bag (for the PL body) overwriting the other
	    (for the DMLBag).
	*/

	/* Additionally, pad the bagId, bagType so that things sort consistently
	when doing development testing (the padding isn't needed otherwise) */

	return fmt.Sprintf("%08d.%03d", bagId, bagType)
}

// tagSimple is used for tagging "simple" commands that start with a consistent
// keyword, are terminated by a semi-colon, and (most importantly) contain no
// additional semi-colons
func tagSimple(e *env.Env, m []FmtToken, bagMap map[string]TokenBag, cmdKwd string) []FmtToken {

	// NB that bagMap is a pointer to the map, not a copy of the map

	var remainder []FmtToken
	var bagTokens []FmtToken

	isInBag := false
	bagId := 0
	bagType := 0
	forObj := ""

	pNcVal := ""
	var pTok FmtToken

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		closeBag := false
		canOpenBag := false
		openBag := false

		switch isInBag {
		case true:
			if ctVal == ";" {
				closeBag = true
			}
		case false:
			switch pNcVal {
			case "", ";":
				canOpenBag = true
			case "/":
				canOpenBag = e.Dialect() == dialect.Oracle
			default:
				canOpenBag = pTok.IsBag()
			}
		}

		if canOpenBag {
			switch ctVal {
			case cmdKwd:
				switch cmdKwd {
				case "GRANT", "REVOKE":
					openBag = true
				case "REASSIGN":
					openBag = e.Dialect() == dialect.PostgreSQL
				case "COMMENT":
					// So far as I know, of the currently intended to be supported
					// dialects, only PostgreSQL, Oracle, and the standard actually
					// support "COMMENT ON object IS ..." syntax.
					//
					// While MySQL and MariaDB do support table/column comments,
					// the approach used is very different.
					//
					// For future reference, the COMMENT ON syntax is apparently
					// also supported by Firebird, Db2, Redshift, and Snowflake.
					switch e.Dialect() {
					case dialect.PostgreSQL, dialect.Oracle, dialect.StandardSQL:
						openBag = true
					}
				}
			}
		}

		switch {
		case isInBag && closeBag:
			bagTokens = append(bagTokens, cTok)
			var lines [][]FmtToken
			lines = append(lines, bagTokens)

			// Close the bag
			isInBag = false

			key := bagKey(bagType, bagId)
			bagMap[key] = TokenBag{
				id:     bagId,
				typeOf: bagType,
				forObj: forObj,
				lines:  lines,
			}

			forObj = ""
			bagType = 0
			bagId = 0
			bagTokens = nil

		case isInBag:
			bagTokens = append(bagTokens, cTok)

		case openBag:

			// Open a new bag
			isInBag = true

			bagCat := 0

			switch cmdKwd {
			case "GRANT", "REVOKE", "REASSIGN":
				bagCat = DCLBag
				bagType = DCLBag
			case "COMMENT":
				bagCat = DDLBag
				bagType = CommentOnBag
				forObj = cmdKwd
			}

			bagId = cTok.id
			bagTokens = nil
			bagTokens = []FmtToken{cTok}

			// Add a token that has the pointer to the new bag
			remainder = append(remainder, FmtToken{
				id:         cTok.id,
				categoryOf: bagCat,
				typeOf:     bagType,
				vSpace:     cTok.vSpace,
				indents:    cTok.indents,
				hSpace:     cTok.hSpace,
				vSpaceOrig: cTok.vSpaceOrig,
				hSpaceOrig: cTok.hSpaceOrig,
			})

		default:
			// We are not currently in a bag and we aren't opening one either
			remainder = append(remainder, cTok)
		}

		// set the previous token(s) data
		pTok = cTok
		if !cTok.IsCodeComment() {
			pNcVal = ctVal
		}
	}

	// On the off chance that the bag wasn't closed properly (incomplete or
	// incorrect statement submitted?), ensure that no tokens are lost.
	if len(bagTokens) > 0 {
		key := bagKey(bagType, bagId)
		var lines [][]FmtToken
		lines = append(lines, bagTokens)

		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: bagType,
			forObj: forObj,
			lines:  lines,
		}
	}

	return remainder
}

func UpsertMappedBag(bagMap map[string]TokenBag, bagType, bagId int, forObj string, lines [][]FmtToken) {

	key := bagKey(bagType, bagId)

	_, ok := bagMap[key]
	if ok {
		delete(bagMap, key)
	}

	bagMap[key] = TokenBag{
		id:     bagId,
		typeOf: bagType,
		forObj: forObj,
		lines:  lines,
	}
}

func AdjustLineWrapping(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

	//   Things to wrap
	//   - DML CASE structures with multiple WHEN clauses or with
	//          multiple booleans (AND/OR)
	//   - Function calls with > n named parameters (fat-commas (=>))
	//
	//   - Long lines that contain some combination of
	//       - IF/ELSEIF/ELSIF lines with with multiple booleans
	//       - Long nested function calls
	//       - Multi-element aggregation functions (i.e. concat_ws () or
	//          coalesce () ) with > x elements or > y total length
	//       - Long csv lists of numbers/identifiers/strings
	//
	//
	//
	//   from the perl code:
	//       # 20170921 - I *think* the following might be close to good enough.
	//       # The problem, of course, is now to figure out how to make it so.
	//       #
	//       # 1. Wrap on boolean operators before comparison operators.
	//       #     Additionally, wrap boolean operators at the lowest parens count
	//       #     before moving towards the highest (most deeply nested) parens
	//       #     count.
	//       #
	//       # 3. Wrap on comparison operators before arithmetic operators.
	//       #
	//       # 4. Wrap on arithmetic operators. As with boolean operators, wrap
	//       #     at the lowest parens count before moving towards the highest
	//       #     parens count.
	//       #
	//       # 5. Wrap on concatenation operators.
	//       #
	//       # 6. That still leaves the question of where do longish "IN ( ... )"
	//       #     blocks fit in this?
	//       #
	//       # Create an array of arrays. For each array, if it is too long, then
	//       # take it to the next level of wrapping. Once each array is short
	//       # enough or all wrapping functions have been exhausted then declare
	//       # it done, add new lines/indentation and call it wrapped.
	//       #
	//       # Each wrapping function needs to know how much initial indent there
	//       # is, how much to indent the wraps, and which tokens it is operating
	//       # on. Strings and comments are also needed so that their length may
	//       # be included in line length calculations.
	//
	//        my %comp_ops = map { $_ => $_ } ( '=', '==', '<', '>', '<>', '!=', '>=', '<=' );
	//        my %math_ops = map { $_ => $_ } ( '+', '-', '*', '/' );

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	// TODO:
	// Wrap each line as needed-- where "as needed" is defined along the lines of:
	// - DML lines that contain more than one CASE statement, or <<< are we sure about this one?
	// - DML CASE statements that are less than "simple" (i.e. multiple WHEN clauses, boolean operators), or
	// - lines that have three or more boolean operators, or
	// - lines that have PL calls with three or more named parameters in the call, or
	// - lines that exceed the maxLineLength

	// On the one hand, most instances will only involve on scenario. On the
	// other hand, I have seen production code that involved multiple scenarios
	// in one line so order is going to be important.

	switch bagType {
	case DMLCaseBag:
		wrapDMLCase(e, bagMap, bagType, bagId, defIndents)
	}

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		parensDepth := 0
		initIndents := max(defIndents, line[0].indents)

		switch line[0].AsUpper() {
		case "SELECT":
			initIndents += 2
		}
		for _, cTok := range line {
			switch {
			case cTok.value == "(":
				parensDepth++
			case cTok.value == ")":
				parensDepth--
			case cTok.typeOf == DMLCaseBag:
				AdjustLineWrapping(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
			}
		}
	}

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		parensDepth := 0
		initIndents := line[0].indents
		if initIndents == 0 {
			initIndents = defIndents
		}

		switch line[0].AsUpper() {
		case "SELECT":
			initIndents += 2
		}
		for _, cTok := range line {
			switch {
			case cTok.value == "(":
				parensDepth++
			case cTok.value == ")":
				parensDepth--
			case cTok.typeOf == DMLCaseBag:
			// already done
			case cTok.IsBag():

				wrapBools(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
				wrapPlCalls(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)

				wrapComparisonOps(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
				wrapMathOps(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
				wrapCsvList(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
			}
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func calcBagLen(e *env.Env, bagMap map[string]TokenBag, bagType, bagId int) int {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return 0
	}

	bagLen := 0

	for _, line := range b.lines {
		bagLen += calcLineLen(e, line)
	}
	return bagLen
}

func calcLineLen(e *env.Env, tokens []FmtToken) int {

	lineLen := 0
	for _, cTok := range tokens {
		lineLen += tokenLen(e, cTok)
	}
	return lineLen
}

func tokenLen(e *env.Env, t FmtToken) int {
	if t.vSpace == 0 {
		return len(t.hSpace) + len(t.value)
	}
	return len(strings.Repeat(e.Indent(), t.indents)) + len(t.value)
}

func wrapBools(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

}

func wrapComparisonOps(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

}

func wrapCsvList(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

}

func wrapDMLCase(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

	// Note that it is possible for a line to contain multiple case statements
	// and/or nested case statements.
	// For example:
	//
	//   case
	//       when expression_1 then 1
	//       when expression_2 then 2
	//       when expression_3 then 3
	//       else 4
	//       end
	//
	//   concat_ws ( ':',
	//       case
	//           when expression_1 then 1
	//           when expression_2 then 2
	//           end,
	//       case
	//           when expression_3 then 3
	//           else 4
	//           end )
	//
	//   case expression_1
	//       when a then 1
	//       when b then 2
	//       when c then
	//           case
	//               when expression_2 then 42
	//               else 43
	//               end
	//       else 4
	//       end

	// Fortunately DML case statements have been separated into their own bag

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	var newLines [][]FmtToken
	var newLine []FmtToken

	if len(b.lines) == 0 {
		return
	}

	bagLen := 0
	bagLenMax := 0
	caseLen := 0
	indentDelta := 0
	initIndents := 0
	isDirty := false
	oCount := 0
	oCountMax := 0
	oCounts := 0
	parensDepth := 0
	sbCount := 0
	wCount := 0
	whenI := 0
	wrapCase := false

	var bagLens []int
	var others []int
	var whens [][]FmtToken
	var when []FmtToken

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		// Determine if the indentation needs adjusting
		if initIndents == 0 {
			initIndents = line[0].indents
			if initIndents < defIndents {
				initIndents = defIndents
			}
			indentDelta = initIndents - line[0].indents
		}

		caseLen += calcLineLen(e, line)

		// Get some stats for determining if, and how, the CASE statement needs wrapping
		for idx := 0; idx < len(line); idx++ {
			switch line[idx].AsUpper() {
			case "WHEN", "ELSE":
				whens = append(whens, when)
				others = append(others, oCount)
				bagLens = append(bagLens, bagLen)
				oCountMax = max(oCountMax, oCount)
				oCount = 0
				bagLen = 0
				when = nil
				wCount++
			case "AND", "OR", "IN":
				oCount++
				oCounts++
			case "END":
				whens = append(whens, when)
				when = nil
			default:
				if line[idx].IsBag() {
					sbCount++
					bl := calcBagLen(e, bagMap, line[idx].typeOf, line[idx].id)
					bagLenMax = max(bagLenMax, bl)
					bagLen += bl
				}
			}
			when = append(when, line[idx])
		}
	}

	// Determine if the CASE statement needs wrapping
	switch {
	case wCount+oCountMax > 2:
		wrapCase = true
	case sbCount > 0:
		wrapCase = true
	case caseLen > e.MaxLineLength():
		wrapCase = true
	case bagLenMax > e.MaxLineLength():
		wrapCase = true
	case len(b.lines) > 1:
		wrapCase = true
	}

	if !wrapCase && indentDelta != 0 {
		return
	}

	if wrapCase {

		// Adjust the wrapping and indentation
		var pTok FmtToken

		whenI = 0
		for _, line := range b.lines {

			if len(line) == 0 {
				continue
			}

			whenLen := 0
			oCount = 0

			for idx := 0; idx < len(line); idx++ {
				cTok := line[idx]
				cVspace := cTok.vSpace
				cIndents := cTok.indents

				switch cTok.AsUpper() {
				case "(":
					parensDepth++
				case ")":
					parensDepth--
				case "CASE":
					if cTok.vSpace > 0 {
						cIndents = initIndents + parensDepth
					}
				case "WHEN", "ELSE":
					cVspace = 1
					cIndents = initIndents + parensDepth + 1
					if whenI <= len(whens)-1 {
						when := whens[whenI]
						when[0].vSpace = 1 // pretend that this is a line
						when[0].indents = initIndents + parensDepth + 1
						whenLen = calcLineLen(e, when) + bagLens[whenI]
						oCount = others[whenI]
					}
					whenI++

				case "THEN":
					if whenLen > e.MaxLineLength() {
						cVspace = 1
						cIndents = initIndents + parensDepth + 2
					}
				case "AND", "OR":
					if oCount > 1 {
						cVspace = 1
						cIndents = initIndents + parensDepth + 2
					} else {
						cVspace = 0
					}
				case "END":
					cVspace = 1
					cIndents = initIndents + parensDepth + 1
				default:
					switch {
					case cTok.IsBag():
						AdjustLineWrapping(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth+2)
					case cTok.IsCodeComment(), pTok.IsCodeComment():
					// nada
					default:
						cVspace = 0
						cIndents = 0
						cTok.AdjustHSpace(e, pTok)
					}
				}

				if cTok.vSpace != cVspace || cTok.indents != cIndents {

					cTok.vSpace = cVspace
					cTok.indents = cIndents
					isDirty = true

					if cTok.vSpace > 0 && len(newLine) > 0 {
						newLines = append(newLines, newLine)
						newLine = nil
					}
				}
				pTok = cTok

				newLine = append(newLine, cTok)
			}

			if len(newLine) > 0 {
				newLines = append(newLines, newLine)
				newLine = nil
			}
		}

	} else if indentDelta != 0 {

		// Adjust the indentation
		for _, line := range b.lines {
			for idx := 0; idx < len(line); idx++ {
				cTok := line[idx]
				if cTok.vSpace > 0 {
					cTok.indents += indentDelta
					isDirty = true
				}
				newLine = append(newLine, cTok)
			}

			if len(newLine) > 0 {
				newLines = append(newLines, newLine)
				newLine = nil
			}
		}
	}

	if len(newLine) > 0 {
		newLines = append(newLines, newLine)
		isDirty = true
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
}

func wrapMathOps(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

}

func wrapPlCalls(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents int) {

}
