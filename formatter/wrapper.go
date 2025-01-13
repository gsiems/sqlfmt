package formatter

import (
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
	mathAddSubOps
	mathMultDivOps
	winFuncOps
)

func opsName(i int) string {
	var names = map[int]string{
		compareOps: "compareOps",
		concatOps:  "concatOps",
		logicOps:   "logicOps",
		mathOps:    "mathOps",
		winFuncOps: "winFuncOps",
	}

	if tName, ok := names[i]; ok {
		return tName
	}
	return ""

}

func calcIndent(bagType int, cTok FmtToken) int {

	indents := cTok.indents

	switch bagType {
	case DMLBag:
		switch cTok.AsUpper() {
		case "SELECT", "INSERT":
			indents += 2
		case "FROM", "GROUP BY", "WHERE", "HAVING", "WINDOW", "ORDER BY",
			"OFFSET", "LIMIT", "FETCH", "FOR", "WITH", "VALUES",
			"RETURNING", "CROSS", "FULL", "INNER", "JOIN",
			"LATERAL", "LEFT", "NATURAL", "OUTER", "RIGHT":
			indents++
		}
	case PLxBody:
		switch cTok.AsUpper() {
		case "IF", "CASE", "LOOP":
			indents++
		case "DECLARE", "BEGIN":
			indents++
		case "FOR":
			indents++
		case "WHEN", "THEN", "ELSE":
			indents++
		case "EXCEPTION":
			indents++
		}
	}
	return indents
}

func adjParensDepth(parensDepth, maxParensDepth int, cTok FmtToken) (int, int) {

	switch cTok.value {
	case "(":
		parensDepth++
		if parensDepth > maxParensDepth {
			maxParensDepth = parensDepth
		}
	case ")":
		parensDepth--
	}
	return parensDepth, maxParensDepth
}

func isLogical(pKwVal string, cTok FmtToken) bool {
	switch cTok.AsUpper() {
	case "AND":
		switch pKwVal {
		case "BETWEEN", "PRECEDING", "FOLLOWING", "ROW":
			return false
		default:
			return true
		}
	case "OR":
		return true
	}
	return false
}

func adjLogicalCnt(logicalCnt int, pKwVal string, cTok FmtToken) int {
	if isLogical(pKwVal, cTok) {
		return logicalCnt + 1
	}
	return logicalCnt
}

// calcLen calculates the length of a token
func calcLen(e *env.Env, cTok FmtToken) int {
	// and if token is a pointer to a bag?

	if cTok.vSpace > 0 {
		return len(strings.Repeat(e.Indent(), cTok.indents)) + len(cTok.value)
	}
	return len(cTok.hSpace) + len(cTok.value)
}

func calcSliceLen(e *env.Env, bagType int, tokens []FmtToken) int {

	if len(tokens) == 0 {
		return 0
	}

	idxMax := len(tokens) - 1
	sliceLen := 0

	for idx := 0; idx <= idxMax; idx++ {
		sliceLen += calcLen(e, tokens[idx])
	}
	return sliceLen
}

func calcLenToLineEnd(e *env.Env, bagType int, tokens []FmtToken) int {

	if len(tokens) == 0 {
		return 0
	}

	idxMax := len(tokens) - 1
	sliceLen := 0

	for idx := 0; idx <= idxMax; idx++ {
		if tokens[idx].vSpace > 0 && idx > 0 {
			return sliceLen
		}
		sliceLen += calcLen(e, tokens[idx])
	}
	return sliceLen
}

func validateWhitespacing(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {
	// TODO
	return tokens
	// line feeds after leading comments
	// line feeds after trailing comments of previous tokens
	// sub-select dropping last token
	// closing parens wrapping dml bag
	// window function indentation
	// wrap pl call doesn't wrap if there's a commented line

	indents := 0
	lpd := 0
	idxMax := len(tokens) - 1
	//var pTok FmtToken
	//	lIndent := 0

	//log.Printf("validateWhitespacing %d ##########################", tokens[0].id)

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			indents = calcIndent(bagType, tokens[idx])
			lpd = 0

		}

		switch tokens[idx].value {
		case "(":
			lpd++
		case ")":
			lpd--
		}

		if len(tokens[idx].ledComments) > 0 {

			tokens[idx].EnsureVSpace()
			if tokens[idx].indents == 0 {
				tokens[idx].AdjustIndents(indents + lpd)
			}
		}

		if idx > 0 {
			if len(tokens[idx-1].trlComments) > 0 {
				tokens[idx].EnsureVSpace()
				if tokens[idx].indents == 0 {
					tokens[idx].AdjustIndents(indents + lpd)
				}
			}
		}

		//pTok = tokens[idx]
	}

	return tokens
}

