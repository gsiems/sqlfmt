package formatter

import (
	"fmt"
	"strings"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
	"github.com/gsiems/sqlfmt/parser"
)

const (
	////////////////////////////////////////////////////////////////////
	// Token Bag Categories and types
	DNFBag       = iota + 400 // A bag of Do-Not-Format tokens (Postgres Pl/Perl, PL/Python, PL/Tcl, COPY, etc.)
	DCLBag                    // A bag of DCL tokens
	DDLBag                    // A bag of DDL tokens
	DMLBag                    // A bag of DML tokens
	DMLCaseBag                // A bag of DML CASE statement tokens
	PLxBag                    // a bag of function/procedure/package tokens
	PLxBody                   // A bag of function/procedure/package body tokens
	CommentOnBag              // A bag of "COMMENT ON ..." tokens
)

func tagBags(e *env.Env, m []FmtToken) (map[string]TokenBag, []FmtToken) {

	bagMap := make(map[string]TokenBag)

	remainder := tagCommentOn(e, m, bagMap)
	remainder = tagDCL(e, remainder, bagMap)
	remainder = tagDML(e, remainder, bagMap)

	// TODO: for now at least. need to revisit once other DBs (especially
	// Oracle) are better sorted
	switch e.Dialect() {
	case dialect.PostgreSQL:
		remainder = tagPLx(e, remainder, bagMap)
	}
	remainder = tagDDL(e, remainder, bagMap)

	// Check for warnings and errors
	var warnings []string // list of (non-fatal) warnings found
	var errors []string   // list of (fatal) errors found

	for _, bag := range bagMap {

		if len(bag.warnings) > 0 {
			warnings = append(warnings, bag.warnings...)
		}
		if len(bag.errors) > 0 {
			errors = append(errors, bag.errors...)
		}

		parensDepth := 0

		for _, t := range bag.tokens {
			switch t.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}
		}

		if parensDepth != 0 {

			label := ""
			switch bag.typeOf {
			case CommentOnBag: // should really never, ever, ever happen unless there was some fat-fingering going on
				label = "COMMENT ON statement"
			case DCLBag:
				label = "DCL statement"
			case DDLBag:
				label = "DDL statement"
			case DMLBag, DMLCaseBag:
				label = "DML statement"
			case PLxBag, PLxBody:
				label = "PL code"
			}

			if bag.forObj != "" {
				errors = append(errors, fmt.Sprintf("%d unbalanced parenthesis found while parsing %s for %s", parensDepth, label, bag.forObj))
			} else {
				errors = append(errors, fmt.Sprintf("%d unbalanced parenthesis found while parsing %s", parensDepth, label))
			}
		}
	}

	return bagMap, remainder
}

func FormatInput(e *env.Env, input string) (string, error) {

	p := parser.NewParser(e.DialectName())
	parsed, err := p.ParseStatements(input)
	if err != nil {
		return "", err
	}

	cleaned := prepParsed(e, parsed)
	bagMap, mainTokens := tagBags(e, cleaned)
	fmtTokens := formatBags(e, mainTokens, bagMap)
	untagged := untagBags(fmtTokens, bagMap)
	unstashed := unstashComments(e, untagged)
	fmtStatement := combineTokens(e, unstashed)

	return fmtStatement, nil
}

// stashComments caches comments with their adjoining non-comment token for the
// purpose of simplifying formatting logic. (also translates parser tokens to
// formatting tokens)
func stashComments(e *env.Env, tokens []parser.Token) []FmtToken {

	var ret []FmtToken
	var lCmts []CmtToken

	for idx, cTok := range tokens {

		vSpace := cTok.VSpace()
		hSpace := ""

		switch vSpace {
		case 0:
			if idx > 0 {
				hSpace = " "
			}
		case 1, 2:
			hSpace = ""
		default:
			vSpace = 2
			hSpace = ""
		}

		switch cTok.Type() {
		case parser.WhiteSpace:
			// We just don't care about the trailing whitespace
			continue
		case parser.LineComment, parser.PoundLineComment, parser.BlockComment:

			nt := CmtToken{
				typeOf: cTok.Type(),
				value:  cTok.Value(),
				vSpace: vSpace,
				hSpace: hSpace,
				//vSpaceOrig: cTok.VSpace()
				//hSpaceOrig  cTok.HSpace()
			}

			// If the comment has no vertical space and is not the first token
			// then it is cached as a trailing comment to the preceding non-comment
			// token.
			// If the comment has one vertical space and is tight to the preceding
			// non-comment then it is also cached as a trailing comment to the
			// preceding non-comment token.
			// If the comment, or any immediate preceding comments, have more than
			// one vertical space then it is cached and added as a leading comment
			// to the next non-comment token

			switch {
			case len(lCmts) > 0:
				lCmts = append(lCmts, nt)
			//case cTok.VSpace() > 1:
			case cTok.VSpace() > 0:
				lCmts = append(lCmts, nt)
			case len(ret) == 0:
				lCmts = append(lCmts, nt)
			default:
				ret[len(ret)-1].AddTrailingComment(nt)
			}
			/*
				switch {
				case len(ret) == 0:
					lCmts = append(lCmts, nt)
				case len(lCmts) > 0:
					lCmts = append(lCmts, nt)
				default:
					switch cTok.VSpace() {
					case 0, 1:
						ret[len(ret)-1].AddTrailingComment(nt)
					default:
						lCmts = append(lCmts, nt)
					}
				}
			*/
		default:
			nt := FmtToken{
				categoryOf: cTok.Category(),
				typeOf:     cTok.Type(),
				value:      cTok.Value(),
				vSpace:     vSpace,
				hSpace:     hSpace,
				vSpaceOrig: cTok.VSpace(),
				hSpaceOrig: cTok.HSpace(),
			}

			if len(lCmts) > 0 {
				nt.AddLeadingComment(lCmts...)
				lCmts = nil
			}
			ret = append(ret, nt)
		}

	}

	// If there are comments at the end of the input then simply append them to
	// the final non-comment token
	if len(lCmts) > 0 && len(ret) > 0 {
		ret[len(ret)-1].AddTrailingComment(lCmts...)
	}

	return ret
}

