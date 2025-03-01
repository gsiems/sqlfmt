package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/db-dialect/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

const (
	compareOps = iota + 400
	concatOps
	logicOps
	mathOps
	mathAddSubOps
	mathMultDivOps
	winFuncOps
	wrapVertical
	wrapHorizontal
	wrapHybrid
	wrapMod2
	wrapOther
	wrapNone
)

func wrapsName(i int) string {
	var names = map[int]string{
		compareOps:     "compareOps",
		concatOps:      "concatOps",
		logicOps:       "logicOps",
		mathOps:        "mathOps",
		mathAddSubOps:  "mathAddSubOps",
		mathMultDivOps: "mathMultDivOps",
		winFuncOps:     "winFuncOps",
		wrapVertical:   "wrapVertical",
		wrapHorizontal: "wrapHorizontal",
		wrapHybrid:     "wrapHybrid",
		wrapMod2:       "wrapMod2",
		wrapOther:      "wrapOther",
		wrapNone:       "wrapNone",
	}

	if tName, ok := names[i]; ok {
		return tName
	}
	return ""
}

func calcIndent(bagType int, cTok FmtToken) int {

	indents := cTok.indents

	switch bagType {
	case CommentOnBag:
		switch cTok.AsUpper() {
		case "COMMENT":
			return indents + 1
		}
	case DCLBag:
		switch cTok.AsUpper() {
		case "GRANT", "REVOKE":
			return indents + 1
		}
	case DDLBag:
		switch cTok.AsUpper() {
		case "ALTER":
			return indents + 1
		}
	case DMLBag:
		switch cTok.AsUpper() {
		case "SELECT", "INSERT", "FOR":
			return indents + 2
		case "FROM", "GROUP BY", "WHERE", "HAVING", "WINDOW", "ORDER BY",
			"OFFSET", "LIMIT", "FETCH", "WITH", "VALUES", "RETURNING", "CROSS",
			"FULL", "INNER", "JOIN", "LATERAL", "LEFT", "NATURAL", "OUTER",
			"RIGHT":
			return indents + 1
		}
	case PLxBody:
		switch cTok.AsUpper() {
		case "DECLARE", "BEGIN", "IF", "CASE", "FOR", "LOOP", "WHEN", "THEN",
			"ELSE", "EXCEPTION", "EXECUTE":
			return indents + 1
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

func isOperator(opType int, cTok FmtToken) bool {
	switch cTok.value {
	case "||":
		return opType == 0 || opType == concatOps
	case "+", "-":
		return opType == 0 || opType == mathAddSubOps
	case "*", "/":
		return opType == 0 || opType == mathMultDivOps
	case "=", "==", "<", ">", "<>", "!=", ">=", "<=":
		return opType == 0 || opType == compareOps
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
		if idx > 0 {
			if tokens[idx].vSpace > 0 {
				return sliceLen
			}
			if bagType == CommentOnBag && tokens[idx-1].AsUpper() == "IS" {
				return sliceLen
			}
		}
		sliceLen += calcLen(e, tokens[idx])
	}
	return sliceLen
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
			tokens = wrapOnMod2Commas(e, bagType, "JSONB_BUILD_OBJECT", true, tokens)
		case dialect.Oracle:
			tokens = wrapOnMod2Commas(e, bagType, "DECODE", false, tokens)
		}

		tokens = wrapDMLCase(e, bagType, tokens)
		tokens = wrapDMLLogical(e, bagType, tokens)

	case PLxBody:
		tokens = wrapPLxCalls(e, bagType, maxParensDepth, tokens)
		tokens = wrapPLxCase(e, bagType, tokens)
		tokens = wrapPLxLogical(e, bagType, tokens)
	}
	//tokens = wrapInto(e, bagType, tokens)

	//////////////////////////////////////////////////
	for idx := 0; idx <= idxMax; idx++ {

		eol := false
		switch {
		case idx < idxMax:
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

	// A work in progress...
	// Order matters but may be/is probably context specific...
	// Maybe consider the original vSpace for operators

	//for pdl := 0; pdl <= mxPd; pdl++ {
	//	tokens = wrapOnCommasY(e, bagType, pdl, tokens)
	//}

	for pdl := 0; pdl <= mxPd; pdl++ {

		tokens = wrapOnCommasY(e, bagType, pdl, tokens)

		tokens = wrapOnCommasX(e, bagType, pdl, tokens)
		tokens = wrapOnCompOps(e, bagType, pdl, tokens)
		tokens = wrapOnMathOps(e, bagType, pdl, tokens)
		tokens = wrapOnConcatOps(e, bagType, pdl, tokens)

	}
	//for pdl := 0; pdl <= mxPd; pdl++ {
	//	tokens = wrapOnParens(e, bagType, pdl, tokens)
	//}

	return tokens
}

func addCsvBreaks(e *env.Env, bagType, indents, idxLineStart, idxStart, idxEnd, wMode int, tokens *[]FmtToken) {

	// Assertion: the idxStart token is the opening token (not part of the csv list)
	// and may be an open parens, ORDER BY, GROUP BY, etc.
	// The idxEnd token is *probably* not part of the csv list???

	var ixs []int
	lCnt := 1
	ipd := 0

	segLen := calcSliceLen(e, bagType, (*tokens)[idxLineStart:idxEnd])
	if (*tokens)[idxEnd].vSpace == 0 {
		segLen += calcLenToLineEnd(e, bagType, (*tokens)[idxEnd:])
	}

	for idx := idxStart + 1; idx <= idxEnd; idx++ {
		switch (*tokens)[idx].value {
		case "(":
			ipd++
		case ")":
			ipd--
		}

		// gather the token indexes to potentially wrap on
		if ipd == 0 {
			if idx == idxStart+1 {
				ixs = append(ixs, idx)
			} else {
				switch (*tokens)[idx-1].value {
				case ",":
					ixs = append(ixs, idx)
				}
			}
		}
		if idx < idxEnd && (*tokens)[idx].vSpace > 0 {
			lCnt++
		}
	}

	if len(ixs) == 1 {
		return
	}

	switch wMode {
	case wrapVertical:
		// nada
	case wrapHorizontal:
		if segLen <= e.MaxLineLength() {
			return
		}
	case wrapHybrid:
		switch {
		case len(ixs) > 3:
			wMode = wrapVertical
		case len(ixs) > 1:
			wMode = wrapVertical
		case segLen > e.MaxLineLength():
			wMode = wrapVertical
		default:
			return
		}
	default:
		wMode = wrapHorizontal
	}

	switch wMode {
	case wrapVertical:

	case wrapHorizontal:

		// 1-- wrap after the open parens as needed

		// 2-- wrap just before the length exceeds the max line length
		breakCount := 0
		iMax := len(ixs) - 1
		for idx := 0; idx <= iMax; idx++ {
			ix := ixs[idx]

			segLen = 0
			switch {
			case idxLineStart == ix:
				segLen = calcLen(e, (*tokens)[idxStart])
			case idxLineStart > ix:
				segLen = calcLenToLineEnd(e, bagType, (*tokens)[idxLineStart:])
			default:
				segLen = calcSliceLen(e, bagType, (*tokens)[idxLineStart:ix])
			}

			nLen := 0
			switch {
			case ix < ixs[iMax]:
				nLen = calcSliceLen(e, bagType, (*tokens)[ix:ixs[idx+1]])
			case ix == ixs[iMax]:
				nLen = calcLenToLineEnd(e, bagType, (*tokens)[ix:])
			case ix < idxEnd:
				nLen = calcSliceLen(e, bagType, (*tokens)[ix:idxEnd])
			}

			if segLen+nLen > e.MaxLineLength() {
				(*tokens)[ix].EnsureVSpace()
				(*tokens)[ix].AdjustIndents(indents)
				breakCount++
				idxLineStart = ix
			}
		}

		if breakCount == 0 && false {
			// no breaks added... try scanning to the first comma that
			// can split the line into two portions that are both less
			// than the max line length
			segLen := calcLenToLineEnd(e, bagType, (*tokens)[idxLineStart:])
			pdLen := len(strings.Repeat(e.Indent(), indents))

			for idx := 0; idx <= iMax; idx++ {
				ix := ixs[idx]
				remLen := calcLenToLineEnd(e, bagType, (*tokens)[ix:])

				switch {
				case segLen <= e.MaxLineLength():
					// nada
				case pdLen+remLen > e.MaxLineLength():
					// nada
				case segLen-remLen > e.MaxLineLength():
					// nada
				default:
					(*tokens)[ix].EnsureVSpace()
					(*tokens)[ix].AdjustIndents(indents)
					breakCount++
					idxLineStart = ix
				}
				if breakCount > 0 {
					break
				}
			}
		}

		// 3-- wrap before the close parens as needed
		//segLen = calcLenToLineEnd(e, bagType, (*tokens)[idxLineStart:])
		//switch {
		//case segLen <= e.MaxLineLength():
		//	// nada
		//case bagType == CommentOnBag:
		//	// nada
		//case idxMax-idx <= 3:
		//	// nada
		//default:
		//	if addBreak {
		//		(*tokens)[idxEnd].EnsureVSpace()
		//		(*tokens)[idxEnd].AdjustIndents(indents)
		//		idxLineStart = idxEnd
		//	}
		//}

	}

}

func addInlineCaseBreaks(e *env.Env, bagType, indents, parensDepth, lineLen, idxStart, idxEnd int, tokens *[]FmtToken) {

	caseLen := 0
	caseCnt := 0
	whenCnt := 0
	caseDepth := 0
	caseInd := indents + parensDepth

	for idx := idxStart; idx <= idxEnd; idx++ {
		caseLen += calcLen(e, (*tokens)[idx])
		switch (*tokens)[idx].AsUpper() {
		case "CASE":
			caseCnt++
		case "WHEN", "ELSE":
			whenCnt++
		}
	}

	// Wrap the start of the CASE statement as appropriate
	if lineLen > e.MaxLineLength() {
		switch {
		case (*tokens)[idxStart].vSpace > 0:
			caseInd = (*tokens)[idxStart].indents + 1
		case idxStart == 0:
			// nada
		default:
			switch (*tokens)[idxStart-1].value {
			case "=>":
				if (*tokens)[idxStart-1].vSpace == 0 {
					(*tokens)[idxStart].EnsureVSpace()
					(*tokens)[idxStart].indents = caseInd + 1
					caseInd += 2
					caseLen += len(strings.Repeat(e.Indent(), (*tokens)[idxStart].indents))
				}
			case "(":
				if (*tokens)[idxStart-1].vSpace == 0 {
					(*tokens)[idxStart].EnsureVSpace()
					(*tokens)[idxStart].indents = caseInd
					caseInd++
					caseLen += len(strings.Repeat(e.Indent(), (*tokens)[idxStart].indents))
				}
			case ",":
				if (*tokens)[idxStart-1].vSpace == 0 {
					(*tokens)[idxStart].EnsureVSpace()
					(*tokens)[idxStart].indents = caseInd
					caseInd++
					caseLen += len(strings.Repeat(e.Indent(), (*tokens)[idxStart].indents))
				}
			default:
				if (*tokens)[idxStart-1].typeOf == parser.Operator {
					if (*tokens)[idxStart-1].vSpace == 0 {
						(*tokens)[idxStart-1].EnsureVSpace()
						(*tokens)[idxStart-1].indents = caseInd + 1
						caseInd += 2
						caseLen += len(strings.Repeat(e.Indent(), (*tokens)[idxStart-1].indents))
					}
				} else {
					caseInd++
				}
			}
		}
	} else {
		caseInd++
	}

	// Wrap the end of the CASE statement as needed
	// ... first determine the length of the line after the statement
	postCaseLineLen := 0
	for idx := idxEnd + 1; idx < len(*tokens); idx++ {
		if (*tokens)[idx].vSpace > 0 {
			break
		}
		postCaseLineLen += calcLen(e, (*tokens)[idx])
	}

	// ... then wrap as needed
	if caseLen+postCaseLineLen > e.MaxLineLength() {
		switch {
		case len(*tokens)-1 == idxEnd:
		// nada
		case (*tokens)[idxEnd+1].typeOf == parser.Operator:
			if (*tokens)[idxEnd+1].vSpace == 0 {
				(*tokens)[idxEnd+1].EnsureVSpace()
				(*tokens)[idxEnd+1].indents = indents + parensDepth + 1
				postCaseLineLen = 0
			}
		default:
			switch (*tokens)[idxEnd+1].AsUpper() {
			case ",":
				if len(*tokens) > idxEnd+2 {
					if (*tokens)[idxEnd+2].vSpace == 0 {
						(*tokens)[idxEnd+2].EnsureVSpace()
						(*tokens)[idxEnd+2].indents = indents + parensDepth
						postCaseLineLen = 1
					}
				}
			case ")":
				if (*tokens)[idxEnd+1].vSpace == 0 {
					(*tokens)[idxEnd+1].EnsureVSpace()
					(*tokens)[idxEnd+1].indents = indents + parensDepth
					postCaseLineLen = 0
				}
			}
		}
	}

	// Check if any further wrapping is needed
	if caseLen+postCaseLineLen < e.MaxLineLength() && caseCnt == 1 && whenCnt < 3 {
		return
	}

	// Determine how long the END... is
	endLen := 0
	for idx := idxEnd; idx < len(*tokens); idx++ {
		if (*tokens)[idx].vSpace > 0 {
			break
		}
		if (*tokens)[idx].value == "," {
			endLen += calcLen(e, (*tokens)[idx])
			break
		}
		if (*tokens)[idx].value == ")" {
			break
		}
	}

	// Add wrapping of WHEN, END, and possibly THEN statements
	if (*tokens)[idxEnd].vSpace == 0 {
		(*tokens)[idxEnd].EnsureVSpace()
		(*tokens)[idxEnd].indents = caseInd
	}
	caseDepth = 0

	for idx := idxStart + 1; idx < idxEnd; idx++ {

		switch (*tokens)[idx].AsUpper() {
		case "CASE":
			caseDepth++
		case "END":
			caseDepth--
		case "WHEN", "ELSE":
			if caseDepth == 0 {
				if (*tokens)[idx].vSpace == 0 {
					(*tokens)[idx].EnsureVSpace()
				}
				(*tokens)[idx].indents = caseInd
				// determine length to end of the THEN statement
				idxThen := 0
				icd := 0
				whenLen := calcLen(e, (*tokens)[idx])
				whenDone := false
				for j := idx + 1; j <= idxEnd; j++ {
					if whenDone {
						break
					}

					ctVal := (*tokens)[j].AsUpper()

					switch ctVal {
					case "CASE":
						icd++
					}

					if icd == 0 {

						switch ctVal {
						case "WHEN", "ELSE", "END":
							if (*tokens)[j].vSpace > 0 {
								(*tokens)[j].indents = caseInd
							}
						case "THEN":
							if (*tokens)[j].vSpace > 0 {
								(*tokens)[j].indents = caseInd + 1
							}
						}

						switch ctVal {
						case "THEN":
							idxThen = j
							if (*tokens)[idxThen].vSpace > 0 {
								(*tokens)[idxThen].indents = caseInd + 1
							}
						case "WHEN", "ELSE":
							if idxThen > 0 {
								if whenLen > e.MaxLineLength() {
									if (*tokens)[idxThen].vSpace == 0 {
										(*tokens)[idxThen].EnsureVSpace()
									}
									(*tokens)[idxThen].indents = caseInd + 1
								}
								whenDone = true
								whenLen = 0
							}
						case "END":
							if idxThen > 0 {
								if whenLen+endLen > e.MaxLineLength() {
									if (*tokens)[idxThen].vSpace == 0 {
										(*tokens)[idxThen].EnsureVSpace()
									}
									(*tokens)[idxThen].indents = caseInd + 1
								}
								whenDone = true
							}
						}
					}

					whenLen += calcLen(e, (*tokens)[j])

					switch ctVal {
					case "END":
						icd--
					}
				}
			}
		}
	}
}

func wrapDMLCase(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

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
		indents := 0
		ipd := 0
		lineLen := 0

		for idx := 0; idx <= idxMax; idx++ {
			if idx == 0 || caseDepth < cdl {
				if tokens[idx].vSpace > 0 {
					lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
					indents = calcIndent(bagType, tokens[idx])
					//switch tokens[idx].AsUpper() {
					//case "GROUP BY", "ORDER BY":
					////case "SELECT", "GROUP BY", "ORDER BY":
					//	indents++
					//}
					ipd = 0
				}
			}

			switch tokens[idx].AsUpper() {
			case "(":
				ipd++
			case ")":
				ipd--
			case "CASE":
				caseDepth++
				if caseDepth == cdl {
					idxStart = idx
				}
			case "END":
				if caseDepth == cdl {
					addInlineCaseBreaks(e, bagType, indents, ipd, lineLen, idxStart, idx, &tokens)
					ipd = 0
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
	pKwVal := ""
	idxStart := 0
	lCnt := 0
	lineLen := 0
	indents := 0
	isOn := false

	for idx := 0; idx <= idxMax; idx++ {

		if tokens[idx].vSpace > 0 {
			lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
			indents = calcIndent(bagType, tokens[idx])
			isOn = tokens[idx].AsUpper() == "ON"
		}

		switch {
		case isLogical(pKwVal, tokens[idx]):
			lCnt++
		}

		doCheck := false
		switch {
		case idx == idxMax:
			doCheck = true
		case tokens[idx+1].vSpace > 0:
			doCheck = true
		}

		if doCheck {
			addBreaks := false
			if lCnt > 0 {
				switch {
				case lineLen > e.MaxLineLength():
					addBreaks = true
				case lCnt > 2:
					addBreaks = true
				}
			}

			if addBreaks {
				ipd := 0
				pkv := ""
				for i := idxStart; i < idx; i++ {
					switch tokens[i].value {
					case "(":
						ipd++
					case ")":
						ipd--
					default:
						if isLogical(pkv, tokens[i]) {
							if tokens[i].vSpace == 0 {
								tokens[i].EnsureVSpace()
								if isOn {
									tokens[i].indents = indents + ipd
								} else {
									tokens[i].indents = indents + ipd + 1
								}
							}
						}
					}

					if tokens[i].IsKeyword() {
						pkv = tokens[i].AsUpper()
					}
				}
			}

			lCnt = 0
			idxStart = idx
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
		idxBase := 0
		indents := 0
		lCnt := 0
		parensDepth := 0
		var idxs []int

		if tokens[0].vSpace > 0 {
			indents = calcIndent(bagType, tokens[0])
		}

		for idx := 0; idx <= idxMax; idx++ {

			if tokens[idx].value == "(" {
				parensDepth++
				if parensDepth == pdl {
					lCnt = 0
					idxs = nil
					idxs = append(idxs, idx)
				}
			}

			doWrap := false
			switch {
			case parensDepth < pdl:
				if tokens[idx].vSpace > 0 {
					idxBase = idx
				}
			case parensDepth == pdl:
				if tokens[idx].vSpace > 0 {
					lCnt++
				}

				switch tokens[idx].value {
				case ")":
					if len(idxs) > 1 {
						segLen := calcSliceLen(e, bagType, tokens[idxBase:idx])
						switch {
						case idx == idxMax:
							// nada... at the end already
						case tokens[idx+1].value == ",":
							segLen++
						default:
							segLen += calcLenToLineEnd(e, bagType, tokens[idx:])
						}

						switch {
						case lCnt > 0:
							doWrap = true
						case segLen > e.MaxLineLength():
							doWrap = true
						}
					}
					idxs = append(idxs, idx)
				case "ORDER BY", "GROUP BY", "PARTITION BY":
					idxs = append(idxs, idx)
				}
			}

			if doWrap {
				tpd := pdl
				for i := idxs[0] + 1; i < idx; i++ {
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
						default:
							if tokens[i].vSpace > 0 {
								tokens[i].AdjustIndents(tokens[i].indents + 1)
							}
						}
					}
				}

				ixst := 0
				ixnd := 0
				for i, j := range idxs {
					if i == 0 {
						ixnd = j
					} else {
						ixst = ixnd
						ixnd = j
						addCsvBreaks(e, bagType, tokens[ixst].indents+1, ixst, ixst, ixnd, wrapHorizontal, &tokens)
					}
				}
			}

			if tokens[idx].value == ")" {
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
	idxLineStart := 0
	ppKwVal := ""
	pKwVal := ""
	wMode := env.WrapNone
	var intoIdxs []int

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		if cTok.vSpace > 0 {
			indents = calcIndent(bagType, cTok)
			if idxStart == 0 {
				idxLineStart = idx
			}
		}

		switch cTok.value {
		case "(":
			parensDepth++
			if ppKwVal == "INSERT" && pKwVal == "INTO" {
				intoParensDepth = parensDepth
				inInto = true
				intoIdxs = nil
				idxStart = idx
				eCnt := 1
				ipd := 0
				idxEnd := 0
				// determine how many elements are in the tuple(s)
				for j := idx; j <= idxMax; j++ {
					done := false
					switch tokens[j].AsUpper() {
					case "(":
						ipd++
					case ")":
						ipd--
						done = ipd < 0
					case ",":
						if ipd == 1 {
							eCnt++
						}
					}

					if done {
						break
						idxEnd = j
					}
				}

				if eCnt > 3 {
					wMode = env.WrapAll
				} else {
					if idxEnd > idxLineStart {
						lineLength := calcSliceLen(e, bagType, tokens[idxLineStart:idxEnd])
						if lineLength > e.MaxLineLength() {
							wMode = env.WrapAll
						}
					}
				}
			}

		case ",":
			if inInto && intoParensDepth == parensDepth {
				if idx < idxMax {
					intoIdxs = append(intoIdxs, idx+1)
				}
			}

		case ")":
			if intoParensDepth == parensDepth {
				if len(intoIdxs) > 0 && wMode == env.WrapAll {
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

// isNotCsvWrappable checks a token value to determine if it should prevent
// wrapping via wrapOnCommasX or wrapOnCommasY
func isCsvWrappable(e *env.Env, v string) bool {
	switch csvWrapMode(e, v) {
	case wrapMod2, wrapNone:
		return false
	}
	return true
}

// isWrapPrefixKw determines if the supplied token is a valid "prefix keyword"
// for the purposes of wrapping lists of elements via wrapOnCommasX or
// wrapOnCommasY
func isWrapPrefixKw(e *env.Env, pdl, parensDepth int, t FmtToken) bool {
	if (pdl == 0 && parensDepth == 0) || parensDepth < pdl {
		switch {
		case t.IsKeyword():
			return true
		case t.value == ":=":
			return true
		case e.Dialect() == dialect.PostgreSQL:
			switch t.AsUpper() {
			case "CALL", "PERFORM", "FORMAT", "RETURNING", "ROW", "AS",
				"CONCAT_WS", "CONCAT": //, "JSON_POPULATE_RECORDSET",
				//"JSONB_POPULATE_RECORDSET":
				return true
			}
		}
	}
	return false
}

// csvWrapMode determines what mode of wrapping to be implemented by
// wrapOnCommasX and wrapOnCommasY (defaults to horizontal (x))
func csvWrapMode(e *env.Env, v string) int {

	switch strings.ToUpper(v) {
	case "AS", "FOR", "INSERT", "INTO", "ON CONFLICT", "RETURNING", "ROW":
		return wrapHybrid
	case "RAISE":
		return wrapHybrid
	//////////////////////////////////////////////
	case "DECODE":
		return wrapMod2
	case "JSON_BUILD_OBJECT", "JSONB_BUILD_OBJECT":
		return wrapMod2
	//////////////////////////////////////////////
	case "ALTER", "GRANT", "REVOKE", "IN":
		return wrapHorizontal
	//////////////////////////////////////////////
	case "FORMAT", "CONCAT", "CONCAT_WS", "COALESCE":
		//return wrapVertical
		return wrapHybrid
	case "CALL", "PERFORM", "TYPE":
		return wrapVertical
	case "JSONB_POPULATE_RECORDSET", "JSON_POPULATE_RECORDSET":
		return wrapVertical
	//////////////////////////////////////////////
	case "=>":
		// s.b. wrapped by wrapPLxCalls
		return wrapNone
	case "GROUP BY", "ORDER BY", "PARTITION BY":
		// s.b. wrapped by wrapDMLWindowFunctions
		return wrapNone
	case "VALUES":
		// s.b. wrapped by wrapValueTuples
		return wrapNone
	}
	// The default is basically wrapHybrid, however we want/need to be able
	// to distinguish between explicit wrapHybrid and default wrapHybrid
	return wrapOther
}

func wrapOnCommasX(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	cCnt := 0
	idxMax := len(tokens) - 1
	idxLineStart := 0
	idxStart := 0
	indents := 0
	parensDepth := 0
	pKwVal := ""
	wrapMode := wrapHorizontal
	debug := false

	carp(debug, fmt.Sprintf("wrapOnCommasX    pdl: %d", pdl))

	for idx := 0; idx <= idxMax; idx++ {

		if (pdl == 0 && parensDepth == 0) || parensDepth < pdl {
			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				idxLineStart = idx
				idxStart = idx
			}
		}

		carp(debug, fmt.Sprintf("    %d - %s", idx, tokens[idx].String()))

		switch tokens[idx].AsUpper() {
		case "(":
			parensDepth++
			if parensDepth == pdl {
				wrapMode = csvWrapMode(e, pKwVal)
				switch bagType {
				case DDLBag:
					if tokens[0].AsUpper() == "ALTER" {
						wrapMode = wrapHorizontal
					}
				case DCLBag, CommentOnBag:
					wrapMode = wrapHorizontal
				}
				idxStart = idx
				cCnt = 0
			}
		}

		if parensDepth == pdl {
			doCheck := false
			switch tokens[idx].AsUpper() {
			case ")", ";", "IN":
				doCheck = cCnt > 1
			case ",":
				cCnt++

				if cCnt == 1 {
					switch pKwVal {
					case "AS":
						indents++
					}
				}

			default:
				switch {
				//case cCnt > 1:
				//	doCheck = true
				case idx == idxMax:
					doCheck = true
				default:
					if tokens[idx+1].vSpace > 0 {
						doCheck = true
					}
				}

				if wrapMode != wrapNone && !isCsvWrappable(e, tokens[idx].value) {
					wrapMode = wrapNone
				}
				switch {
				case idx == idxMax:
					doCheck = true
				case tokens[idx+1].vSpace > 0:
					doCheck = true
				}
			}

			doWrap := false
			if doCheck {
				segLen := calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
				switch wrapMode {
				case wrapHorizontal:
					doWrap = cCnt > 0 && segLen > e.MaxLineLength()
				}

				carp(debug, fmt.Sprintf("    doCheck idxStart: %d, idx: %d, cCnt: %d, segLen: %d, pKwVal: %s, wrapMode: %s:, doWrap: %t",
					idxStart, idx, cCnt, segLen, pKwVal, wrapsName(wrapMode), doWrap))

			}

			if doWrap {
				ipd := 0
				var ixs []int

				// gather the token indexes to potentially wrap on
				for i := idxStart + 1; i <= idx; i++ {
					switch tokens[i].value {
					case "(":
						ipd++
					case ")":
						ipd--
					}

					if ipd == 0 {
						if i == idxStart+1 {
							ixs = append(ixs, i)
						} else {
							switch tokens[i-1].value {
							case ",":
								ixs = append(ixs, i)
							}
						}
					}
				}
				// wrap
				addBreak := false
				segLen := 0

				// 1-- wrap after the open parens as needed
				switch {
				//				case bagType == CommentOnBag :
				//					segLen = calcSliceLen(e, bagType, tokens[idxLineStart:idx])
				//if segLen > e.MaxLineLength() {
				//	addBreak = true
				//}
				case idxLineStart == idxStart:
					segLen = calcLen(e, tokens[idxStart])
				case idxLineStart > idxStart:
					segLen = calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
				default:
					segLen = calcSliceLen(e, bagType, tokens[idxLineStart:idxStart])
				}

				if segLen > e.MaxLineLength() {
					addBreak = true
				} else {
					switch pKwVal {
					case "INTO", "INSERT", "IN":
						addBreak = true
					case "CALL", "PERFORM", "ROW":
						if e.Dialect() == dialect.PostgreSQL {
							addBreak = true
						}
					}
				}

				carp(debug, fmt.Sprintf("        1) - segLen: %d, addBreak: %t - %s",
					segLen, addBreak, tokens[idx].String()))

				//if addBreak {
				//	tokens[idxStart].EnsureVSpace()
				//	tokens[idxStart].AdjustIndents(indents + pdl)
				//}

				// 2-- wrap just before the length exceeds the max line length
				lineLen := calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
				if lineLen > e.MaxLineLength() {
					breakCount := 0

					iMax := len(ixs) - 1
					for i := 0; i <= iMax; i++ {
						ix := ixs[i]
						nLen := 0
						segLen = 0
						switch {
						case idxLineStart == ix:
							segLen = calcLen(e, tokens[idxStart])
						case idxLineStart > ix:
							segLen = calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
						default:
							segLen = calcSliceLen(e, bagType, tokens[idxLineStart:ix])
						}

						addBreak = false

						carp(debug, fmt.Sprintf("        2a) - i: %d, ix: %d, segLen: %d - %s",
							i, ix, segLen, tokens[ix].String()))

						if segLen > e.MaxLineLength() {
							addBreak = true
						} else {
							ixN := 0
							if i < iMax {
								ixN = ixs[i+1]
								nLen = calcSliceLen(e, bagType, tokens[ix:ixN])
							} else {
								nLen = calcSliceLen(e, bagType, tokens[ix:idx])
							}

							if segLen+nLen > e.MaxLineLength() {
								addBreak = true
							}
						}

						carp(debug, fmt.Sprintf("        2a) - i: %d, ix: %d, segLen: %d, nLen: %d, addBreak: %t",
							i, ix, segLen, nLen, addBreak))

						if addBreak {
							breakCount++
							tokens[ix].EnsureVSpace()
							tokens[ix].AdjustIndents(indents + pdl)
							idxLineStart = ix
						}
					}

					if breakCount == 0 {
						// no breaks added... try scanning to the first comma that
						// can split the line into two portions that are both less
						// than the max line length
						lineLen := calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
						pdLen := len(strings.Repeat(e.Indent(), indents+pdl))

						carp(debug, fmt.Sprintf("        2b) - lineLen: %d, pdLen: %d",
							lineLen, pdLen))

						for i := 0; i <= iMax; i++ {
							if breakCount > 0 {
								break
							}

							ix := ixs[i]
							remLen := calcLenToLineEnd(e, bagType, tokens[ix:])

							carp(debug, fmt.Sprintf("        2b) - i: %d, ix: %d, remLen: %d",
								i, ix, remLen))

							switch {
							//case bagType == CommentOnBag :
							// nada
							case lineLen <= e.MaxLineLength():
								// nada
							case pdLen+remLen > e.MaxLineLength():
								// nada
							case lineLen-remLen > e.MaxLineLength():
								// nada
							default:
								tokens[ix].EnsureVSpace()
								tokens[ix].AdjustIndents(indents + pdl)
								idxLineStart = ix
								breakCount++
							}
						}
					}
				}

				// 3-- wrap before the close parens as needed
				switch bagType {
				case DDLBag, DCLBag, CommentOnBag:
					addBreak = false
					segLen = calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
					if segLen > e.MaxLineLength() {
						switch {
						case bagType == CommentOnBag:
							// nada
						case idxMax-idx <= 3:
							// nada
						default:
							addBreak = true
						}
					}

					carp(debug, fmt.Sprintf("        3) - segLen: %d, addBreak: %t",
						segLen, addBreak))

					if addBreak {
						tokens[idx].EnsureVSpace()
						tokens[idx].AdjustIndents(indents + pdl)
						idxLineStart = idx
					}
				}
			}
		}

		switch tokens[idx].value {
		case ")":
			if parensDepth == pdl {
				cCnt = 0
			}
			parensDepth--
		case ";", "IN":
			cCnt = 0
		}

		if isWrapPrefixKw(e, pdl, parensDepth, tokens[idx]) {
			pKwVal = tokens[idx].AsUpper()
		}

	}
	return tokens
}

func wrapOnCommasY(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	cCnt := 0
	idxLineStart := 0
	idxMax := len(tokens) - 1
	idxStart := 0
	pdi := 0
	indents := 0
	parensDepth := 0
	var pTok FmtToken

	wrapMode := csvWrapMode(e, tokens[0].AsUpper())

	for idx := 0; idx <= idxMax; idx++ {
		cTok := tokens[idx]
		ctVal := cTok.AsUpper()

		if (pdl == 0 && parensDepth == 0) || parensDepth < pdl {
			if cTok.vSpace > 0 {

				idxLineStart = idx
				idxStart = idx
				pdi = 0

				wm := csvWrapMode(e, ctVal)
				switch wm {
				case wrapOther:
					// nada
				default:
					wrapMode = wm
				}

				indents = calcIndent(bagType, cTok)
				switch ctVal {
				case "INSERT":
					indents--
				case "SET":
					indents++
				}
			} else if parensDepth < pdl {
				wm := csvWrapMode(e, ctVal)
				switch wm {
				case wrapOther:
					// nada
				default:
					wrapMode = wm
					switch ctVal {
					case "AS":
						indents++
					}
				}
			}
		}

		switch ctVal {
		case "(":
			parensDepth++
			pdi++
			if parensDepth == pdl {
				idxStart = idx
				cCnt = 0

				wm := csvWrapMode(e, pTok.AsUpper())
				switch wm {
				case wrapOther:
					// nada
				default:
					wrapMode = wm
				}
			}
		}

		doWrap := false
		if parensDepth == pdl {
			doCheck := false
			switch ctVal {
			case ",":
				cCnt++
			case ")", ";", "IN":
				doCheck = cCnt > 0
			default:
				if idx == idxMax {
					doCheck = cCnt > 0
				}
			}

			if doCheck {
				segLen := calcSliceLen(e, bagType, tokens[idxLineStart:idx]) + calcLenToLineEnd(e, bagType, tokens[idx:])

				switch wrapMode {
				case wrapVertical, wrapHybrid, wrapOther:
					for i := idxStart; i <= idx; i++ {
						switch {
						case tokens[i].HasLeadingComments():
							doWrap = i > idxStart
						case tokens[i].HasTrailingComments():
							doWrap = i < idx
						case tokens[i].vSpace > 0:
							doWrap = true
						}
					}
					if doWrap {
						wrapMode = wrapVertical
					}
				}
				if !doWrap {
					switch wrapMode {
					case wrapVertical:
						doWrap = cCnt > 0 && segLen > e.MaxLineLength()
					case wrapHybrid, wrapOther:
						doWrap = cCnt > 2 || segLen > e.MaxLineLength()
					}
				}

				if doWrap {
					for i := idxStart; i <= idx; i++ {
						switch tokens[i].AsUpper() {
						case "GROUP BY", "ORDER BY", "PARTITION BY":
							doWrap = false
						}
					}
				}
			}
		}

		if doWrap {
			ipd := 0
			for i := idxStart + 1; i <= idx; i++ {
				switch tokens[i].value {
				case "(":
					ipd++
				case ")":
					ipd--
				}

				if ipd == 0 {
					switch tokens[i-1].value {
					case "(", ",":
						if tokens[i].value != ")" {
							if tokens[i].vSpace == 0 {
								tokens[i].EnsureVSpace()
								tokens[i].AdjustIndents(indents + max(pdi, 1))
							}
						}
					}
				}
			}
		}

		switch ctVal {
		case ")":
			if parensDepth == pdl {
				cCnt = 0
			}
			parensDepth--
			pdi--
		case ";", "IN":
			cCnt = 0
		}
		pTok = cTok
	}

	return tokens
}

func wrapOnCompOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	return wrapOnOps(e, bagType, compareOps, pdl, tokens)
}

func wrapOnConcatOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	return wrapOnOps(e, bagType, concatOps, pdl, tokens)
}

func wrapOnMathOps(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {
	tokens = wrapOnOps(e, bagType, mathAddSubOps, pdl, tokens)
	return wrapOnOps(e, bagType, mathMultDivOps, pdl, tokens)
}

func wrapOnOps(e *env.Env, bagType, opType, pdl int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	idxMax := len(tokens) - 1
	indents := 0
	ipd := 0
	parensDepth := 0
	pKwVal := ""
	debug := false
	isDirty := false

	carp(debug, fmt.Sprintf("wrapOnParens    bagType: %s, opType: %s, pdl: %d",
		nameOf(bagType), wrapsName(opType), pdl))

	lineLen := calcLenToLineEnd(e, bagType, tokens)
	if tokens[0].vSpace > 0 {
		indents = calcIndent(bagType, tokens[0]) + 1
	}

	for idx := 0; idx <= idxMax; idx++ {
		if idx > 0 && tokens[idx].vSpace > 0 {
			lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])

			switch opType {
			case mathAddSubOps, mathMultDivOps:
				if !isOperator(mathAddSubOps, tokens[idx]) && !isOperator(mathMultDivOps, tokens[idx]) {
					indents = tokens[idx].indents + 1
				}
			default:
				if !isOperator(opType, tokens[idx]) {
					indents = tokens[idx].indents + 1
				}
			}
			ipd = 0

		}

		switch tokens[idx].value {
		case "(":
			parensDepth++
			ipd++
		case ")":
			parensDepth--
			ipd--
		}

		switch {
		case parensDepth != pdl:
			// nada
		case lineLen <= e.MaxLineLength():
			// nada
		case isOperator(opType, tokens[idx]):

			addBreak := true
			switch {
			case pKwVal == "SET":
				if tokens[idx].value == "=" {
					if tokens[idx-1].value != ")" {
						addBreak = false
					}
				}
			case opType == concatOps:
				if idx < idxMax {
					switch tokens[idx+1].value {
					case "','", "', '":
						addBreak = false
					}
				}
			}

			if addBreak {
				switch {
				case tokens[idx].vSpace == 0:
					isDirty = true
				case tokens[idx].indents != indents+ipd:
					isDirty = true
				}

				tokens[idx].EnsureVSpace()
				tokens[idx].AdjustIndents(indents + ipd)
			}

		}

		if isWrapPrefixKw(e, pdl, parensDepth, tokens[idx]) {
			pKwVal = tokens[idx].AsUpper()
		}
	}

	if isDirty {
		parensDepth = 0

		for idx := 0; idx <= idxMax; idx++ {
			switch tokens[idx].value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}

			switch {
			case isOperator(0, tokens[idx]):
				idt := indents + parensDepth
				if idx > 0 && tokens[idx].vSpace > 0 && tokens[idx].indents != idt {
					tokens[idx].AdjustIndents(idt)
				}
			case tokens[idx].vSpace > 0:
				indents = max(calcIndent(bagType, tokens[idx]), tokens[idx].indents+1)
				switch tokens[idx].AsUpper() {
				case "SELECT", "SET":
					indents++
				}
				parensDepth = 0
			}
		}
	}
	return tokens
}

func wrapOnParens(e *env.Env, bagType, pdl int, tokens []FmtToken) []FmtToken {

	// wrap on open parens
	// for DML and PL only?
	// for each line (vspace)
	// scan the line until the 1/3 of maxLineLength point??? or not
	// if the remainder of the line is < maxLineLength then wrap on the parens
	// ... provided that the wrapped line is still < maxLineLength when considering the indents on the new line

	//   select coalesce ( func_01 ( ... ), func_02 ( ... ) ) ;
	//   vs.
	//   select coalesce ( func_01 ( func_02 ( func_03 ( ... ) ) ) ) ;

	switch bagType {
	case DMLBag, PLxBody:
		// nada
	default:
		return tokens
	}
	if len(tokens) == 0 {
		return tokens
	}
	if pdl < 1 {
		return tokens
	}
	if pdl > 3 {
		return tokens
	}

	idxMax := len(tokens) - 1
	idxLineStart := 0
	indents := 0
	parensDepth := 0
	pKwVal := ""
	wrapMode := wrapNone
	debug := false

	carp(debug, fmt.Sprintf("wrapOnParens    pdl: %d", pdl))

	for idx := 0; idx <= idxMax; idx++ {

		if !isCsvWrappable(e, tokens[idx].value) {
			return tokens
		}

		if (pdl == 0 && parensDepth == 0) || parensDepth < pdl {
			if tokens[idx].vSpace > 0 {
				indents = calcIndent(bagType, tokens[idx])
				idxLineStart = idx
			}
		}

		carp(debug, fmt.Sprintf("    %d - %s", idx, tokens[idx].String()))

		switch tokens[idx].value {
		case "(":
			parensDepth++
		case ")":
			parensDepth--
		}

		doCheck := false
		if parensDepth == pdl {
			if idx > 0 && tokens[idx-1].value == "(" {
				wrapMode = csvWrapMode(e, pKwVal)
				doCheck = wrapMode == wrapHorizontal
			}
		}

		addBreak := false
		if doCheck {
			lineLen := calcLenToLineEnd(e, bagType, tokens[idxLineStart:])
			remLen := calcLenToLineEnd(e, bagType, tokens[idx:])
			pdLen := len(strings.Repeat(e.Indent(), indents+pdl))

			switch {
			//case oCnt > 0:
			//	// nada
			case lineLen <= e.MaxLineLength():
				// nada
			case pdLen+remLen > e.MaxLineLength():
				// nada
			case lineLen-remLen > e.MaxLineLength():
				// nada
			default:
				idp := pdl
				for i := idx + 1; i < idxMax; i++ {
					if addBreak {
						break
					}

					if tokens[i].vSpace > 0 {
						addBreak = true
					}
					switch tokens[i].value {
					case "(":
						idp++
					case ")":
						addBreak = idp == pdl
						idp--
					}

					if !isCsvWrappable(e, tokens[i].value) {
						break
					}
				}
			}
		}

		if addBreak {
			tokens[idx].EnsureVSpace()
			tokens[idx].AdjustIndents(indents + pdl)
			idxLineStart = idx
		}

		if isWrapPrefixKw(e, pdl, parensDepth, tokens[idx]) {
			pKwVal = tokens[idx].AsUpper()
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
								case "(":
									if i == idxStart+1 {
										tokens[i].EnsureVSpace()
										tokens[i].AdjustIndents(tpi + tpd)
									}
								case ",":
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

func wrapPLxCase(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	if len(tokens) == 0 {
		return tokens
	}

	// The difference between in-lined case structures and "regular" case
	// structures is that the in-lined ones will not have any semi-colons
	// between the "CASE" and the "END" tokens

	// For in-lined use the same code as for DML casing, otherwise wrap similar
	// to IF ... THEN ... END IF structures.

	caseDepth := 0
	cdMax := 0
	idxMax := len(tokens) - 1

	// determine the max case depth
	for idx := 0; idx <= idxMax; idx++ {
		switch tokens[idx].AsUpper() {
		case "CASE":
			caseDepth++
			cdMax = max(cdMax, caseDepth)
		case "END":
			// in-line case statement
			caseDepth--
		case "CASE END":
			// case structure
			caseDepth--
		}
	}

	if cdMax == 0 {
		return tokens
	}

	for cdl := 1; cdl <= cdMax; cdl++ {

		caseDepth := 0
		caseIdxs := make(map[int]string)
		idxEnd := 0
		idxStart := 0
		ifCnt := 0
		indents := 0
		ipd := 0
		lineLen := 0
		scCnt := 0

		for idx := 0; idx <= idxMax; idx++ {

			if caseDepth < cdl {
				if tokens[idx].vSpace > 0 {
					if tokens[idx].AsUpper() == "CASE" {
						indents = tokens[idx].indents
					} else {
						indents = calcIndent(bagType, tokens[idx])
					}
					lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
					ipd = 0
				}
			}

			cTok := tokens[idx]
			ctVal := cTok.AsUpper()

			switch ctVal {
			case "(":
				ipd++
			case ")":
				ipd--
			case "CASE":
				caseDepth++
				if caseDepth == cdl {
					caseIdxs[idx] = ctVal
					idxEnd = 0
					idxStart = idx
					ifCnt = 0
					scCnt = 0
				}
			}

			if caseDepth == cdl {
				switch ctVal {
				case "IF":
					ifCnt++
				case "END IF":
					ifCnt--
				}

				if ifCnt == 0 {
					switch ctVal {
					case "IF":
						ifCnt++
					case "END IF":
						ifCnt--
					case ";":
						scCnt++
					case "WHEN", "THEN", "ELSE":
						caseIdxs[idx] = ctVal
					}
				}
			}

			doCheck := false
			switch ctVal {
			case "END", "END CASE":
				if caseDepth == cdl {
					caseIdxs[idx] = ctVal
					doCheck = true
					idxEnd = idx
				}
				caseDepth--
			}

			if !doCheck {
				continue
			}

			var ary []string
			for i := idxStart; i <= idxEnd; i++ {
				if _, ok := caseIdxs[i]; ok {
					ary = append(ary, fmt.Sprintf("%d", tokens[i].id))
				}
			}

			if scCnt == 0 {
				////////////////////////////////////////////////////////
				// in-line case statement
				addInlineCaseBreaks(e, bagType, indents, ipd, lineLen, idxStart, idx, &tokens)
				idxStart = 0
				idxEnd = 0

			} else {
				////////////////////////////////////////////////////////
				// case block
				tokens[idxStart].EnsureVSpace()
				tokens[idxStart].AdjustIndents(indents)
				tokens[idxEnd].EnsureVSpace()
				tokens[idxEnd].AdjustIndents(indents)

				for i := idxStart + 1; i < idxEnd; i++ {
					switch tokens[i].AsUpper() {
					case "WHEN", "ELSE":
						if _, ok := caseIdxs[i]; ok {
							tokens[i].EnsureVSpace()
							tokens[i].AdjustIndents(indents + 1)
						}
					default:
						switch tokens[i-1].AsUpper() {
						case "THEN", "ELSE":
							if _, ok := caseIdxs[i-1]; ok {
								tokens[i].EnsureVSpace()
								tokens[i].AdjustIndents(indents + 2)
							}
						default:
							if tokens[i].vSpace > 0 {
								tokens[i].AdjustIndents(tokens[i].indents + 2)
							}
						}
					}
				}
				idxStart = 0
				idxEnd = 0
			}
		}
	}
	return tokens
}

func addPLxLogicalBreaks(e *env.Env, bagType, indents, lineLen, idxStart, idxEnd int, tokens *[]FmtToken) {

	ipd := 0
	lCnt := 0
	pKwVal := ""

	for idx := idxStart; idx <= idxEnd; idx++ {
		switch (*tokens)[idx].AsUpper() {
		case "AND", "OR":
			lCnt = adjLogicalCnt(lCnt, pKwVal, (*tokens)[idx])
		}
		if (*tokens)[idx].IsKeyword() {
			pKwVal = (*tokens)[idx].AsUpper()
		}
	}

	switch lCnt {
	case 0:
		return
	case 1, 2, 3:
		if lineLen <= e.MaxLineLength() {
			return
		}
	}

	pKwVal = ""
	for idx := idxStart; idx <= idxEnd; idx++ {

		switch (*tokens)[idx].AsUpper() {
		case "(":
			ipd++
		case ")":
			ipd--
			if (*tokens)[idx].vSpace > 0 {
				(*tokens)[idx].AdjustIndents(ipd - 1)
			}
		case "AND", "OR":
			if isLogical(pKwVal, (*tokens)[idx]) {
				(*tokens)[idx].EnsureVSpace()
				(*tokens)[idx].AdjustIndents(indents + ipd + 1)
			}
		default:
			if idx > idxStart {
				if (*tokens)[idx].vSpace > 0 {
					(*tokens)[idx].AdjustIndents(ipd + (*tokens)[idx].indents)
				}
			}
		}

		if (*tokens)[idx].IsKeyword() {
			pKwVal = (*tokens)[idx].AsUpper()
		}
	}
}

func wrapPLxLogical(e *env.Env, bagType int, tokens []FmtToken) []FmtToken {

	indents := 0
	inLogical := false
	logicalStart := 0
	lineLen := 0

	for _, st := range []string{"BLK", "ASN"} {
		for idx, cTok := range tokens {

			if !inLogical {
				if cTok.vSpace > 0 {
					indents = cTok.indents
					lineLen = calcLenToLineEnd(e, bagType, tokens[idx:])
				}
			}

			switch st {
			case "BLK":
				switch cTok.AsUpper() {
				case "IF", "ELSIF", "ELSEIF", "WHEN":
					inLogical = true
					logicalStart = idx
				case "THEN":
					if inLogical {
						addPLxLogicalBreaks(e, bagType, indents, lineLen, logicalStart, idx, &tokens)
					}
					logicalStart = 0
					inLogical = false
				}
			case "ASN":
				switch cTok.AsUpper() {
				case ":=":
					inLogical = true
					logicalStart = idx
				case ";":
					if inLogical {
						addPLxLogicalBreaks(e, bagType, indents, lineLen, logicalStart, idx, &tokens)
					}
					logicalStart = 0
					inLogical = false
				}
			}
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

	eCnt := 0
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
				eCnt = 1

				ipd := 0
				// determine how many elements are in the tuple(s)
				for j := idx; j <= idxMax; j++ {
					done := false
					switch tokens[j].AsUpper() {
					case "(":
						ipd++
					case ")":
						ipd--
						done = ipd < 0
					case ",":
						if ipd == 1 {
							eCnt++
						}
						done = ipd <= 0
					case ";", "WHEN":
						done = ipd <= 0
					}

					//carp(debug, fmt.Sprintf("        j: %d, ipd: %d, eCnt: %d, done: %t", j, ipd, eCnt, done))

					if done {
						break
					}

				}
				//carp(debug, fmt.Sprintf("        eCnt: %d", eCnt))

			case ",":
				if inValues {
					if parensDepth == pdl {
						tplStart = idx
						tokens[idx].EnsureVSpace()
						tokens[idx].AdjustIndents(indents)
						if tokens[idxStart].vSpace == 0 {
							tokens[idxStart].EnsureVSpace()
							tokens[idxStart].AdjustIndents(indents)
						}
					}
				}
			}

		case ")":
			switch {
			case parensDepth == pdl:
				if inValues {
					// Wrap, or not, the elements within the tuple

					// Check the next token to determine if there are multiple
					// tuples involved and, if so, determine how to wrap them
					if tplWrapping == 0 {
						if idx < idxMax {
							if tokens[idx+1].value == "," {
								hasMultiTuples = true
								tplWrapping = e.WrapMultiTuples()
							}
						}

						// If a values statement has but one tuple and that
						// tuple has less than 4 elements only wrap it if it
						// would be too long. If there are more elements then
						// wrap regardless of tuple length

						// if there is only one tuple then it gets wrapped based
						// on element count/length
						if !hasMultiTuples {
							if eCnt > 3 {
								tplWrapping = env.WrapAll
							} else {
								tplWrapping = env.WrapLong
							}
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
		case ";", "WHEN":
			inValues = false
			idxStart = 0
			pdl = 0
		}
		ptVal = tokens[idx].value
	}

	return tokens
}