func wrapLines(e *env.Env, bagType int, tokens []FmtToken) (ret []FmtToken) {
	//return tokens
	if len(tokens) == 0 {
		return tokens
	}

	stIdx := 0
	idxMax := len(tokens) - 1
	maxParensDepth := 0
	parensDepth := 0
	//maxDMLCaseDepth := 0
	//dmlCaseDepth := 0

	for idx := 0; idx <= idxMax; idx++ {

		switch bagType {
		case DMLBag:
			parensDepth, maxParensDepth = adjParensDepth(parensDepth, maxParensDepth, tokens[idx])
			// Assert that no supported DB uses END for anything other than
			// CASE statements
			//switch tokens[idx].AsUpper() {
			//case "CASE":
			//	dmlCaseDepth++
			//	if dmlCaseDepth > maxDMLCaseDepth {
			//		maxDMLCaseDepth = dmlCaseDepth
			//	}
			//case "END":
			//	if dmlCaseDepth > 0 {
			//		dmlCaseDepth--
			//	}
			//}
		default:
			parensDepth, maxParensDepth = adjParensDepth(parensDepth, maxParensDepth, tokens[idx])

		}
	}

	switch bagType {
	case DMLBag:
		tokens = wrapValueTuples(e, bagType, tokens)
		tokens = wrapDMLWindowFunctions(e, bagType, maxParensDepth, tokens)
		tokens = wrapPLxCalls(e, bagType, maxParensDepth, tokens)

		switch e.Dialect() {
		case dialect.PostgreSQL:
			tokens = wrapOnMod2Commas(e, bagType, "JSON_BUILD_OBJECT", true, tokens)
		case dialect.Oracle:
			tokens = wrapOnMod2Commas(e, bagType, "DECODE", false, tokens)
		}

		// Note the following need to either be updated to better handle an
		// entire token bag or moved to the line-by line block below (or both)

		tokens = wrapDMLCase(e, bagType, tokens)
		//tokens = wrapDMLLogical(e, bagType, tokens)

	case PLxBody:
		tokens = wrapPLxCalls(e, bagType, maxParensDepth, tokens)
		tokens = wrapPLxLogical(e, bagType, tokens)
	}
	tokens = wrapInto(e, bagType, tokens)

	//////////////////////////////////////////////////
	// return tokens ////////////////////////////////////
	//////////////////////////////////////////////////

	for idx := 0; idx <= idxMax; idx++ {

		eol := false
		switch {
		case idx < idxMax:
			//eol = tokens[idx+1].fbp
			eol = tokens[idx].fbp
		case idx == idxMax:
			eol = true
		}

		if eol && idx > stIdx {
			wt := wrapLine(e, bagType, maxParensDepth, tokens[stIdx:idx])
			ret = append(ret, wt...)
			stIdx = idx
		}
	}
	switch {
	case stIdx < idxMax:
		wt := wrapLine(e, bagType, maxParensDepth, tokens[stIdx:])
		ret = append(ret, wt...)
	case stIdx == idxMax:
		ret = append(ret, tokens[stIdx])
	}
	return ret
}

// wrapLine takes "one lines worth" of tokens and attempts to add line breaks
// as needed
func wrapLine(e *env.Env, bagType, mxPd int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	lineLen := calcSliceLen(e, bagType, tokens)

	if lineLen > e.MaxLineLength() {

		for pdl := 0; pdl <= mxPd; pdl++ {

			// A work in progress...
			// Order matters but may be/is probably context specific...
			// Maybe consider the original vSpace for operators
			//switch pdl {
			//case 0:
			//
			//	tokens = wrapOnCommas(e, bagType, pdl, tokens)
			//	tokens = wrapOnConcatOps(e, bagType, pdl, tokens)
			//	tokens = wrapOnMathOps(e, bagType, pdl, tokens)
			//	tokens = wrapOnCompOps(e, bagType, pdl, tokens)
			//
			//default:

			tokens = wrapOnCompOps(e, bagType, pdl, tokens)
			tokens = wrapOnMathOps(e, bagType, pdl, tokens)
			tokens = wrapOnConcatOps(e, bagType, pdl, tokens)
			tokens = wrapOnCommas(e, bagType, pdl, tokens)
			//}

		}
	}

	return tokens
}