// consolidateMWTokens consolidates multi-word tokens such as "END IF",
// "END CASE", or "END LOOP" by combining them into a single token.
func consolidateMWTokens(e *env.Env, tokens []FmtToken) []FmtToken {

	var ret []FmtToken

	idxMax := len(tokens) - 1

	// If, for some reason, there is are comments between the tokens then
	// append them as trailing comments
	skipNext := false
	for idx := 0; idx <= idxMax; idx++ {
		if skipNext {
			skipNext = false
			continue
		}

		cTok := tokens[idx]

		combineNext := false
		if idx < idxMax {

			switch cTok.AsUpper() {
			case "END":
				switch tokens[idx+1].AsUpper() {
				case "IF", "CASE", "LOOP":
					combineNext = true
				}
			case "GROUP", "ORDER", "PARTITION":
				switch tokens[idx+1].AsUpper() {
				case "BY":
					combineNext = true
				}
			case "FROM":
				switch tokens[idx+1].AsUpper() {
				case "DISTINCT":
					combineNext = true
				}
			case "FOR":
				switch tokens[idx+1].AsUpper() {
				case "UPDATE":
					combineNext = true
				}
			case "ON":
				switch tokens[idx+1].AsUpper() {
				case "CONFLICT":
					combineNext = true
				}
			case "MERGE":
				switch tokens[idx+1].AsUpper() {
				case "INTO":
					combineNext = true
				}
			case "DEFAULT":
				switch tokens[idx+1].AsUpper() {
				case "VALUES":
					combineNext = true
				}
			}

			if combineNext {
				nTok := tokens[idx+1]
				ntVal := nTok.AsUpper()

				cTok.value = cTok.value + " " + ntVal
				skipNext = true

				if len(nTok.ledComments) > 0 {
					cTok.AddTrailingComment(nTok.ledComments...)
				}
				if len(nTok.trlComments) > 0 {
					cTok.AddTrailingComment(nTok.trlComments...)
				}
			}
		}

		ret = append(ret, cTok)
	}

	return ret
}

// consolidateDatatypes consolidates tokens that make up a datatype declaration
// by combining them into a single token.
func consolidateDatatypes(e *env.Env, tokens []FmtToken) []FmtToken {

	var ret []FmtToken

	idxMax := len(tokens) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := tokens[idx]

		switch cTok.categoryOf {
		case parser.Comment, parser.String, parser.Punctuation, parser.Data, parser.Other:
			ret = append(ret, cTok)
			continue
		}

		idxEnd := min(idxMax, idx+8)

		if idxEnd == idx {
			ret = append(ret, cTok)
			continue
		}

		idxLen := idxEnd - idx
		dtLen := 0

		for i := 1; i <= idxLen; i++ {
			switch tokens[idx+i].categoryOf {
			case parser.Comment, parser.String, parser.Data:
				break
			}
			switch tokens[idx+i].typeOf {
			case parser.BindParameter, parser.Label, parser.Operator, parser.NullItem, parser.WhiteSpace:
				break
			}

			if isDatatype(e, tokens[idx:idx+i]) {
				dtLen = max(dtLen, i)
			}
		}

		if dtLen > 1 {
			dts := asDatatypeString(tokens[idx : idx+dtLen])
			cTok.value = dts
			idx += dtLen - 1
		}
		ret = append(ret, cTok)
	}
	return ret
}

