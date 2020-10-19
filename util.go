package main

const (
	Unknown = iota
	// work unit types
	DCL       // A work unit that belongs to a DCL (GRANT/REVOKE) statement
	DDL       // A work unit that belongs to a DDL statement
	DML       // A work unit that belongs to a DML statement
	PL        // A work unit that belongs to a section of Procedural Language
	Formatted // A work unit that contains a formatted statement
	Final     // A work unit that indicates the end of work units
	// New line requirements
	NoNewLine
	NewLineAllowed
	NewLineRequired
	// DDL clause types
	Select
	Insert
	Update
	Delete
	Merge
	Upsert
	// Control flow
	IfBlock   // an "if ... then ... [elsif ... then ...] [else ...] end if" block
	CaseBlock // a "case when ... then ... [when ... then ...] [else ...] end case" block
	LoopBlock // a "[...] loop ... end loop" block
	// PL blocks
	FuncBlock  // function or procedure block
	BeginBlock // a "begin ... [exception ...] end" block
)

var ConstNames = map[int]string{
	Unknown:         "Unknown",
	DCL:             "DCL",
	DDL:             "DDL",
	DML:             "DML",
	PL:              "PL",
	Formatted:       "Formatted",
	Final:           "Final",
	NoNewLine:       "NoNewLine",
	NewLineAllowed:  "NewLineAllowed",
	NewLineRequired: "NewLineRequired",
	Select:          "SELECT",
	Insert:          "INSERT",
	Update:          "UPDATE",
	Delete:          "DELETE",
	Merge:           "MERGE",
	Upsert:          "UPSERT",
	IfBlock:         "IF",
	CaseBlock:       "CASE",
	LoopBlock:       "LOOP",
	FuncBlock:       "FuncBlock",
	BeginBlock:      "BEGIN",
}

var ConstVals = map[string]int{
	"Unknown":         Unknown,
	"DCL":             DCL,
	"DDL":             DDL,
	"DML":             DML,
	"PL":              PL,
	"Formatted":       Formatted,
	"Final":           Final,
	"NoNewLine":       NoNewLine,
	"NewLineAllowed":  NewLineAllowed,
	"NewLineRequired": NewLineRequired,
	"SELECT":          Select,
	"INSERT":          Insert,
	"UPDATE":          Update,
	"DELETE":          Delete,
	"MERGE":           Merge,
	"UPSERT":          Upsert,
	"IF":              IfBlock,
	"CASE":            CaseBlock,
	"LOOP":            LoopBlock,
	"FuncBlock":       FuncBlock,
	"BEGIN":           BeginBlock,
}

func maxInt(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func chkCommentNL(curr, prev wu, nlChk int) int {

	// if a new line is already required then a new line is still required
	if nlChk == NewLineRequired {
		return NewLineRequired
	}

	// if prev was a line comment then require a newline/indent before
	// if prev was a block comment then allow for a newline/indent before
	// if current is any kind of comment then allow for a newline/indent before
	switch {
	case prev.isLineComment():
		return NewLineRequired
	case prev.isComment():
		return NewLineAllowed
	case curr.isComment():
		return NewLineAllowed
	}
	return nlChk
}