func wrapDMLCase(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

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

	idxMax := len(tokens) - 1
	cdMax := 0
	caseDepth := 0

	// determine the max case depth
	for idx := 0; idx <= idxMax; idx++ {
		switch tokens[idx].AsUpper() {
		case "CASE":
			caseDepth++
			cdMax = max(cdMax, caseDepth)
		case "END":
			caseDepth--
		}
	}

	if cdMax == 0 {
		return tokens
	}

	for cdl := 1; cdl <= cdMax; cdl++ {

		caseDepth = 0
		idxStart := 0
		idxWhen := 0
		indents := 0
		idxThens := make(map[int]int)
		whenLens := make(map[int]int)
		whenLgcl := make(map[int][]int)
		var idxWhens []int

		for idx := 0; idx <= idxMax; idx++ {

			switch tokens[idx].AsUpper() {
			case "CASE":
				caseDepth++
			}

			if caseDepth < cdl {
				if tokens[idx].vSpace > 0 {
					indents = calcIndent(bagType, tokens[idx])
				}
			}

			if caseDepth == cdl {

				switch tokens[idx].AsUpper() {
				case "CASE":
					idxStart = idx
					idxWhens = nil

				case "WHEN", "ELSE":
					if idxWhen > 0 {
						whenLens[idxWhen] = calcSliceLen(e, bagType, tokens[idxWhen:idx])
					}
					idxWhens = append(idxWhens, idx)

					idxWhen = idx
				case "THEN":
					idxThens[idxWhen] = idx

				case "END":
					whenLens[idxWhen] = calcSliceLen(e, bagType, tokens[idxWhen:idx])
					idxEnd := idx
					lenCase := calcSliceLen(e, bagType, tokens[idxStart:idxEnd])
					lenToEnd := 0
					// TODO: if the next token is "AS" then include lenToEnd?
					if idx < idxMax {
						switch tokens[idx+1].AsUpper() {
						case "AS": //, "||", "+", "-", "/", "*" :
							lenToEnd = calcLenToLineEnd(e, bagType, tokens[idxEnd:])
						}
					}

					addBreaks := false
					switch {
					case len(idxWhens) > 2:
						addBreaks = true
					case lenCase+lenToEnd > e.MaxLineLength():
						addBreaks = true
					}

					if addBreaks {

						tpi := indents
						tpd := 0
						leadLen := len(strings.Repeat(e.Indent(), tpi))

						for i := idxStart; i <= idxEnd; i++ {

							switch tokens[i].value {
							case "(":
								tpd++
							case ")":
								tpd--
							}

							if whenLen, ok := whenLens[i]; ok {
								tokens[i].EnsureVSpace()
								tokens[i].AdjustIndents(tpi + 1)

								if leadLen+whenLen > e.MaxLineLength() {
									if thenIdx, ok := idxThens[i]; ok {
										tokens[thenIdx+1].EnsureVSpace()
										tokens[thenIdx+1].AdjustIndents(tpi + 2)
									}
								}

								if idxLgcls, ok := whenLgcl[i]; ok {
									if len(idxLgcls) > 1 {
										for _, j := range idxLgcls {
											tokens[j].EnsureVSpace()
											tokens[j].AdjustIndents(tpi + tpd + 2)
										}
									}
								}
							}
						}
						tokens[idxEnd].EnsureVSpace()
						tokens[idxEnd].AdjustIndents(tpi + 1)

						indents = calcIndent(bagType, tokens[idx]) - 1
					}
				default:
					if idx > 0 {
						if isLogical(tokens[idx-1].AsUpper(), tokens[idx]) {
							whenLgcl[idxWhen] = append(whenLgcl[idxWhen], idx)
						}
					}
				}
			}

			switch tokens[idx].AsUpper() {
			case "END":
				if caseDepth == cdl {
					idxStart = idx
				}
				caseDepth--
			}
		}
	}
	return tokens
}

func wrapDMLLogical(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	idxMax := len(tokens) - 1
	indents := 0
	parensDepth := 0
	pKwVal := ""

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			indents = calcIndent(bagType, tokens[idx])
		}

		switch tokens[idx].AsUpper() {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		case "AND", "OR":
			if isLogical(pKwVal, tokens[idx]) {
				tokens[idx].EnsureVSpace()
				tokens[idx].AdjustIndents(indents + parensDepth)
			}
		}

		if tokens[idx].IsKeyword() {
			pKwVal = tokens[idx].AsUpper()
		}
	}
	return tokens
}

func wrapDMLWindowFunctions(e *env.Env, bagType, mxPd int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}
	if mxPd == 0 {
		return tokens
	}

	idxMax := len(tokens) - 1

	for pdl := 1; pdl <= mxPd; pdl++ {
		cCnt := 0
		idxStart := 0
		indents := 0
		lSegLen := 0
		parensDepth := 0

		for idx := 0; idx <= idxMax; idx++ {

			cTok := tokens[idx]

			if cTok.value == "(" {
				parensDepth++
				if parensDepth == pdl {
					cCnt = 0
					idxStart = idx
				}
			}

			if parensDepth < pdl {
				if cTok.vSpace > 0 {
					indents = calcIndent(bagType, cTok)
					lSegLen = calcLenToLineEnd(e, bagType, tokens[idx:])
				}
			}

			//if parensDepth == pdl && lSegLen > e.MaxLineLength() {
			if parensDepth == pdl {

				switch cTok.value {
				case ")":
					if cCnt > 1 {
						idxEnd := idx
						tpd := pdl
						if lSegLen > e.MaxLineLength() {
							for i := idxStart + 1; i < idxEnd; i++ {

								switch tokens[i].value {
								case "(":
									tpd++
								case ")":
									tpd--
								}

								if tpd == pdl {
									switch tokens[i].value {
									case "ORDER BY", "GROUP BY", "PARTITION BY":
										tokens[i].EnsureVSpace()
										tokens[i].AdjustIndents(indents + pdl)
										//default:
										//	if tokens[i].vSpace > 0 {
										//		tokens[i].AdjustIndents(tpi + tpd)
										//	}
									}
								}
							}
						}
					}
				case "ORDER BY", "GROUP BY", "PARTITION BY":
					cCnt++
				}
			}

			if cTok.value == ")" {
				parensDepth--
			}
		}
	}
	return tokens

}

