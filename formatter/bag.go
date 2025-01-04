package formatter

import (
	"fmt"

	"github.com/gsiems/sqlfmt/dialect"
	"github.com/gsiems/sqlfmt/env"
)

type TokenBag struct {
	id       int        // the ID for the bag
	typeOf   int        // the type of token bag
	forObj   string     // the name of the kind of object that the bag is for (not all bag types care)
	tokens   []FmtToken // the list of token that make up the bag
	warnings []string   // list of (non-fatal) warnings found
	errors   []string   // list of (fatal) errors found
}

func (t *TokenBag) HasLeadingComments() bool {
	if len(t.tokens) == 0 {
		return false
	}
	return len(t.tokens[0].ledComments) > 0
}

func (t *TokenBag) HasTrailingComments() bool {
	if len(t.tokens) == 0 {
		return false
	}
	return len(t.tokens[len(t.tokens)-1].trlComments) > 0
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

			// Close the bag
			isInBag = false

			key := bagKey(bagType, bagId)
			bagMap[key] = TokenBag{
				id:     bagId,
				typeOf: bagType,
				forObj: forObj,
				tokens: bagTokens,
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

		bagMap[key] = TokenBag{
			id:     bagId,
			typeOf: bagType,
			forObj: forObj,
			tokens: bagTokens,
		}
	}

	return remainder
}

func UpsertMappedBag(bagMap map[string]TokenBag, bagType, bagId int, forObj string, tokens []FmtToken) {

	key := bagKey(bagType, bagId)

	_, ok := bagMap[key]
	if ok {
		delete(bagMap, key)
	}

	bagMap[key] = TokenBag{
		id:     bagId,
		typeOf: bagType,
		forObj: forObj,
		tokens: tokens,
	}
}
