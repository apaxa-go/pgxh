// Package pgxh implements some helper functions for "github.com/jackc/pgx".
package pgxh

import (
	"github.com/apaxa-go/helper/strconvh"
	"github.com/jackc/pgx"
)

const stmtNamePrefix = "stmt"

var stmtNum uint

// PgxPreparer interface can hold any object that can prepare SQL statements.
// Currently (and is primary used for) it can hold pgx.Conn & pgx.ConnPool.
type PgxPreparer interface {
	Prepare(name, sql string) (*pgx.PreparedStatement, error)
}

// MustPrepare is like pgxConn[Pool].Prepare but panics if the SQL cannot be parsed.
// It simplifies safe initialization of global variables holding prepared statements.
// It also assign name to prepared statement (currently name is "stmt<number>").
func MustPrepare(p PgxPreparer, sql string) (stmtName string) {
	stmtName = stmtNamePrefix + strconvh.FormatUint(stmtNum)
	stmtNum++
	if _, err := p.Prepare(stmtName, sql); err != nil {
		panic(`pgxhelper: Prepare(` + sql + `): ` + err.Error())
	}
	return
}

// MustPrepareAll is like MustPrepare but accept multiple sql to prepare.
// It return slice of statements names.
func MustPrepareAll(p PgxPreparer, sqls ...string) (stmtNames []string) {
	stmtNames = make([]string, len(sqls))
	for i, sql := range sqls {
		stmtNames[i] = MustPrepare(p, sql)
	}
	return
}

// MustPrepareInPlace is modification of MustPrepare with resulting statement name stored in original string with sql.
func MustPrepareInPlace(p PgxPreparer, stmt *string) {
	*stmt = MustPrepare(p, *stmt)
}

// MustPrepareAllInPlace is modification of MustPrepareAll with resulting statements names stored in original strings with sql.
func MustPrepareAllInPlace(p PgxPreparer, stmts ...*string) {
	for _, stmt := range stmts {
		MustPrepareInPlace(p, stmt)
	}
}