func wrapOnMod2Commas(e *env.Env, bagType int, fcnName string, wrapEven bool, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	cCnt := 0
	idxStart := 0
	indents := 0
	inFcn := false
	fcnParensDepth := 0
	parensDepth := 0

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			if !inFcn {
				indents = calcIndent(bagType, cTok)
			}
		}
		ptVal := ""
		if idx > 0 {
			ptVal = tokens[idx-1].AsUpper()
		}

		switch cTok.value {
		case "(":
			parensDepth++
			if ptVal == fcnName {
				fcnParensDepth = parensDepth
				inFcn = true
				cCnt = 0
				idxStart = idx
			}
		case ")":
			if fcnParensDepth == parensDepth {
				inFcn = false
				fcnParensDepth = 0
			}
			parensDepth--
		}

		switch {
		case inFcn:
			switch ptVal {
			case ",":
				cCnt++
				addWrap := false
				switch {
				case wrapEven && cCnt%2 == 0:
					addWrap = true
				case !wrapEven && cCnt%2 == 1:
					addWrap = true
				}
				if addWrap {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents + parensDepth)
					if tokens[idxStart+1].vSpace == 0 {
						tokens[idxStart+1].EnsureVSpace()
						tokens[idxStart+1].AdjustIndents(indents + parensDepth)
					}
				}
			}
		}
	}
	return tokens
}

func wrapInto(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	// If there is only one column then there is nothing to wrap
	// If there are more than one column then each element should be on a separate line
	indents := 0
	inInto := false
	intoParensDepth := 0
	parensDepth := 0
	idxStart := 0
	ppKwVal := ""
	pKwVal := ""
	var intoIdxs []int

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
		}

		switch cTok.value {
		case "(":
			parensDepth++
			if ppKwVal == "INSERT" && pKwVal == "INTO" {
				intoParensDepth = parensDepth
				inInto = true
				intoIdxs = nil
				idxStart = idx
			}

		case ",":
			if inInto && intoParensDepth == parensDepth {
				if idx < idxMax {
					intoIdxs = append(intoIdxs, idx+1)
				}
			}

		case ")":
			if intoParensDepth == parensDepth {
				if len(intoIdxs) > 0 {
					for _, i := range intoIdxs {
						tokens[i].EnsureVSpace()
						tokens[i].AdjustIndents(indents + parensDepth - 1)
					}
					tokens[idxStart+1].EnsureVSpace()
					tokens[idxStart+1].AdjustIndents(indents + parensDepth - 1)
				}

				inInto = false
				idxStart = 0
				intoIdxs = nil
				intoParensDepth = 0
			}
			parensDepth--
		}

		if cTok.IsKeyword() && pKwVal != "AS" {
			ppKwVal = pKwVal
			pKwVal = cTok.AsUpper()
		}
	}
	return tokens
}

