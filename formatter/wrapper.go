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
	winFuncOps
)

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
		case "WHEN", "THEN", "ELSE":
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
			// nada
		default:
			return true
		}
	case "OR":
		return true
	}
	return false
}

func adjLogicalCnt(logicalCnt int, pKwVal string, cTok FmtToken) int {

	switch cTok.AsUpper() {
	case "AND":

		//log.Printf("adjLogicalCnt [%s] [%s] ", pKwVal, cTok.value)

		switch pKwVal {
		case "BETWEEN", "PRECEDING", "FOLLOWING", "ROW":
			// nada
		default:
			return logicalCnt + 1
		}
	case "OR":
		//log.Printf("adjLogicalCnt [%s]", cTok.value)
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

	for idx := 0; idx <= idxMax; idx++ {

		eol := false
		switch {
		case idx < idxMax:
			eol = tokens[idx].vSpace > 0
		case idx == idxMax:
			eol = true
		}

		if eol && idx > stIdx {
			wt := wrapLine(e, bagType, tokens[stIdx:idx])
			ret = append(ret, wt...)
			stIdx = idx
		}
	}
	switch {
	case stIdx < idxMax:
		wt := wrapLine(e, bagType, tokens[stIdx:])
		ret = append(ret, wt...)
	case stIdx == idxMax:
		ret = append(ret, tokens[stIdx])
	}

	return ret
}

// wrapLine takes "one lines worth" of tokens and attempts to add line breaks
// as needed
func wrapLine(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	//log.Printf("  wrapLine  (%d, %d) [%s] [%s]", tokens[0].id, tokens[len(tokens)-1].id, tokens[0].value, tokens[len(tokens)-1].value)

	// ASSERTION: The line is properly indented prior to wrapping.
	// Therefore, the only parensDepth that should matter is the depth within
	// the line itself. This does imply that negative depths are possible.

	fatCommaCnt := 0
	intoCnt := 0
	jboCnt := 0
	dmlLogicalCnt := 0
	plxLogicalCnt := 0
	maxParensDepth := 0
	parensDepth := 0
	pKwVal := ""
	valuesCnt := 0
	winFncCnt := 0
	dmlCaseCnt := 0

	lineLen := calcSliceLen(e, bagType, tokens)
	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {
		cTok := tokens[idx]
		parensDepth, maxParensDepth = adjParensDepth(parensDepth, maxParensDepth, cTok)

		switch cTok.AsUpper() {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		case "AND", "OR":
			switch bagType {
			case DMLBag:
				dmlLogicalCnt = adjLogicalCnt(dmlLogicalCnt, pKwVal, cTok)
			case PLxBody:
				plxLogicalCnt = adjLogicalCnt(plxLogicalCnt, pKwVal, cTok)
			}

		case "CASE":
			if bagType == DMLBag {
				dmlCaseCnt++
			}

		case "=>":
			fatCommaCnt++
		case "JSON_BUILD_OBJECT":
			jboCnt++
		case "VALUES":
			valuesCnt++
		case "INTO":
			intoCnt++
		case "ORDER BY", "GROUP BY", "PARTITION BY":
			if parensDepth > 0 {
				winFncCnt++
			}
		}

		if cTok.IsKeyword() {
			pKwVal = cTok.AsUpper()
		}
	}

	if valuesCnt > 0 {
		tokens = wrapValueTuples(e, bagType, tokens)
	}

	switch bagType {
	case PLxBody:
		if plxLogicalCnt > 0 {
			tokens = wrapPLxLogical(e, bagType, tokens)
		}
	case DMLBag:
		if winFncCnt > 0 {
			tokens = wrapDMLWindowFunctions(e, bagType, maxParensDepth, tokens)
		}

		if dmlCaseCnt > 0 {
			tokens = wrapDMLCase(e, bagType, maxParensDepth, tokens)
		}

		if dmlLogicalCnt > 0 {
			tokens = wrapDMLLogical(e, bagType, tokens)
		}

		if jboCnt > 0 {
			tokens = wrapJSONBuildObject(e, bagType, tokens)
		}
	}

	if fatCommaCnt > 1 {
		tokens = wrapPLxCalls(e, bagType, maxParensDepth, tokens)
	}

	if intoCnt > 0 {
		tokens = wrapInto(e, bagType, tokens)
	}

	if lineLen > e.MaxLineLength() {

		for pdl := 0; pdl <= maxParensDepth; pdl++ {

			tokens = wrapOnConcatOps(e, bagType, pdl, tokens)
			tokens = wrapOnCommas(e, bagType, pdl, tokens)
			tokens = wrapOnCompOps(e, bagType, pdl, tokens)
			tokens = wrapOnMathOps(e, bagType, pdl, tokens)

		}

	}

	return validateWhitespacing(e, bagType, tokens)
}

func wrapDMLCase(e *env.Env, bagType, mxPd int, tokens []FmtToken) []FmtToken {

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

	// Fortunately DML case statements have been separated into their own bag

	caseDepth := 0
	caseIdx := 0
	idxMax := len(tokens) - 1
	indents := 0
	parensDepth := 0
	pKwVal := ""
	var caseIdxs intStack

	caseWraps := make(map[int]bool)
	caseIndents := make(map[int]int)
	logicCnts := make(map[int]int)
	whenCnts := make(map[int]int)

	lineLen := calcSliceLen(e, bagType, tokens)

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			indents = calcIndent(bagType, tokens[idx])
		}

		switch tokens[idx].AsUpper() {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		case "CASE":
			caseIdx = idx
			caseIdxs.Push(idx)
			caseDepth++
		case "WHEN", "ELSE":
			whenCnts[caseIdx]++
		case "AND", "OR":
			if isLogical(pKwVal, tokens[idx]) {
				logicCnts[caseIdx]++
			}
		case "END":
			if caseDepth > 0 {
				//caseLen := calcSliceLen(e, bagType, tokens[caseIdx:idx])
				//caseLens[caseIdx] = caseLen

				whenCnt, _ := whenCnts[caseIdx]
				logicCnt, _ := logicCnts[caseIdx]

				switch {
				case whenCnt > 2, logicCnt > 0, lineLen > e.MaxLineLength():
					caseWraps[caseIdx] = true
					caseIndents[caseIdx] = indents + parensDepth + caseDepth
				default:
					caseWraps[caseIdx] = false
				}

				caseDepth--
			}

			_ = caseIdxs.Pop()
			caseIdx = caseIdxs.Peek()
		}
		if tokens[idx].IsKeyword() {
			pKwVal = tokens[idx].AsUpper()
		}
	}

	caseIdxs.Reset()
	addWraps := false
	cIndent := 0
	for idx := 0; idx <= idxMax; idx++ {

		switch tokens[idx].AsUpper() {
		case "CASE":
			caseIdx = idx
			caseIdxs.Push(idx)
			addWraps, _ = caseWraps[caseIdx]
			cIndent, _ = caseIndents[caseIdx]
		case "AND", "OR":
			if isLogical(pKwVal, tokens[idx]) {
				if addWraps {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(cIndent + 1)
				}
			}
		case "WHEN", "ELSE", "END":
			if addWraps {
				tokens[idx].EnsureVSpace()
				tokens[idx].AdjustIndents(cIndent)
			}
		}

		switch tokens[idx].AsUpper() {
		case "END":
			_ = caseIdxs.Pop()
			caseIdx = caseIdxs.Peek()
			if caseIdxs.Len() > 0 {
				addWraps, _ = caseWraps[caseIdx]
				cIndent, _ = caseIndents[caseIdx]
			} else {
				addWraps = false
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
		//lSegLen := 0
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
					//lSegLen = calcLenToLineEnd(e, bagType, tokens[idx:])
				}
			}

			//if parensDepth == pdl && lSegLen > e.MaxLineLength() {
			if parensDepth == pdl {

				switch cTok.value {
				case ")":
					if cCnt > 1 {
						idxEnd := idx
						tpd := pdl
						lSegLen := calcLenToLineEnd(e, bagType, tokens[idxStart:idxEnd])
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

func wrapJSONBuildObject(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	switch e.Dialect() {
	case dialect.PostgreSQL:
		// nada
	default:
		return tokens
	}

	// If there is only one element then there is nothing to wrap
	// If there are more than one element then each element should be on a separate line

	cCnt := 0
	idxStart := 0
	indents := 0
	inJBO := false
	jboParensDepth := 0
	parensDepth := 0

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
		}
		ptVal := ""
		if idx > 0 {
			ptVal = tokens[idx-1].AsUpper()
		}

		switch cTok.value {
		case "(":
			parensDepth++
			if ptVal == "JSON_BUILD_OBJECT" {
				jboParensDepth = parensDepth
				inJBO = true
				cCnt = 0
				idxStart = idx
			}
		case ")":
			if jboParensDepth == parensDepth {
				inJBO = false
				jboParensDepth = 0
			}
			parensDepth--
		}

		if inJBO {
			switch ptVal {
			case ",":
				cCnt++
				if cCnt%2 == 0 {
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

	//addBreaks := false
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
	//pSetLen := 0
	lSegLen := 0
	//lMaxSegLen := 0
	ptVal := ""
	cCnt := 0
	//var cIdxs []int
	wrapOnOpenParens := false
	debug := false

	if debug {
		log.Printf("%d    pdl: %d, len(tokens): %d [%s]", tokens[0].id, pdl, len(tokens), tokens[0].value)
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
						tokens[idx].EnsureVSpace()
						tokens[idx].AdjustIndents(indents + parensDepth + 1)
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
				formatTokens = true
				wrapOnLength = false
				wrapOnOpenParens = false
				idxStart = idx

				if idx > 0 {
					ptVal = tokens[idx-1].AsUpper()
				} else {
					ptVal = ""
				}

				switch {
				case ppKwVal == "INSERT" && pKwVal == "INTO":
					formatTokens = false
				case ptVal == "JSON_BUILD_OBJECT":
					formatTokens = false
				case ptVal == "IN":
					wrapOnLength = true

				default:
					switch pKwVal {
					case "VALUES":
						wrapOnOpenParens = true
					case "CALL":
						if e.Dialect() == dialect.PostgreSQL {
							wrapOnOpenParens = true
						}
					}
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

		if parensDepth == pdl {

			switch tokens[idx].value {
			case ",":
				cCnt++
			case "=>":
				formatTokens = false
			case ")":
				doCheck = true
				if idxLineStart > idxStart {
					idxStart = idxLineStart
				}

			}
		}

		if parensDepth >= pdl {
			if tokens[idx].vSpace > 0 {
				doCheck = true
				if idxLineStart > idxStart {
					idxStart = idxLineStart
				}
			}
		}

		if doCheck {
			idxEnd = idx
			lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxEnd])
			if debug {
				log.Printf("%d        doCheck tokens[%d]: id: %d, idxLineStart: %d, idxStart: %d, idxEnd: %d, cCnt: %d [%s]", tokens[0].id, idx, tokens[idx].id, idxLineStart, idxStart, idxEnd, cCnt, tokens[idx].value)
				log.Printf("%d        doCheck cCnt: %d, lSegLen: %d, formatTokens: %t", tokens[0].id, cCnt, lSegLen, formatTokens)
			}
			if formatTokens && cCnt > 0 {
				lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxEnd])
				addBreaks = lSegLen > e.MaxLineLength()
			}

			if debug {
				log.Printf("%d        doCheck addBreaks: %t", tokens[0].id, addBreaks)
			}
		}

		if addBreaks {

			if wrapOnLength {
				lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxStart])
				cIdx := idxStart + 1
				nextLens := make(map[int]int)
				tpd := pdl
				var cIdxs []int
				tplZero := len(strings.Repeat(e.Indent(), indents+pdl))

				for i := idxStart + 1; i < idxEnd; i++ {
					switch tokens[i].value {
					case "(":
						tpd++
					case ")":
						tpd--
					default:
						if tpd == pdl {
							if tokens[i-1].value == "," {
								cIdxs = append(cIdxs, cIdx)
								nextLens[cIdx] = calcSliceLen(e, bagType, tokens[cIdx:i])
								cIdx = i
							}
						}
					}
				}
				if cIdx < idxEnd {
					cIdxs = append(cIdxs, cIdx)
					nextLens[cIdx] = calcSliceLen(e, bagType, tokens[cIdx:idxEnd])
				}

				tpl := lSegLen
				for _, i := range cIdxs {
					nl, ok := nextLens[i]
					if ok {
						if tpl+nl > e.MaxLineLength() {
							tokens[i].EnsureVSpace()
							tokens[i].AdjustIndents(indents + pdl)
							tpl = tplZero
						}
						tpl += nl
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

		if parensDepth >= pdl {
			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				idxLineStart = idx
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

	///////////////////////////////////////
	/*
		for idx := 0; idx <= idxMax; idx++ {


			cTok := tokens[idx]

			if cTok.value == "(" {
				parensDepth++

				if parensDepth == pdl {
					formatTokens = true
					wrapOnLength = false
					cCnt = 0
					if idx > 0 {
						ptVal = tokens[idx-1].AsUpper()
					} else {
						ptVal = ""
					}

					switch {
					case ppKwVal == "INSERT" && pKwVal == "INTO":
						formatTokens = false
					case ptVal == "JSON_BUILD_OBJECT":
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

			if cTok.vSpace > 0 {
				idxLineStart = idx

				if parensDepth < pdl {
					indents = calcIndent(bagType, cTok)
					idxLineStart = idx
					lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idx])
				}

				if parensDepth == pdl {

					if idx > idxLineStart {
						lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idx])
						if lSegLen > lMaxSegLen {
							lMaxSegLen = lSegLen
						}
						idxLineStart = idx
					}
				}

			}

			if parensDepth == pdl {

				if formatTokens {

					switch cTok.value {
					case "(":
						idxStart = idx
						if idxStart > idxLineStart {
							lSegLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxStart])
							if lSegLen > lMaxSegLen {
								lMaxSegLen = lSegLen
							}
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

						// determine if commas were found and if the line is too long
						if cCnt > 0 && lMaxSegLen+pSetLen > e.MaxLineLength() {
							addBreaks = true
						}

						//log.Printf("formatTokens: %t, wrapOnLength: %t, addBreaks: %t", formatTokens, wrapOnLength, addBreaks)

					case ",":
						cCnt++
					case "=>":
						formatTokens = false
					}
				}
			}

			if cTok.value == ")" {
				parensDepth--
			}

			if cTok.IsKeyword() && pKwVal != "AS" {
				ppKwVal = pKwVal
				pKwVal = cTok.AsUpper()
			}

			if !addBreaks {
				continue
			}
			addBreaks = false

			///////////////////////////////////////////

			if wrapOnLength {
				cIdx := idxStart
				nextLens := make(map[int]int)
				tpd := pdl

				for i := idxStart + 1; i < idxEnd; i++ {

					switch tokens[i].value {
					case "(":
						tpd++
					case ")":
						tpd--
					default:
						if tpd == pdl && i > 0 {
							if tokens[i-1].value == "," {
								nextLens[cIdx] = calcSliceLen(e, bagType, tokens[cIdx:i])
								cIdx = i
							}
						}
					}
				}
				if cIdx < idxEnd {
					nextLens[cIdx] = calcSliceLen(e, bagType, tokens[cIdx:idxEnd])
				}

				tpl := lSegLen
				tpi := indents
				for i := idxStart; i < idxEnd; i++ {

					if tokens[i].vSpace > 0 {
						tpi = calcIndent(bagType, tokens[i])
						tpl = 0
					}

					nl, ok := nextLens[i]
					if ok {

						if tpl+nl > e.MaxLineLength() {

							//log.Printf("lSegLen: %d, indents: %d, pdl: %d, tpi: %d, tpl: %d, nextLen: %d", lSegLen, indents, pdl, tpi, tpl, nl)
							//log.Printf("     [%s] -- [%s]", tokens[], tokens[], tokens[])

							tokens[i].EnsureVSpace()
							tokens[i].AdjustIndents(tpi + pdl)
							tpl = len(strings.Repeat(e.Indent(), tokens[i].indents))
						}

						tpl += nl
					}
				}
				continue
			}

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
					switch tokens[i-1].value {
					case ",", "(":
						//case ",":
						tokens[i].EnsureVSpace()
						tokens[i].AdjustIndents(tpi + tpd)
					default:
						if tokens[i].vSpace > 0 {
							tokens[i].AdjustIndents(tpi + tpd)
						}
					}
				}

			}

		}
	*/
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
	return wrapOnOps(e, bagType, mathOps, pdl, tokens)
}

func wrapOnOps(e *env.Env, bagType, opType, pdl int, tokens []FmtToken) []FmtToken {
	//log.Printf("    wrapOnOps (%d)", pdl)

	if len(tokens) == 0 {
		return tokens
	}

	addBreaks := false
	cCnt := 0
	idxEnd := 0
	idxLineStart := 0
	idxStart := 0
	indents := 0
	lSegLen := 0
	parensDepth := 0
	pSetLen := 0

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
				case "+", "-", "*", "/":
					addBreak = opType == mathOps
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

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
			idxLineStart = idx
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
					addBreaks = true
				}
			case "||":
				if opType == concatOps {
					cCnt++
				}
			case "+", "-", "*", "/":
				if opType == mathOps {
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

		if !addBreaks {
			continue
		}
		addBreaks = false

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
				case "+", "-", "*", "/":
					addBreak = opType == mathOps
				case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
					addBreak = opType == compareOps
				}
				if addBreak {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(tpi + tpd + 1)
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
								switch tokens[i-1].value {
								case "(", ",":
									tokens[i].EnsureVSpace()
									tokens[i].AdjustIndents(tpi + tpd)
								default:
									if tokens[i].vSpace > 0 {
										tokens[i].AdjustIndents(tpi + tpd)
									}
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

	idxMax := len(tokens) - 1
	//log.Printf("wrapPLxLogical  [%s] [%s]", tokens[0].value, tokens[idxMax].value)

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			indents = tokens[idx].indents
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
				logicalLen := calcSliceLen(e, bagType, tokens[logicalStart:idx])
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
				case 1:
					splitOnLogical = logicalLen > e.MaxLineLength()
				default:
					splitOnLogical = true
				}

				if splitOnLogical {

					pkv = ""
					ipd := logicalIndents + 1

					for i := logicalStart; i <= idx; i++ {

						switch tokens[i].AsUpper() {
						case "(":
							ipd++
						case ")":
							ipd--
						case "AND", "OR":
							if isLogical(pkv, tokens[i]) {
								tokens[i].EnsureVSpace()
								tokens[i].AdjustIndents(ipd)
							}
						default:
							if i > logicalStart {
								if tokens[i].vSpace > 0 {
									tokens[i].AdjustIndents(ipd)
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

	//log.Print("wrapValueTuples")

	if len(tokens) == 0 {
		return tokens
	}

	// If there is only one values tuple then there is nothing to wrap
	// If there are more than one values tuple then each tuple should be on a separate line

	// TODO: proper wrapping of long tuples per wrapOnCommas

	idxStart := 0
	indents := 0
	inValues := false
	parensDepth := 0
	pdl := 0
	ptVal := ""

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
		}

		switch cTok.value {
		case "(":
			parensDepth++

			switch ptVal {
			case "VALUES":
				pdl = parensDepth
				inValues = true
				idxStart = idx

			case ",":
				if parensDepth == pdl && inValues {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents + parensDepth - 1)
					if tokens[idxStart].vSpace == 0 {
						tokens[idxStart].EnsureVSpace()
						tokens[idxStart].AdjustIndents(indents + parensDepth - 1)
					}
				}
			}

		case ")":
			switch {
			case parensDepth == pdl:
				if ptVal == ")" {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents + parensDepth - 1)
				}
			case parensDepth < pdl:
				if ptVal == ")" {
					tokens[idx].EnsureVSpace()
					tokens[idx].AdjustIndents(indents + parensDepth - 1)
				}
				inValues = false
				idxStart = 0
				pdl = 0
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