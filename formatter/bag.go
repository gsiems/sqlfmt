package formatter

import (
	"fmt"
	"log"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

const (
	compareOps = iota + 400
	concatOps
	logicOps
	mathOps
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
	/*
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
	*/
	// wrap IN
	// comp_ops   "=", "==", "<", ">", "<>", "!=", ">=", "<="
	// math_ops   "+", "-", "*", "/"
	// concat_ops "||"
	// logic_ops  "AND", "OR"
	// start at parensDepth == 0, increment and re-run as needed

	for pdl := 0; pdl <= 5; pdl++ {

		//wrapCsv(e, bagMap, bagType, bagId, defIndents, pdl)

		wrapCsv2(e, bagMap, bagType, bagId, defIndents, pdl)

		//wrapOps(e, bagMap, bagType, bagId, defIndents, pdl, logicOps)
		//wrapOps(e, bagMap, bagType, bagId, defIndents, pdl, concatOps)

		//wrapOps(e, bagMap, bagType, bagId, defIndents, pdl, compareOps)
		//wrapOps(e, bagMap, bagType, bagId, defIndents, pdl, mathOps)
	}

	for pdl := 1; pdl <= 5; pdl++ {
		wrapPlCalls(e, bagMap, bagType, bagId, defIndents, pdl)
	}

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		parensDepth := 0
		initIndents := calcIndents(bagType, parensDepth, line[0])
		initIndents = max(defIndents, line[0].indents)

		for _, cTok := range line {
			switch {
			case cTok.value == "(":
				parensDepth++
			case cTok.value == ")":
				parensDepth--
			case cTok.typeOf == DMLCaseBag:
				// already done
			case cTok.IsBag():
				AdjustLineWrapping(e, bagMap, cTok.typeOf, cTok.id, initIndents+parensDepth)
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

func min(a, b int) int {
	if a < b {
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
		bagLen += calcLineLen(e, bagMap, line)
	}
	return bagLen
}

func calcLineLen(e *env.Env, bagMap map[string]TokenBag, tokens []FmtToken) int {

	lineLen := 0
	for _, cTok := range tokens {
		lineLen += tokenLen(e, bagMap, cTok)
	}
	return lineLen
}

func lenToLineEnd(e *env.Env, bagMap map[string]TokenBag, tokens []FmtToken) (int, bool) {

	lineLen := 0

	if len(tokens) == 0 {
		return 0, false
	}

	for idx, cTok := range tokens {
		switch {
		case cTok.vSpace > 0:
			if idx > 0 {
				return lineLen, true
			} else {
				lineLen += len(strings.Repeat(e.Indent(), cTok.indents)) + len(cTok.value)
			}
		case cTok.IsBag():
			key := bagKey(cTok.typeOf, cTok.id)

			b, ok := bagMap[key]
			if !ok {
				return 0, false
			}

			for _, line := range b.lines {
				switch {
				case len(line) == 0:
					// nada
				case line[0].vSpace > 0:
					return lineLen, true
				default:
					tl, done := lenToLineEnd(e, bagMap, line)
					lineLen += tl
					if done {
						return lineLen, done
					}
				}
			}
		default:
			lineLen += len(cTok.hSpace) + len(cTok.value)
		}
	}
	return lineLen, false
}

func tokenLen(e *env.Env, bagMap map[string]TokenBag, t FmtToken) int {

	tl := 0
	switch {
	case t.IsBag():
		tl = calcBagLen(e, bagMap, t.typeOf, t.id)
	default:
		tl = len(t.value)
	}

	if t.vSpace == 0 {
		return len(t.hSpace) + tl
	}
	return len(strings.Repeat(e.Indent(), t.indents)) + tl
}

func wrapOps(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents, pdl, opsType int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.lines) == 0 {
		return
	}

	var newLines [][]FmtToken
	var newLine []FmtToken

	isDirty := false
	parensDepth := 0

	for _, line := range b.lines {

		if len(line) == 0 {
			continue
		}

		lineLen := calcLineLen(e, bagMap, line)
		tooLong := lineLen > e.MaxLineLength()

		if !tooLong {
			newLines = append(newLines, line)
			continue
		}

		idxMax := len(line) - 1
		pKwVal := ""
		parensDepth = 0

		initIndents := calcIndents(bagType, parensDepth, line[0])
		initIndents = max(initIndents, defIndents)

		for idx := 0; idx <= idxMax; idx++ {

			cTok := line[idx]
			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}

			matches := false
			switch opsType {
			case compareOps:
				switch cTok.value {
				case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
					matches = true
				}
			case concatOps:
				switch cTok.value {
				case "||":
					matches = true
				}
			case mathOps:
				switch cTok.value {
				case "+", "-", "*", "/":
					matches = true
				}
			case logicOps:
				switch cTok.AsUpper() {
				case "OR":
					switch pKwVal {
					case "CREATE":
					// nada
					default:
						matches = true
					}
				case "AND":
					switch pKwVal {
					case "BETWEEN", "PRECEDING", "FOLLOWING", "ROW":
					// nada
					default:
						matches = true
					}
				}
			}

			if matches && parensDepth == pdl {
				if len(newLine) > 0 {
					newLines = append(newLines, newLine)
					newLine = nil
				}
				cTok.EnsureVSpace()
				cTok.AdjustIndents(initIndents + parensDepth + 1)
				isDirty = true
			}

			if !cTok.IsKeyword() {
				pKwVal = cTok.AsUpper()
			}

			newLine = append(newLine, cTok)
		}

		if len(newLine) > 0 {
			newLines = append(newLines, newLine)
			newLine = nil
		}
	}

	if len(newLine) > 0 {
		newLines = append(newLines, newLine)
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
}

func wrapCsv(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents, pdl int) {

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	var newLines [][]FmtToken

	if len(b.lines) == 0 {
		return
	}

	isDirty := false
	inPdl := pdl == 0
	inPre := !inPdl

	for _, line := range b.lines {

		if len(line) == 0 {
			continue
		}

		lineLen, _ := lenToLineEnd(e, bagMap, line)
		tooLong := lineLen > e.MaxLineLength()

		if !tooLong {
			newLines = append(newLines, line)
			continue
		}

		idxMax := len(line) - 1
		pNcVal := ""
		parensDepth := 0

		initIndents := calcIndents(bagType, parensDepth, line[0])
		initIndents = max(initIndents, defIndents)

		cCnt := 0
		isIN := false

		var csi [][]FmtToken
		var csin []FmtToken
		var pre []FmtToken
		var post []FmtToken

		for idx := 0; idx <= idxMax; idx++ {

			cTok := line[idx]

			switch {
			case inPre:
				pre = append(pre, cTok)
			case inPdl:
				if isIN {
					csin = append(csin, cTok)
				} else {
					for len(csi) <= cCnt {
						csi = append(csi, []FmtToken{})
					}
					csi[cCnt] = append(csi[cCnt], cTok)
				}
			default:
				// post
				post = append(post, cTok)
			}

			switch cTok.value {
			case "(":
				parensDepth++
				inPdl = parensDepth >= pdl

				if parensDepth == pdl && !isIN {
					isIN = pNcVal == "IN"
				}

				if inPdl {
					inPre = false
				}

			case ")":
				parensDepth--
				inPdl = parensDepth >= pdl
				if !inPdl {
					isIN = false
				}

			case ",":
				if parensDepth == pdl {
					cCnt++
				}
			}

			if !cTok.IsCodeComment() {
				pNcVal = cTok.AsUpper()
			}
		}

		if len(csin) > 0 || len(csi) > 0 {
			log.Printf("    len(cs*) %d %d", len(csin), len(csi))

			for _, t := range pre {
				log.Printf("        pre [%s]", t.value)
			}

			for _, t := range post {
				log.Printf("        post [%s]", t.value)
			}
		}

		switch {
		case len(csin) > 0:

			var acc []FmtToken
			acc = append(acc, pre...)
			cumLen, _ := lenToLineEnd(e, bagMap, pre)

			for i := 0; i < len(csin); i++ {
				cTok := csin[i]

				tl := 0
				switch {
				case cTok.IsBag():

					key := bagKey(cTok.typeOf, cTok.id)

					b, ok := bagMap[key]
					if ok {
						if len(b.lines) > 0 {
							tl, _ = lenToLineEnd(e, bagMap, b.lines[0])

						}
					}

				default:
					tl = tokenLen(e, bagMap, cTok)
				}

				cumLen += tl

				switch {
				case cTok.value == ",":
					acc = append(acc, cTok)

				case cumLen >= e.MaxLineLength():
					newLines = append(newLines, acc)
					isDirty = true
					cTok.EnsureVSpace()
					cTok.AdjustIndents(initIndents + max(pdl, 1))

					acc = nil
					acc = append(acc, cTok)
					cumLen = tl

				default:
					acc = append(acc, cTok)
				}
			}

			postLen, _ := lenToLineEnd(e, bagMap, post)

			switch {
			case cumLen+postLen == 0:
			// nada
			case postLen == 0:
				newLines = append(newLines, acc)
			case cumLen == 0:
				newLines = append(newLines, post)
			case cumLen+postLen > e.MaxLineLength():
				newLines = append(newLines, acc)
				post[0].EnsureVSpace()
				post[0].AdjustIndents(initIndents + max(pdl, 1))
				newLines = append(newLines, post)
			default:
				acc = append(acc, post...)
				newLines = append(newLines, acc)
			}

		case len(csi) > 0:

			if len(pre) > 0 {
				newLines = append(newLines, pre)
			}

			for i := 0; i < len(csi); i++ {

				isDirty = true
				cToks := csi[i]

				if len(cToks) == 0 {
					continue
				}

				switch i {
				case 0:
					// The first line
					if len(pre) > 0 {
						cToks[0].EnsureVSpace()
						cToks[0].AdjustIndents(initIndents + max(pdl, 1))
					}
					newLines = append(newLines, cToks)

				case len(csi) - 1:
					// The last line
					cToks[0].EnsureVSpace()
					cToks[0].AdjustIndents(initIndents + max(pdl, 1))
					cLen, _ := lenToLineEnd(e, bagMap, cToks)
					postLen, _ := lenToLineEnd(e, bagMap, post)

					switch {
					case cLen+postLen == 0:
						// nada
					case postLen == 0:
						newLines = append(newLines, cToks)
					case cLen == 0:
						post[0].EnsureVSpace()
						post[0].AdjustIndents(initIndents + max(pdl, 1))
						newLines = append(newLines, post)
					case cLen+postLen >= e.MaxLineLength():
						// cToks and post are separate lines
						newLines = append(newLines, cToks)
						post[0].EnsureVSpace()
						post[0].AdjustIndents(initIndents + max(pdl, 1))
						newLines = append(newLines, post)
					default:
						// cToks and post are one line
						post[0].vSpace = 0
						post[0].indents = 0
						post[0].hSpace = " "
						cToks = append(cToks, post...)
						newLines = append(newLines, cToks)
					}

				default:
					cToks[0].EnsureVSpace()
					cToks[0].AdjustIndents(initIndents + pdl)
					newLines = append(newLines, cToks)
				}
			}
		default:
			// both csi and csin are empty
			newLines = append(newLines, line)
		}
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
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

	if bagType != DMLCaseBag {
		return
	}

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	/*
	   log.Print("#########################################################")
	   		for _, line := range b.lines {
	   			for _, t := range line {
	   				ts := t.String()
	   				log.Printf("%6d: %s", bagId, ts)
	   			}
	   		}
	*/

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
	/*
	   log.Print("#########################################################")
	*/
	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		// Determine if the indentation needs adjusting
		initIndents = max(line[0].indents, defIndents)
		indentDelta = initIndents - line[0].indents
		/*
		   log.Printf(" %d   line[0].indents: %d, defIndents: %d, initIndents: %d, indentDelta: %d",
		   	bagId, line[0].indents, defIndents, initIndents, indentDelta)
		*/

		/*
			if initIndents == 0 {
				initIndents = line[0].indents
				if initIndents < defIndents {
					initIndents = defIndents
				}
				indentDelta = initIndents - line[0].indents
			}
		*/
		caseLen += calcLineLen(e, bagMap, line)

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

	/*
	   log.Print("#########################################################")
	   log.Printf(" %d   wCount: %d, len(whens): %d, oCountMax: %d, sbCount: %d, caseLen: %d, bagLenMax: %d, len(b.lines): %d, indentDelta: %d",
	   	bagId, wCount, len(whens), oCountMax, sbCount, caseLen, bagLenMax, len(b.lines), indentDelta)
	*/

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

	if !wrapCase && indentDelta == 0 {
		return
	}

	if wrapCase {

		// Adjust the wrapping and indentation
		var pTok FmtToken

		whenI = 1
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
					if cVspace > 0 {
						cIndents = initIndents + parensDepth
					}
				case "WHEN", "ELSE":
					cVspace = 1
					cIndents = initIndents + parensDepth + 1
					/*
					   log.Printf(" %d   [%s] whenI: %d, id: %d, cVspace: %d, cIndents: %d",
					   	bagId, cTok.AsUpper(), whenI, cTok.id, cVspace, cIndents)
					*/
					if whenI < len(whens)-1 {
						when := whens[whenI]
						when[0].EnsureVSpace() // pretend that this is a line
						when[0].AdjustIndents(initIndents + parensDepth + 1)
						whenLen = calcLineLen(e, bagMap, when) + bagLens[whenI]
						oCount = others[whenI]

						/*
							for _, t := range when {
								log.Printf(" %d      %6d %d %s", bagId, cTok.id, whenI, t.String())
							}
						*/
					}
					/*
					   log.Printf(" %d      whenLen: %d", bagId, whenLen)
					*/

					whenI++

				case "THEN":
					/*
					   log.Printf(" %d   [%s] ID: %d, whenLen: %d, cVspace: %d, cIndents: %d",
					   	bagId, cTok.AsUpper(), cTok.id, whenLen, cVspace, cIndents)
					*/

					if whenLen > e.MaxLineLength() {
						cVspace = 1
						cIndents = initIndents + parensDepth + 2
					} else if cVspace > 0 {
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
						if cVspace > 0 {
							cIndents = initIndents + parensDepth + 2
						}
					default:
						cVspace = 0
						cIndents = 0
						cTok.AdjustHSpace(e, pTok)
					}
				}

				if cTok.vSpace != cVspace || cTok.indents != cIndents {
					isDirty = true
					if !cTok.IsBag() {
						cTok.vSpace = cVspace
						cTok.AdjustIndents(cIndents)
					}

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

func isOp(opsType int, op, pKwVal string) bool {

	var ops []string

	switch opsType {
	case compareOps:
		ops = []string{"=", "==", "<", ">", "<>", "!=", ">=", "<="}
	case concatOps:
		ops = []string{"||"}
	case mathOps:
		ops = []string{"+", "-", "*", "/"}
	case logicOps:
		switch strings.ToUpper(op) {
		case "OR":
			switch strings.ToUpper(pKwVal) {
			case "CREATE":
			// nada
			default:
				return true
			}
		case "AND":
			switch strings.ToUpper(pKwVal) {
			case "BETWEEN", "PRECEDING", "FOLLOWING", "ROW":
			// nada
			default:
				return true
			}
		}
	}

	if len(ops) > 0 {
		for _, v := range ops {
			if op == v {
				return true
			}
		}
	}

	return false
}

func wrapOps2(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents, pdl, opsType int) {
	return

	if pdl == 0 {
		return
	}

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.lines) == 0 {
		return
	}

	isDirty := false
	parensDepth := 0
	pKwVal := ""

	var newLines [][]FmtToken

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		opCnt := 0 // count of operators

		var newLine []FmtToken
		for _, cTok := range line {
			newLine = append(newLine, cTok)

			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			default:
				tst := isOp(opsType, cTok.value, pKwVal)
				if tst {
					//log.Printf("   isOp(%d, %q, %q) == %t", opsType, cTok.value, pKwVal, tst)
					opCnt++
				}
			}

		}

		//log.Printf("   opCnt: %d", opCnt)

		if opCnt > 0 {

			log.Printf("%d    opCnt: %d, defIndents: %d, pdl: %d", bagId, opCnt, defIndents, pdl)

			indents := defIndents
			idStart := 0
			opCnt = 0  // Count of operators
			lnCnt := 0 // Count of new lines
			lineLen := 0
			pKwVal := ""

			for idx := 0; idx < len(newLine); idx++ {

				// If we are at the start of a new line then determine the
				// indentation for line wrapping and also scan ahead to
				// determine the length of the line
				if newLine[idx].vSpace > 0 {

					setIndents := false
					switch pdl {
					case 0:
						setIndents = parensDepth == pdl
					default:
						setIndents = parensDepth < pdl
					}

					if setIndents {
						indents = calcIndents(bagType, parensDepth, newLine[idx])
					}
					lineLen, _ = lenToLineEnd(e, bagMap, newLine[idx:])

					log.Printf("%d        indents: %d, lineLen: %d", newLine[idx].id, indents, lineLen)

				}

				addBreaks := false

				switch newLine[idx].value {
				case "(":
					parensDepth++
					indents++
					if parensDepth == pdl {
						idStart = idx
					}
				case ")":
					if parensDepth == pdl {

						switch {
						case opCnt <= 1:
						// nada
						case opCnt > 1:
							addBreaks = true
						case lnCnt > 0:
							addBreaks = true
						case lineLen >= e.MaxLineLength():
							addBreaks = true
						}

						log.Printf("%d            (%d - %d) opCnt: %d, lnCnt: %d, lineLen: %d, addBreaks: %t", newLine[idx].id, idStart, idx, opCnt, lnCnt, lineLen, addBreaks)

						if !addBreaks {
							idStart = 0
							opCnt = 0
							lnCnt = 0
						}
					}
					indents--
					parensDepth--

				default:

					if pdl == 0 {
						if newLine[idx].vSpace > 0 {
							if idStart == 0 {
								idStart = idx
							}
							if idx > idStart {
								switch {
								case opCnt <= 1:
								// nada
								case opCnt > 1:
									addBreaks = true
								case lnCnt > 0:
									addBreaks = true
								case lineLen >= e.MaxLineLength():
									addBreaks = true
								}
							}

						}

					}

					tst := isOp(opsType, newLine[idx].value, pKwVal)

					if parensDepth == pdl {
						if tst {
							opCnt++
						}
						if newLine[idx].vSpace > 0 {
							lnCnt++
						}
					}
					if tst {
						log.Printf("%d                isOp: %t - parensDepth: %d, pdl: %d, opCnt: %d, lnCnt: %d", newLine[idx].id, tst, parensDepth, pdl, opCnt, lnCnt)
					}
				}

				if addBreaks {
					isDirty = true
					tpd := pdl
					tpi := indents
					pkv := pKwVal

					for i := idStart + 1; i < idx; i++ {
						if newLine[i].value == "(" {
							tpd++
							tpi++
						}

						if tpd == pdl {
							if isOp(opsType, newLine[i].value, pkv) {
								newLine[i].EnsureVSpace()
								newLine[i].AdjustIndents(tpi)
							} else {
								if newLine[i].vSpace > 0 {
									newLine[i].AdjustIndents(tpi)
								}
							}
						}

						if newLine[i].value == ")" {
							tpd--
							tpi--
						}

						if newLine[i].IsKeyword() {
							pkv = newLine[i].value
						}
					}
					idStart = 0
					opCnt = 0
					lnCnt = 0
				}

				if newLine[idx].IsKeyword() {
					pKwVal = newLine[idx].value
				}

			}
		}
		newLines = append(newLines, newLine)
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
}

func calcIndents(bagType, parensDepth int, cTok FmtToken) int {

	indents := cTok.indents + parensDepth

	if bagType == DMLBag {
		switch cTok.AsUpper() {
		case "SELECT":
			indents += 2
		case "FROM", "WHERE", "GROUP", "HAVING", "WINDOW",
			"ORDER", "OFFSET", "LIMIT", "FETCH", "FOR", "WITH",
			"RETURNING", "CROSS", "FULL", "INNER", "JOIN",
			"LATERAL", "LEFT", "NATURAL", "OUTER", "RIGHT",
			"VALUES":
			indents++
		}
	}
	return indents
}

func wrapCsv2(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents, pdl int) {

	if pdl == 0 {
		return
	}

	// do not perform for "COMMENT ON...", "ALTER...", or "CREATE..."
	switch bagType {
	case DCLBag, DDLBag, CommentOnBag:
		return
	}

	// if json_build_object then wrap every second comma

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.lines) == 0 {
		return
	}

	isDirty := false
	parensDepth := 0
	pNcVal := ""

	var newLines [][]FmtToken

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		cCnt := 0 // count of "fat-commas"

		var newLine []FmtToken
		for _, cTok := range line {
			newLine = append(newLine, cTok)

			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			case ",":
				if parensDepth == pdl {
					cCnt++
				}
			}
		}

		if cCnt > 0 {

			indents := defIndents
			idStart := 0
			cCnt = 0   // Count of "fat-commas"
			lnCnt := 0 // Count of new lines
			lineLen := 0
			isIn := false

			for idx := 0; idx < len(newLine); idx++ {

				// If we are at the start of a new line then determine the
				// indentation for line wrapping and also scan ahead to
				// determine the length of the line
				if newLine[idx].vSpace > 0 {

					setIndents := false
					switch pdl {
					case 0:
						setIndents = parensDepth == pdl
					default:
						setIndents = parensDepth < pdl
					}

					if setIndents {
						indents = calcIndents(bagType, parensDepth, newLine[idx])
					}
					lineLen, _ = lenToLineEnd(e, bagMap, newLine[idx:])
				}

				switch newLine[idx].value {
				case "(":
					parensDepth++
					indents++
					if parensDepth == pdl {
						idStart = idx
						if pNcVal == "IN" {
							isIn = true
						}
					}
				case ")":
					if parensDepth == pdl {

						addBreaks := false
						switch {
						case isIn:
							// nada
						case lineLen < e.MaxLineLength():
							// nada
						case cCnt <= 1:
						// nada
						case cCnt > 1:
							addBreaks = true
						case lnCnt > 0:
							addBreaks = true
						}

						if addBreaks {
							isDirty = true
							tpd := pdl
							tpi := indents

							pnv := newLine[idStart].value

							for i := idStart + 1; i < idx; i++ {
								if newLine[i].value == "(" {
									tpd++
									tpi++
								}

								if tpd == pdl {
									switch pnv {
									case "(", ",":
										newLine[i].EnsureVSpace()
										newLine[i].AdjustIndents(tpi)
									default:
										if newLine[i].vSpace > 0 {
											newLine[i].AdjustIndents(tpi)
										}
									}
								}

								if newLine[i].value == ")" {
									tpd--
									tpi--
								}
								if !newLine[i].IsCodeComment() {
									pnv = newLine[i].value
								}
							}
						}
						idStart = 0
						cCnt = 0
						lnCnt = 0
						isIn = false
					}
					indents--
					parensDepth--
				case ",":
					if parensDepth == pdl {
						cCnt++
					}
				default:
					if newLine[idx].vSpace > 0 {
						if parensDepth == pdl {
							lnCnt++
						}
					}
				}

				if !newLine[idx].IsCodeComment() {
					pNcVal = newLine[idx].value
				}

			}

		}
		newLines = append(newLines, newLine)
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
}

func wrapPlCalls(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, defIndents, pdl int) {

	// Note that it is possible for a line of code to contain multiple PL calls
	// and/or for a PL call to contain nested PL calls
	// For example:
	//
	//   select coalesce ( func_01 ( ... ), func_02 ( ... ) ) ;
	//
	//   var := func_01 (
	//           param_1 => 1,
	//           param_2 => func_02 ( ... ),
	//           param_3 => 42 ) ;

	key := bagKey(bagType, bagId)

	b, ok := bagMap[key]
	if !ok {
		return
	}

	if len(b.lines) == 0 {
		return
	}

	switch e.Dialect() {
	case dialect.PostgreSQL, dialect.Oracle:
	// nada
	default:
		return
	}

	isDirty := false
	parensDepth := 0

	var newLines [][]FmtToken

	for _, line := range b.lines {
		if len(line) == 0 {
			continue
		}

		fcCnt := 0 // count of "fat-commas"

		var newLine []FmtToken
		for _, cTok := range line {
			newLine = append(newLine, cTok)

			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			case "=>":
				if parensDepth == pdl {
					fcCnt++
				}
			}
		}

		if fcCnt > 0 {

			indents := defIndents
			idStart := 0
			fcCnt = 0  // Count of "fat-commas"
			lnCnt := 0 // Count of new lines
			lineLen := 0

			for idx := 0; idx < len(newLine); idx++ {

				// If we are at the start of a new line then determine the
				// indentation for line wrapping and also scan ahead to
				// determine the length of the line
				if newLine[idx].vSpace > 0 {

					setIndents := false
					switch pdl {
					case 0:
						setIndents = parensDepth == pdl
					default:
						setIndents = parensDepth < pdl
					}

					if setIndents {
						indents = calcIndents(bagType, parensDepth, newLine[idx])
					}
					lineLen, _ = lenToLineEnd(e, bagMap, newLine[idx:])
				}

				switch newLine[idx].value {
				case "(":
					parensDepth++
					indents++
					if parensDepth == pdl {
						idStart = idx
					}
				case ")":
					if parensDepth == pdl {

						addBreaks := false
						switch {
						case fcCnt <= 1:
						// nada
						case fcCnt > 1:
							addBreaks = true
						case lnCnt > 0:
							addBreaks = true
						case lineLen >= e.MaxLineLength():
							addBreaks = true
						}

						if addBreaks {
							isDirty = true
							tpd := pdl
							tpi := indents

							pNcVal := newLine[idStart].value

							for i := idStart + 1; i < idx; i++ {
								if newLine[i].value == "(" {
									tpd++
									tpi++
								}

								if tpd == pdl {
									switch pNcVal {
									case "(", ",":
										newLine[i].EnsureVSpace()
										newLine[i].AdjustIndents(tpi)
									default:
										if newLine[i].vSpace > 0 {
											newLine[i].AdjustIndents(tpi)
										}
									}
								}

								if newLine[i].value == ")" {
									tpd--
									tpi--
								}
								if !newLine[i].IsCodeComment() {
									pNcVal = newLine[i].value
								}
							}
						}
						idStart = 0
						fcCnt = 0
						lnCnt = 0
					}
					indents--
					parensDepth--
				case "=>":
					if parensDepth == pdl {
						fcCnt++
					}
				default:
					if newLine[idx].vSpace > 0 {
						if parensDepth == pdl {
							lnCnt++
						}
					}
				}
			}
		}
		newLines = append(newLines, newLine)
	}

	if isDirty {
		UpsertMappedBag(bagMap, b.typeOf, b.id, "", newLines)
	}
}