func wrapOnCommas(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {

	/*

	   func calcSliceLen(e *env.Env, bagType int, tokens []FmtToken) int {

	   	sliceLen := 0
	   	for _, cTok := range tokens {
	   		sliceLen += calcLen(e, cTok)
	   	}
	   	return sliceLen
	   }

	   func calcLenToLineEnd(e *env.Env, bagType int, tokens []FmtToken) int {


	*/
	/*
	   LINE       => the set of tokens passed into the wrap function
	   line       => a set of tokens bounded by vSpaces or LINE boundaries
	   parens set => a pair of matching open/close parens and the tokens contained therein
	   pSetLen    => length of tokens within a parens set
	   lSegLen    => the length of tokens starting from the last vSpace to the next vSpace (or end of LINE)
	   lCurLen    => the length of tokens starting from the last vSpace to the current token
	   lRemLen    => the length of tokens from the current token to the next vSpace (or end of LINE)

	   for each pdl from 0 => maxParensDepth

	       - check commas, then comparison ops, then math ops, and finally concatenation ops

	       - for each parens set for the pdl

	           - when checking commas

	               - if the lSegLen < MaxLineLength then skip the current parens set
	               - if there are any fat-commas then skip the current parens set
	               - if is part of JSON_BUILD_OBJECT then skip the current parens set
	               - if is preceded by IN then wrap at length (as needed)
	               - If is part of DCL or DDL then wrap at length (as needed)

	*/

	if len(tokens) == 0 {
		return tokens
	}

	idxLineStart := 0
	idxMax := len(tokens) - 1
	idxStart := 0
	idxEnd := 0
	indents := 0
	formatTokens := true
	parensDepth := 0
	pKwVal := ""
	ppKwVal := ""
	wrapOnLength := false
	lSegLen := 0
	ptVal := ""
	cCnt := 0
	wrapOnOpenParens := false
	debug := false
	disableWrapping := false
	isWindowFunction := false

	if debug {
		log.Printf("wrapOnCommas bagType: %s, pdl: %d, len(tokens): %d", nameOf(bagType), pdl, len(tokens))
	}

	// RAISE NOTICE etc. wrap on length???

	if pdl == 0 {
		lSegLen = calcLenToLineEnd(e, bagType, tokens)

		for idx := 0; idx <= idxMax; idx++ {

			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				lSegLen = calcLenToLineEnd(e, bagType, tokens[idx:])
			}

			switch tokens[idx].value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			default:
				if idx > 0 && parensDepth == pdl && lSegLen > e.MaxLineLength() {
					switch tokens[idx-1].value {
					case ",":
						if tokens[idx].vSpace == 0 {
							tokens[idx].EnsureVSpace()
							tokens[idx].AdjustIndents(indents + parensDepth + 1)
						}
					}
				}
			}
		}
		return tokens
	}

	// pdl is > 0 /////////////////////////////////////////////////////////////
	for idx := 0; idx <= idxMax; idx++ {

		addBreaks := false
		doCheck := false

		if parensDepth < pdl {
			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				idxLineStart = idx
			}
		}

		switch tokens[idx].value {
		case "(":
			parensDepth++

			if parensDepth == pdl {
				cCnt = 0
				idxStart = idx

				if idx > 0 {
					ptVal = tokens[idx-1].AsUpper()
				} else {
					ptVal = ""
				}

				switch pKwVal {
				case "VALUES":
					// taken care of by wrapValueTuples
					disableWrapping = true
				case "INTO", "INSERT":
					wrapOnOpenParens = true
				case "ORDER BY", "GROUP BY", "PARTITION BY":
					isWindowFunction = true
				case "CALL", "ROW":
					if e.Dialect() == dialect.PostgreSQL {
						wrapOnOpenParens = true
					}
				}

				switch {
				case isWindowFunction:
					wrapOnLength = true
				case ppKwVal == "INSERT" && pKwVal == "INTO":
					formatTokens = false
				case ptVal == "JSON_BUILD_OBJECT":
					formatTokens = false
				case ptVal == "DECODE":
					formatTokens = false
				case ptVal == "IN":
					wrapOnLength = true

				default:
					switch bagType {
					case DDLBag, DCLBag, CommentOnBag:
						wrapOnLength = true
					}
				}
			}
		}

		// if we are in the pdl AND there are commas within the pdl prior to the current token,
		// AND if we hit a vSpace AND the line up to this point is too long THEN wrap
		// OR we hit the pdl end AND the line up to this point is too long is too long then wrap

		// if idxStart is less than idxLineStart then wrap starting at the idxLineStart
		// if idxStart is greater than idxLineStart then wrap starting at the idxStart

		switch tokens[idx].value {
		case ",":
			if parensDepth == pdl {
				cCnt++
			}
		case "=>":
			if parensDepth == pdl {
				formatTokens = false
			}
		case ";":
			disableWrapping = false
		case ")":
			switch {
			case parensDepth == pdl:
				doCheck = !disableWrapping
			case parensDepth < pdl:
				disableWrapping = false
				isWindowFunction = false
				wrapOnOpenParens = false
				wrapOnLength = false
				formatTokens = true
			}
		}

		if doCheck {
			idxEnd = idx
			lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxEnd])
			if debug {
				log.Printf("    %d        doCheck tokens[%d]: id: %d, idxLineStart: %d, idxStart: %d, idxEnd: %d, cCnt: %d [%s]", tokens[0].id, idx, tokens[idx].id, idxLineStart, idxStart, idxEnd, cCnt, tokens[idx].value)
				log.Printf("    %d        doCheck cCnt: %d, lSegLen: %d, formatTokens: %t", tokens[0].id, cCnt, lSegLen, formatTokens)
			}
			addBreaks = false
			if formatTokens && cCnt > 0 {
				lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxEnd]) + 1

				//lineLen := calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
				switch {
				case lSegLen > e.MaxLineLength():
					addBreaks = true
					//case lineLen > e.MaxLineLength():
					//addBreaks = true
				}
			}

			if debug {
				log.Printf("    %d        doCheck addBreaks: %t", tokens[0].id, addBreaks)
			}
		}

		if addBreaks {

			if wrapOnLength {
				tpl := calcSliceLen(e, bagType, tokens[idxLineStart:idxStart+1])
				tpd := pdl
				for i := idxStart + 1; i < idxEnd; i++ {
					switch tokens[i].value {
					case "(":
						tpd++
					case ")":
						tpd--
					default:
						if tpd != pdl {
							continue
						}
						// if the previous was a comma, or was the start,
						// then scan ahead to find the next comma where
						// tpd == pdl (or the closing parens that drops
						// below pdl). If the length between goes over the
						// limit then add a vSpace to the current token.

						if i != idxStart+1 && tokens[i-1].value != "," {
							continue
						}

						jpd := pdl
						jdone := false

						for j := i + 1; j < idxEnd; j++ {
							if jdone {
								break
							}
							switch tokens[j].value {
							case "(":
								jpd++
							case ")":
								jpd--
							default:
								if jpd != pdl {
									continue
								}
								if j+1 != idxEnd && tokens[j-1].value != "," {
									continue
								}
								nl := calcSliceLen(e, bagType, tokens[i:j])

								if tpl+nl > e.MaxLineLength() {
									tokens[i].EnsureVSpace()
									tokens[i].AdjustIndents(indents + pdl)
									tpl = 0
								}
								jdone = true
							}
							tpl += calcLen(e, tokens[i]) + 1
						}
					}
				}

			} else {

				tpd := pdl

				for i := idxStart + 1; i < idxEnd; i++ {

					switch tokens[i].value {
					case "(":
						tpd++
					case ")":
						tpd--
					default:
						if tpd == pdl {

							// wrapOnOpenParens
							switch tokens[i-1].value {
							case ",":
								tokens[i].EnsureVSpace()
								tokens[i].AdjustIndents(indents + pdl)
							case "(":
								if wrapOnOpenParens {
									tokens[i].EnsureVSpace()
									tokens[i].AdjustIndents(indents + pdl)
								}
							}
						}
					}
				}
			}
		}

		switch tokens[idx].value {
		case ")":
			if parensDepth == pdl {
				cCnt = 0
				formatTokens = false
			}
			parensDepth--
		}

		if tokens[idx].IsKeyword() && pKwVal != "AS" {
			ppKwVal = pKwVal
			pKwVal = tokens[idx].AsUpper()
		}
	}
	return tokens
}

func wrapOnCompOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	//log.Print("wrapOnCompOps")
	return wrapOnOps(e, bagType, compareOps, pdl, tokens)
}

func wrapOnConcatOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	//log.Print("wrapOnConcatOps")
	return wrapOnOps(e, bagType, concatOps, pdl, tokens)
}

func wrapOnMathOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	//log.Print("wrapOnMathOps")
	tokens = wrapOnOps(e, bagType, mathAddSubOps, pdl, tokens)
	return wrapOnOps(e, bagType, mathMultDivOps, pdl, tokens)
}

func wrapOnOps(e *env.Env, bagType, opType, pdl int, tokens []FmtToken) []FmtToken {
	//log.Printf("    wrapOnOps (%d)", pdl)

	if len(tokens) == 0 {
		return tokens
	}

	//addBreaks := false
	cCnt := 0
	idxEnd := 0
	idxLineStart := 0
	idxStart := 0
	indents := 0
	lSegLen := 0
	parensDepth := 0
	pSetLen := 0
	lineLen := 0
	idxMax := len(tokens) - 1

	if pdl == 0 {

		lSegLen = calcLenToLineEnd(e, bagType, tokens)
		//log.Printf("    len(tokens): %d, lSegLen: %d", len(tokens), lSegLen)

		for idx := 0; idx <= idxMax; idx++ {

			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				lSegLen = calcLenToLineEnd(e, bagType, tokens[idx:])
			}

			switch tokens[idx].value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}

			if parensDepth == pdl && lSegLen > e.MaxLineLength() {
				addBreak := false
				switch tokens[idx].value {
				case "||":
					addBreak = opType == concatOps
				case "+", "-":
					addBreak = opType == mathAddSubOps
				case "*", "/":
					addBreak = opType == mathMultDivOps
				case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
					addBreak = opType == compareOps
				}
				if addBreak {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents + parensDepth + 1)
				}
			}
		}

		return tokens
	}

	// pdl is > 0 /////////////////////////////////////////////////////////////

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]
		doCheck := false
		//if cTok.vSpace > 0 {
		//	indents = calcIndent(bagType, cTok)
		//	idxLineStart = idx
		//}

		if parensDepth < pdl {
			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, cTok)

				lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
				idxLineStart = idx
			}
		}

		if cTok.value == "(" {
			parensDepth++
			if parensDepth == pdl {
				cCnt = 0
			}
		}

		if parensDepth == pdl {

			switch cTok.value {
			case "(":
				idxStart = idx
				if idxStart > idxLineStart {
					lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxStart])
				} else {
					lSegLen = 0
				}

			case ")":
				idxEnd = idx
				if idxEnd > idxStart {
					pSetLen = calcSliceLen(e, bagType, tokens[idxStart:idxEnd])
				} else {
					pSetLen = 0
				}

				// determine if concat operators were found and if the line is too long
				if cCnt > 0 && lSegLen+pSetLen > e.MaxLineLength() {
					doCheck = true
				}
				if cCnt > 0 && lineLen > e.MaxLineLength() {
					doCheck = true
				}

			case "||":
				if opType == concatOps {
					cCnt++
				}
			case "+", "-":
				if opType == mathAddSubOps {
					cCnt++
				}
			case "*", "/":
				if opType == mathMultDivOps {
					cCnt++
				}
			case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
				if opType == compareOps {
					cCnt++
				}
			}
		}

		if cTok.value == ")" {
			parensDepth--
		}

		//log.Printf("lSegLen: %d, indents: %d, pdl: %d, idxStart: %d, idxEnd: %d, addBreaks: %t", lSegLen, indents, pdl, idxStart, idxEnd, addBreaks)

		if !doCheck {
			continue
		}
		doCheck = false

		///////////////////////////////////////////

		//log.Printf("lSegLen: %d, indents: %d, pdl: %d, idxStart: %d, idxEnd: %d", lSegLen, indents, pdl, idxStart, idxEnd)

		tpi := indents
		tpd := pdl
		for i := idxStart + 1; i < idxEnd; i++ {

			if tokens[i].vSpace > 0 {
				tpi = calcIndent(bagType, tokens[i])
			}

			switch tokens[i].value {
			case "(":
				tpd++
			case ")":
				tpd--
			}

			if tpd == pdl {
				addBreak := false
				switch tokens[i].value {
				case "||":
					addBreak = opType == concatOps
				case "+", "-":
					addBreak = opType == mathAddSubOps
				case "*", "/":
					addBreak = opType == mathMultDivOps
				case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
					addBreak = opType == compareOps
				}
				if addBreak {
					tokens[i].EnsureVSpace()
					tokens[i].AdjustIndents(tpi + tpd + 1)
				}
			}
		}
	}
	return tokens
}

