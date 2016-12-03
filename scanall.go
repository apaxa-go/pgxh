package pgxh

import (
	"github.com/apaxa-go/helper/databaseh/sqlh"
	"github.com/jackc/pgx"
)

// PgxQueryer interface can hold any object that can query SQL statements.
// Currently (and is primary used for) it can hold pgx.Conn & pgx.ConnPool.
type PgxQueryer interface {
	Query(sql string, args ...interface{}) (*pgx.Rows, error)
}

// ScanAll is adaptation of "github.com/apaxa-go/helper/databaseh/sqlh" StmtScanAll for "github.com/jackc/pgx".
// ScanAll performs query sql on connection conn with arguments 'args' and stores all result rows in dst.
// sql passed as-is to conn.Query so it is possible to pass prepared statement name as sql.
// ScanAll stop working on first error.
// Example:
// 	type Label struct {
//  		Id       int32
//  		Name     string
// 	}
//
// 	func (l *Label) SQLScanInterface() []interface{} {
// 		return []interface{}{
// 			&l.Id,
// 			&l.Name,
// 		}
// 	}
//
// 	type Labels []*Label
//
// 	func (l *Labels) SQLNewElement() sqlh.SingleScannable {
// 		e := &Label{}
// 		*l = append(*l, e)
// 		return e
// 	}
// 	...
// 	var labels Labels
// 	if err := pgxh.ScanAll(conn, "SELECT id, name FROM LABELS where amount>$1", &labels, someAmount); err != nil {
// 		return err
// 	}
func ScanAll(conn PgxQueryer, sql string, dst sqlh.MultiScannable, args ...interface{}) error {
	rows, err := conn.Query(sql, args...)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		rowContainer := dst.SQLNewElement()
		if err := rows.Scan(rowContainer.SQLScanInterface()...); err != nil {
			return err
		}
	}

	return rows.Err()
}