func isDatatype(e *env.Env, s []FmtToken) bool {
	dbdialect := dialect.NewDialect(e.DialectName())
	var ary []string
	for _, t := range s {
		ary = append(ary, t.value)
	}
	return dbdialect.IsDatatype(ary...)
}

func asDatatypeString(s []FmtToken) string {
	var z []string
	pv := ""

	for _, t := range s {
		v := t.value
		switch v {
		case "(":
			z = append(z, " "+v)
		case ")", ",", "[", "]":
			z = append(z, v)
		default:
			switch pv {
			case "(", ",", "":
				z = append(z, v)
			default:
				z = append(z, " "+v)
			}
		}
		pv = v
	}
	return strings.Join(z, "")
}

func prepParsed(e *env.Env, parsed []parser.Token) (ret []FmtToken) {

	dbdialect := dialect.NewDialect(e.DialectName())

	foldingCase := dbdialect.CaseFolding()
	// 1. Give each token a unique ID.
	// 2. Review the tokens to unquote those identifiers as may be unquoted
	// 3. Adjust the token type as needed
	// 4. Perform case folding of identifiers, datatypes, and keywords as
	//      specified in the env

	p1 := stashComments(e, parsed)
	p2 := consolidateDatatypes(e, p1)
	p3 := consolidateMWTokens(e, p2)

	idxMax := len(p3) - 1

	for idx := 0; idx <= idxMax; idx++ {

		cTok := p3[idx]

		tText := cTok.value
		tType := cTok.typeOf
		tCategory := cTok.categoryOf

		switch tCategory {
		case parser.Identifier:

			if !e.PreserveQuoting() {
				tryUnquoting := false
				switch tType {
				case parser.DoubleQuoted:
					tryUnquoting = true
				case parser.BracketQuoted:
					switch e.Dialect() {
					case dialect.MSSQL, dialect.SQLite:
						tryUnquoting = true
					}
				case parser.BacktickQuoted:
					switch e.Dialect() {
					case dialect.MariaDB, dialect.MySQL, dialect.SQLite:
						tryUnquoting = true
					}
				}

				if tryUnquoting {

					// Determine if the quoted identifier can be unquoted.
					// Unquote the identifier for testing
					tTest := tText[1 : len(tText)-1]

					// IIF the unquoted token is still a valid identifier (no funky chars)
					// AND is not a reserved word
					//if dbdialect.IsIdentifier(tTest) && !dbdialect.IsReservedKeyword(tTest) {
					if dbdialect.IsIdentifier(tTest) && !dbdialect.IsKeyword(tTest) {

						// if the folding is upper AND the token is upper then the token can be unquoted.
						// if the folding is lower AND the token is lower then the token can be unquoted.
						switch foldingCase {
						case dialect.FoldUpper:
							if tTest == strings.ToUpper(tTest) {
								tText = tTest
								tType = parser.Identifier
							}
						case dialect.FoldLower:
							if tTest == strings.ToLower(tTest) {
								tText = tTest
								tType = parser.Identifier
							}
						}
					}
				}
			}
		}

		switch tType {
		case parser.Identifier, parser.Datatype, parser.Keyword:
			tText = strings.ToLower(tText)
		}

		cTok.id = idx
		cTok.categoryOf = tCategory
		cTok.typeOf = tType
		cTok.value = tText
		ret = append(ret, cTok)
	}

	return ret
}

func formatBags(e *env.Env, m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// remember that bagMap is a pointer to the map, not a copy of the map

	var mainTokens []FmtToken
	parensDepth := 0
	var pTok FmtToken // The previous token

	for _, cTok := range m {

		switch {
		case cTok.IsBag():
			formatBag(e, bagMap, cTok.typeOf, cTok.id, parensDepth, false)
		case cTok.IsCodeComment():
			cTok = formatCodeComment(e, cTok, parensDepth)
		default:
			switch parensDepth {
			case 0:
				if cTok.IsKeyword() && !cTok.IsDatatype() {
					cTok.SetKeywordCase(e, []string{cTok.AsUpper()})
				}
			}

			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			}
			cTok.AdjustVSpace(false, true)
			if cTok.vSpace == 0 {
				cTok.AdjustHSpace(e, pTok)
			}
		}

		pTok = cTok
		mainTokens = append(mainTokens, cTok)
	}
	/*
		parensDepth = 0
		for _, cTok := range mainTokens {
			switch cTok.value {
			case "(":
				parensDepth++
			case ")":
				parensDepth--
			default:
				if cTok.IsBag() {
					AdjustLineWrapping(e, bagMap, cTok.typeOf, cTok.id, parensDepth)
				}
			}
		}
	*/
	return mainTokens
}

