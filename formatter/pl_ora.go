package formatter

// tagOraPL ensures that the DDL for creating Oracle functions, procedures,
// packages and triggers are properly tagged
func tagOraPL(m []FmtToken, bagMap map[string]TokenBag) []FmtToken {

	// TODO

	// One issue with tagging Oracle functions, procedures, and packages is
	// that they can have sub-functions and procedures (which is pretty much the
	// definition of a package). The same might be true for triggers.

	var remainder []FmtToken
	tokMap := make(map[int][]FmtToken) // map[bagID][]FmtToken
	var blkStack plStack

	// CREATE OR REPLACE ... [PACKAGE|PACKAGE BODY|FUNCTION|PROCEDURE|TRIGGER]

	// 1. Loop through the tokens until the end of the function/package/procedure

	bagId := 0
	isInBag := false
	pKwVal := ""
	objType := ""

	for _, cTok := range m {

		ctVal := cTok.AsUpper()

		switch ctVal {
		case "FUNCTION", "PROCEDURE", "TRIGGER", "PACKAGE":
			if objType == "" {
				objType = ctVal
			}
		case "BODY":
			switch pKwVal {
			case "PACKAGE", "TYPE":
				objType = pKwVal + " BODY"
			}
		}

		switch isInBag {
		case true:

			tokMap[bagId] = append(tokMap[bagId], cTok)

			switch ctVal {
			case "FUNCTION", "PROCEDURE":
				if objType == "PACKAGE BODY" {
					blkStack.Upsert(ctVal)
				}
			case "BEGIN":
				blkStack.Upsert(ctVal)
			case ";":
				if pKwVal == "END" {
					_ = blkStack.Pop()
					if blkStack.IsEmpty() {
						isInBag = false
						pKwVal = ""
						objType = ""
					}
				}
			}

		case false:

			// check for the beginning of the PL object
			switch ctVal {
			case "FUNCTION", "PACKAGE", "PROCEDURE", "TRIGGER":
				// Open a new bag
				isInBag = true
				bagId = cTok.id
				tokMap[bagId] = []FmtToken{cTok}
				blkStack.Push(ctVal)

				// Add a token that has the pointer to the new bag
				remainder = append(remainder, FmtToken{
					id:          bagId,
					categoryOf:  PLxBag,
					typeOf:      PLxBag,
					vSpace:      cTok.vSpace,
					indents:     cTok.indents,
					hSpace:      cTok.hSpace,
					vSpaceOrig:  cTok.vSpaceOrig,
					hSpaceOrig:  cTok.hSpaceOrig,
					ledComments: cTok.ledComments,
					trlComments: cTok.trlComments,
				})

			default:
				// Not in any PL object
				remainder = append(remainder, cTok)
			}
		}

		////////////////////////////////////////////////////////////////
		// Cache the previous token(s) data
		if cTok.IsKeyword() {
			pKwVal = ctVal
		}
	}

	// for now, just to figure out grabbing all the PL
	for id, bt := range tokMap {
		key := bagKey(PLxBag, id)
		bagMap[key] = TokenBag{
			id:     id,
			typeOf: PLxBag,
			tokens: bt,
		}
	}

	/*
		// 2. Loop through the extracted tokens to split [nested] headers from [nested] bodies
		// (where the body is between the first BEGIN and the final END)
		for wrapBagId, bt := range tokMap {

			bagToks := make([]FmtToken, 0)
			bodyToks := make([]FmtToken, 0)
			bodyMap := make(map[int][]FmtToken) // map[bagID][]FmtToken

			bodyBagId := 0
			isInBody := false
			//objType = ""
			pKwVal := ""

			blkStack.Reset()

			for _, cTok := range bt {
				ctVal := cTok.AsUpper()

				//switch ctVal {
				//case "FUNCTION", "PROCEDURE", "TRIGGER", "PACKAGE":
				//	if objType == "" {
				//		objType = ctVal
				//	}
				//case "BODY":
				//	if pKwVal == "PACKAGE" {
				//		objType = ctVal
				//	}
				//}
				//blen := blkStack.Length()
				//inb := isInBody
				switch isInBody {
				case true:

					bodyToks = append(bodyToks, cTok)

					switch ctVal {
					case "BEGIN":
						blkStack.Upsert(ctVal)
					case ";":
						if pKwVal == "END" {
							_ = blkStack.Pop()
							if blkStack.IsEmpty() {

								bodyMap[bodyBagId] = bodyToks
								bodyToks = make([]FmtToken, 0)

								isInBody = false
								pKwVal = ""
								bodyBagId = 0
							}
						}
					}

				default:
					// assume false is the default

					switch ctVal {
					case "BEGIN":
						isInBody = true
					case "IS", "AS":
						// if the previous keyword is "PACKAGE" or "BODY" then we are not yet in the body of any function/procedure
						switch pKwVal {
						case "PACKAGE", "BODY":
						// nada
						default:
							isInBody = true
						}
					}

					if isInBody {
						blkStack.Upsert(ctVal)

						bodyBagId = cTok.id
						bodyToks = nil
						bodyToks = []FmtToken{cTok}

						bagToks = append(bagToks, FmtToken{
							id:         bodyBagId,
							categoryOf: PLxBag,
							typeOf:     PLxBody,
							vSpace:     cTok.vSpace,
							indents:    cTok.indents,
							hSpace:     cTok.hSpace,
							vSpaceOrig: cTok.vSpaceOrig,
							hSpaceOrig: cTok.hSpaceOrig,
						})
					} else {
						// we're not in the body so we must be in the wrapper
						bagToks = append(bagToks, cTok)
					}

				}

				//log.Printf(" %d, %d, %t, %t, %q", blen, blkStack.Length(), inb, isInBody, ctVal)

				// Track the previous keyword token
				if cTok.IsKeyword() {
					pKwVal = ctVal
				}

			}

			if len(bodyToks) > 0 {
				bodyMap[bodyBagId] = bodyToks
			}

			key := bagKey(PLxBag, wrapBagId)
			bagMap[key] = TokenBag{
				typeOf: PLxBag,
				id:     wrapBagId,
				tokens: bagToks,
			}

			for id, bt := range bodyMap {
				key := bagKey(PLxBody, id)
				bagMap[key] = TokenBag{
					typeOf: PLxBody,
					id:     id,
					tokens: bt,
				}
			}

		}
	*/
	/*
		for _, t := range remainder {
			log.Printf("                     %s", t.String())
		}
			keys := make([]string, 0, len(bagMap))

			for key := range bagMap {
				keys = append(keys, key)
			}
			sort.Strings(keys)

			for _, key := range keys {
				log.Print("")
				bagId := bagMap[key].id
				bagType := bagMap[key].TypeOf

				for _, t := range bagMap[key].tokens {
					log.Printf("%6d %-12s: %s", bagId, TokenName(bagType), t.String())
				}
			}
	*/
	//log.Printf("    10 (%d)", len(remainder))

	return remainder
}