func wrapPLxCalls(e *env.Env, bagType, mxPd int, tokens []FmtToken) []FmtToken {

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

	if len(tokens) == 0 {
		return tokens
	}
	if mxPd == 0 {
		return tokens
	}

	switch e.Dialect() {
	case dialect.PostgreSQL, dialect.Oracle:
		// nada
	default:
		return tokens
	}

	idxMax := len(tokens) - 1

	for pdl := 1; pdl <= mxPd; pdl++ {
		fcCnt := 0
		parensDepth := 0
		idxStart := 0
		indents := 0

		for idx := 0; idx <= idxMax; idx++ {

			cTok := tokens[idx]

			if cTok.value == "(" {
				parensDepth++
				if parensDepth == pdl {
					fcCnt = 0
					idxStart = idx
				}
			}

			if parensDepth < pdl {
				if cTok.vSpace > 0 {
					indents = calcIndent(bagType, cTok)
				}
			}

			if parensDepth == pdl {

				switch cTok.value {
				case ")":
					if fcCnt > 1 {
						idxEnd := idx
						tpi := indents
						tpd := pdl

						for i := idxStart + 1; i < idxEnd; i++ {
							switch tokens[i].value {
							case "(":
								tpd++
							case ")":
								tpd--
							}

							switch {
							case tpd == pdl:
								switch tokens[i-1].value {
								case "(", ",":
									tokens[i].EnsureVSpace()
									tokens[i].AdjustIndents(tpi + tpd)
								default:
									if tokens[i].vSpace > 0 {
										tokens[i].AdjustIndents(tpd + tokens[i].indents)
									}
								}
							case tpd > pdl:
								if tokens[i].vSpace > 0 {
									tokens[i].AdjustIndents(tpi + tpd + tokens[i].indents)
								}
							}
						}
					}
				case "=>":
					fcCnt++
				}
			}

			if cTok.value == ")" {
				parensDepth--
			}
		}
	}
	return tokens
}

