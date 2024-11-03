package dialect

const (
	////////////////////////////////////////////////////////////////////
	// SQL Dialects
	StandardSQL = iota + 200
	PostgreSQL
	SQLite
	MySQL
	MariaDB
	Oracle
	MSSQL
	////////////////////////////////////////////////////////////////////
	// Case folding
	FoldLower
	FoldUpper
	NoFolding
)