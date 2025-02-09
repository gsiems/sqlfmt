package formatter

import (
	"github.com/gsiems/sqlfmt/env"
)

type plObj struct {
	id         int
	objType    string
	hasIs      bool
	hasLang    bool
	beginDepth int
}

// tagOraPL ensures that the DDL for creating Oracle functions, procedures,
// packages and triggers are properly tagged
func tagOraPL(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// One issue with tagging Oracle functions, procedures, and packages is
	// that they can have sub-functions and procedures (which is pretty much the
	// definition of a package). The same might be true for triggers.

	var remainder []FmtToken
	tokMap := make(map[int][]FmtToken) // map[bagID][]FmtToken
	bagId := 0
	pKwVal := ""
	plCnt := 0
	objs := make(map[int]plObj)

	for _, cTok := range m {

		ctVal := cTok.AsUpper()
		openBag := false
		closeBag := false

		switch ctVal {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "PACKAGE", "PACKAGE BODY", "TYPE BODY":
			switch pWkVal {
			case "DROP", "ALTER":
				// nada
			default:
				openBag = true
			}
		default:
			if plCnt > 0 {
				tokMap[bagId] = append(tokMap[bagId], cTok)

				switch ctVal {
				case "IS", "AS":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.hasIs = true
						objs[plCnt] = n
					}
				case "LANGUAGE":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.hasLang = true
						objs[plCnt] = n
					}
				case "DECLARE":
					if _, ok := objs[plCnt]; !ok {
						openBag = true
					}
				case "BEGIN":
					_, ok := objs[plCnt]
					if ok {
						n := objs[plCnt]
						n.beginDepth++
						objs[plCnt] = n
					} else {
						openBag = true
					}
				case "END":
					if _, ok := objs[plCnt]; ok {
						n := objs[plCnt]
						n.beginDepth--
						objs[plCnt] = n
					}
				case ";":
					if obj, ok := objs[plCnt]; ok {
						switch {
						case pKwVal == "END":
							if obj.beginDepth <= 0 {
								closeBag = true
							}
						case obj.hasLang:
							closeBag = true
						case !obj.hasIs:
							closeBag = true
						}
					}
				}
			} else {
				switch ctVal {
				case "DECLARE", "BEGIN":
					openBag = true
				default:
					remainder = append(remainder, cTok)
				}
			}
		}

		switch {
		case closeBag:
			if _, ok := objs[plCnt]; ok {
				delete(objs, plCnt)
			}
			plCnt--
			bagId = 0
			switch {
			case plCnt > 0:
				if obj, ok := objs[plCnt]; ok {
					bagId = obj.id
				}
			}

		case openBag:
			parentId := 0
			if plCnt > 0 {
				if obj, ok := objs[plCnt]; ok {
					parentId = obj.id
				}
			}

			hasIs := false
			bd := 0
			if ctVal == "BEGIN" {
				bd++
				hasIs = true // not really, but this is probably an anonymous pl block so...
			}

			bagId = cTok.id
			plCnt++
			objs[plCnt] = plObj{id: cTok.id, objType: ctVal, hasIs: hasIs, hasLang: false, beginDepth: bd}

			nt := FmtToken{
				id:          cTok.id,
				categoryOf:  PLxBag,
				typeOf:      PLxBag,
				vSpace:      cTok.vSpace,
				indents:     cTok.indents,
				hSpace:      cTok.hSpace,
				vSpaceOrig:  cTok.vSpaceOrig,
				hSpaceOrig:  cTok.hSpaceOrig,
				ledComments: cTok.ledComments,
				trlComments: cTok.trlComments,
			}

			tokMap[bagId] = []FmtToken{cTok}

			switch parentId {
			case 0:
				// token is added to the map, remainder gets new pointer token
				remainder = append(remainder, nt)
			default:
				// token is added to the child map, parent map gets new pointer token (to the child)
				tokMap[parentId] = append(tokMap[parentId], nt)
			}
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	////////////////////////////////////////////////////////////////////
	// If the token map is not empty (PL was found and tagged) then populate
	// the bagMap
	for id, tokens := range tokMap {
		key := bagKey(PLxBag, id)
		bagMap[key] = TokenBag{
			id:     id,
			typeOf: PLxBag,
			tokens: tokens,
		}
	}

	return remainder
}