func wrapPLxLogical(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	indents := 0
	inLogical := false
	logicalIndents := 0
	logicalStart := 0
	pKwVal := ""
	lineLen := 0

	idxMax := len(tokens) - 1
	//log.Printf("wrapPLxLogical  [%s] [%s]", tokens[0].value, tokens[idxMax].value)

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			indents = tokens[idx].indents
			lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
		}

		switch tokens[idx].AsUpper() {

		case "IF", "ELSIF", "ELSEIF", "WHEN":
			//log.Printf("wrapPLxLogical   [%s]", tokens[idx].value)
			inLogical = true
			logicalStart = idx
			logicalIndents = indents
		case "THEN":
			if inLogical {
				logicalCnt := 0
				//logicalLen := calcSliceLen(e, bagType, tokens[logicalStart:idx])
				pkv := pKwVal

				for i := logicalStart; i <= idx; i++ {

					switch tokens[i].AsUpper() {
					case "AND", "OR":
						logicalCnt = adjLogicalCnt(logicalCnt, pkv, tokens[i])
					}

					if tokens[i].IsKeyword() {
						pkv = tokens[i].AsUpper()
					}
				}

				splitOnLogical := false

				//log.Printf("wrapPLxLogical   logicalCnt: %d, logicalLen > e.MaxLineLength(): %t", logicalCnt, logicalLen > e.MaxLineLength())

				switch logicalCnt {
				case 0:
					// nada
				case 1, 2, 3:
					//splitOnLogical = logicalLen > e.MaxLineLength()
					splitOnLogical = lineLen > e.MaxLineLength()
				default:
					splitOnLogical = true
				}

				if splitOnLogical {

					pkv = ""
					//ipd := logicalIndents + 1
					//ipd := logicalIndents
					ipd := 0

					for i := logicalStart; i <= idx; i++ {

						switch tokens[i].AsUpper() {
						case "(":
							ipd++
						case ")":
							ipd--
							//if i > logicalStart {
							if tokens[i].vSpace > 0 {
								tokens[i].AdjustIndents(ipd - 1)
							}
							//}
						case "AND", "OR":
							if isLogical(pkv, tokens[i]) {
								tokens[i].EnsureVSpace()
								//if ipd == 0 {
								//	tokens[i].AdjustIndents(logicalIndents + 1)
								//} else {
								tokens[i].AdjustIndents(logicalIndents + ipd + 1)
								//}
							}
						default:
							if i > logicalStart {
								if tokens[i].vSpace > 0 {
									tokens[i].AdjustIndents(ipd + tokens[i].indents)
								}
							}
						}

						if tokens[i].IsKeyword() {
							pkv = tokens[i].AsUpper()
						}
					}
				}
			}
			logicalStart = 0
			inLogical = false
		}

		if tokens[idx].IsKeyword() {
			pKwVal = tokens[idx].AsUpper()
		}
	}

	return tokens
}

func wrapValueTuples(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	// If there is only one values tuple then there is nothing to wrap (other
	// than the elements in the tuple)

	// If there are more than one values tuple then each tuple should be on a
	// separate line and the elements within should be wrapped according to
	// the e.WrapMultiTuples() value

	idxStart := 0
	indents := 0
	inValues := false
	parensDepth := 0
	pdl := 0
	ptVal := ""
	tplStart := 0
	indentLen := 0
	tplWrapping := 0
	hasMultiTuples := false
	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
			indentLen = len(strings.Repeat(e.Indent(), indents))
		}

		switch cTok.value {
		case "(":
			parensDepth++

			switch ptVal {
			case "VALUES":
				pdl = parensDepth
				inValues = true
				idxStart = idx
				tplStart = idx
			case ",":
				if parensDepth == pdl && inValues {
					tplStart = idx
					//tplCnt++
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents)
					if tokens[idxStart].vSpace == 0 {
						tokens[idxStart].EnsureVSpace()
						tokens[idxStart].AdjustIndents(indents)
					}
				}
			}

		case ")":
			switch {
			case parensDepth == pdl:
				if inValues {
					// Wrap, or not, the elements within the tuple

					// TODO: If a values statement has but one tuple and that
					// tuple has less than n elements (n TBD) then does it
					// need to be wrapped? Wrapping one element does seem silly
					// so n > 1...

					// Check the next token to determine if there are multiple
					// tuples involved and, if so, determine how to wrap them
					if tplWrapping == 0 {
						if idx < idxMax {
							if tokens[idx+1].value == "," {
								hasMultiTuples = true
								tplWrapping = e.WrapMultiTuples()
							}
						}
						// if there is only one tuple then it gets wrapped regardless
						if !hasMultiTuples {
							tplWrapping = env.WrapAll
						}
					}

					wrapTuple := false
					switch tplWrapping {
					case env.WrapNone:
						// nada
					case env.WrapAll:
						wrapTuple = true
					default:
						// Wrap Auto
						tplLen := calcSliceLen(e, bagType, tokens[tplStart:idx])
						if tplLen+indentLen > e.MaxLineLength() {
							wrapTuple = true
						}
					}

					if wrapTuple {
						tpd := 0
						for i := tplStart; i <= idx; i++ {
							switch tokens[i].value {
							case "(":
								tpd++
								switch tokens[i-1].value {
								case ",":
									// in the off chance that there is a parenthetical value in the tuple
									if tpd == 2 {
										tokens[i].EnsureVSpace()
										tokens[i].AdjustIndents(indents + 1)
									}
								}
							case ")":
								tpd--
							default:
								switch tokens[i-1].value {
								case ",", "(":
									if tpd == 1 {
										tokens[i].EnsureVSpace()
										tokens[i].AdjustIndents(indents + 1)
									}
								}
							}
						}
					}
				}

				if ptVal == ")" {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents)
				}
			case parensDepth < pdl:
				if ptVal == ")" {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents)
				}
				inValues = false
				idxStart = 0
				pdl = 0
				tplStart = 0
			}
			parensDepth--
		case ";":
			inValues = false
			idxStart = 0
			pdl = 0
		}
		ptVal = tokens[idx].value
	}

	return tokens
}