func formatBag(e *env.Env, bagMap map[string]TokenBag, bagType, bagId, baseIndents int, forceInitVSpace bool) {

	// remember that bagMap is a pointer to the map, not a copy of the map

	key := bagKey(bagType, bagId)

	if b, ok := bagMap[key]; ok {

		switch b.typeOf {
		case CommentOnBag:
			formatCommentOn(e, bagMap, b.typeOf, b.id, baseIndents, forceInitVSpace)
		case DCLBag:
			formatDCLBag(e, bagMap, b.typeOf, b.id, baseIndents, forceInitVSpace)
		case DDLBag:
			formatDDLBag(e, bagMap, b.typeOf, b.id, baseIndents, forceInitVSpace)
		case DMLBag, DMLCaseBag:
			formatDMLBag(e, bagMap, b.typeOf, b.id, baseIndents, forceInitVSpace)
		case PLxBag, PLxBody:
			formatPLxBag(e, bagMap, b.typeOf, b.id, baseIndents, forceInitVSpace)
		case DNFBag:
			// nada
		}
	}
}

func commentsToTokens(toks []CmtToken) []FmtToken {

	var ret []FmtToken
	for _, cTok := range toks {
		ret = append(ret, FmtToken{
			categoryOf: parser.Comment,
			typeOf:     cTok.typeOf,
			value:      cTok.value,
			vSpace:     cTok.vSpace,
			hSpace:     cTok.hSpace,
			indents:    cTok.indents,
		})
	}
	return ret
}

func unpackTokens(toks ...FmtToken) []FmtToken {
	var ret []FmtToken

	for _, cTok := range toks {
		if len(cTok.ledComments) > 0 {
			ret = append(ret, commentsToTokens(cTok.ledComments)...)
			cTok.ledComments = nil
		}
		ret = append(ret, cTok)
		if len(cTok.trlComments) > 0 {
			ret = append(ret, commentsToTokens(cTok.trlComments)...)
			cTok.trlComments = nil
		}
	}

	return ret
}

func untagBags(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	var ret []FmtToken

	for _, cTok := range m {

		switch {
		case cTok.IsBag():

			// get the key
			// look up the key in the map
			// copy the tokens from the map
			key := bagKey(cTok.typeOf, cTok.id)

			tb, ok := bagMap[key]
			if ok {
				ret = append(ret, untagBags(tb.tokens, bagMap)...)
			}
		default:
			ret = append(ret, cTok)
		}
	}
	return ret
}

func unstashComments(e *env.Env, tokens []FmtToken) []FmtToken {

	var ret []FmtToken

	lIndents := 0
	tIndents := 0
	lpd := 0

	for _, cTok := range tokens {

		if cTok.vSpace > 0 {
			lIndents = cTok.indents
			lpd = 0

			// TODO: the desired indentation of the trailing comments isn't
			// always the same as for the leading comments. Having lost the
			// bagType information at this point we need another way of
			// determining the trailing indentation.

			tIndents = lIndents
		}

		switch cTok.value {
		case "(":
			lpd++
		case ")":
			lpd--
		}

		if len(cTok.ledComments) > 0 {
			for _, ct := range cTok.ledComments {
				nt := FmtToken{
					categoryOf: parser.Comment,
					typeOf:     ct.typeOf,
					value:      ct.value,
					vSpace:     ct.vSpace,
					hSpace:     ct.hSpace,
					indents:    ct.indents,
				}
				if nt.vSpace > 0 {
					nt.indents = lIndents + lpd
				}
				ret = append(ret, nt)
			}
			cTok.ledComments = nil
		}

		ret = append(ret, cTok)

		if len(cTok.trlComments) > 0 {
			for _, ct := range cTok.trlComments {
				nt := FmtToken{
					categoryOf: parser.Comment,
					typeOf:     ct.typeOf,
					value:      ct.value,
					vSpace:     ct.vSpace,
					hSpace:     ct.hSpace,
					indents:    ct.indents,
				}
				if nt.vSpace > 0 {
					nt.indents = tIndents + lpd
				}
				ret = append(ret, nt)
			}
			cTok.trlComments = nil
		}

	}
	return ret
}

func combineTokens(e *env.Env, tokens []FmtToken) string {

	var z []string

	for _, cTok := range tokens {
		if cTok.vSpace > 0 {
			z = append(z, strings.Repeat("\n", cTok.vSpace))
			if cTok.indents > 0 {
				z = append(z, strings.Repeat(e.Indent(), cTok.indents))
			}
		} else if cTok.hSpace != "" {
			z = append(z, cTok.hSpace)
		}

		z = append(z, cTok.value)
	}

	return strings.TrimSpace(strings.Join(z, "")) + "\n"
}
