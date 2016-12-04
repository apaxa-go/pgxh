// Expected ENV:
// PG_HOST
// PG_PORT
// PG_DATABASE
// PG_USER
// PG_PASSWORD

package pgxh

import (
	"github.com/apaxa-go/helper/strconvh"
	"github.com/jackc/pgx"
	"os"
	"reflect"
	"testing"
)

var testTable = "pgxh_table"
var validStmt0 = "SELECT id, name FROM " + testTable + " WHERE id > $1" // Also used for ScanAll tests
var validStmt1 = "SELECT id FROM " + testTable + " WHERE id > $1"       // Also used for ScanAll tests
var invalidStmt = "SELECT id, origin FROM " + testTable
var conn0 *pgx.Conn
var conn1 *pgx.ConnPool
var preparers []PgxPreparer
var queryers []PgxQueryer

func setupDB() {
	var err error
	var conf pgx.ConnConfig
	conf.Host = os.Getenv("PG_HOST")
	conf.Port, _ = strconvh.ParseUint16(os.Getenv("PG_PORT"))
	conf.Database = os.Getenv("PG_DATABASE")
	conf.User = os.Getenv("PG_USER")
	conf.Password = os.Getenv("PG_PASSWORD")

	// Make connections
	conn0, err = pgx.Connect(conf)
	if err != nil {
		panic(err)
	}
	conn1, err = pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: conf, MaxConnections: 10})
	if err != nil {
		panic(err)
	}
	preparers = append(preparers, conn0, conn1)
	queryers = append(queryers, conn0, conn1)

	// Create & fill DB
	cSQL := "CREATE TABLE " + testTable + " (id integer, name text);" +
		"INSERT INTO " + testTable + " (id, name) VALUES (1,'one'),(2,'two'),(3,'three');"
	if _, err = conn0.Exec(cSQL); err != nil {
		panic(err)
	}
}

func cleanDB() {
	cSQL := "DROP TABLE " + testTable + ";"
	if _, err := conn0.Exec(cSQL); err != nil {
		panic(err)
	}

	if conn0 != nil {
		conn0.Close()
	}
	if conn1 != nil {
		conn1.Close()
	}
}

func TestMain(m *testing.M) {
	main := func() int {
		setupDB()
		defer cleanDB()
		return m.Run()
	}
	r := main()
	os.Exit(r)
}

func TestMustPrepare(t *testing.T) {
	tFunc := func(p PgxPreparer, sql string) (stmt string, ok bool) {
		defer func() {
			ok = recover() == nil
		}()
		stmt = MustPrepare(p, sql)
		return stmt, true
	}

	type testElement struct {
		sql string
		ok  bool
	}
	tests := []testElement{
		{validStmt0, true},
		{invalidStmt, false},
	}

	n := stmtNum
	for _, conn := range preparers {
		for _, v := range tests {
			rStmt := "stmt" + strconvh.FormatUint(n)
			if stmt, ok := tFunc(conn, v.sql); ok != v.ok || (ok && stmt != rStmt) {
				t.Errorf("expect '%v' %v, got '%v' %v", rStmt, v.ok, stmt, ok)
			}
			n++
		}
	}
}

func TestMustPrepareInPlace(t *testing.T) {
	tFunc := func(p PgxPreparer, sql string) (stmt string, ok bool) {
		defer func() {
			ok = recover() == nil
		}()
		MustPrepareInPlace(p, &sql)
		return sql, true
	}

	type testElement struct {
		sql string
		ok  bool
	}
	tests := []testElement{
		{validStmt0, true},
		{invalidStmt, false},
	}

	n := stmtNum
	for _, conn := range preparers {
		for _, v := range tests {
			rStmt := "stmt" + strconvh.FormatUint(n)
			if stmt, ok := tFunc(conn, v.sql); ok != v.ok || (ok && stmt != rStmt) {
				t.Errorf("expect '%v' %v, got '%v' %v", rStmt, v.ok, stmt, ok)
			}
			n++
		}
	}
}

func TestMustPrepareAll(t *testing.T) {
	tFunc := func(p PgxPreparer, sqls []string) (stmts []string, ok bool) {
		defer func() {
			ok = recover() == nil
		}()
		stmts = MustPrepareAll(p, sqls...)
		return stmts, true
	}

	type testElement struct {
		sqls []string
		ok   bool
	}
	tests := []testElement{
		{[]string{validStmt0}, true},
		{[]string{validStmt0, validStmt1}, true},
		{[]string{invalidStmt}, false},
		{[]string{validStmt0, invalidStmt}, false},
		{[]string{invalidStmt, validStmt0}, false},
	}

	for _, conn := range preparers {
		for _, v := range tests {
			// Calc right stmts names
			var rStmts []string
			for i := range v.sqls {
				rStmts = append(rStmts, "stmt"+strconvh.FormatUint(stmtNum+uint(i)))
			}

			if stmts, ok := tFunc(conn, v.sqls); ok != v.ok || (ok && !reflect.DeepEqual(rStmts, stmts)) {
				t.Errorf("%v: expect %v '%v', got %v '%v'", v.sqls, v.ok, rStmts, ok, stmts)
			}
		}
	}
}

func TestMustPrepareAllInPlace(t *testing.T) {
	tFunc := func(p PgxPreparer, sqls []string) (stmts []string, ok bool) {
		sqlsCopy := make([]string, len(sqls))
		copy(sqlsCopy, sqls)
		sqlsPointers := make([]*string, len(sqlsCopy))
		for i := range sqlsCopy {
			sqlsPointers[i] = &sqlsCopy[i]
		}

		defer func() {
			ok = recover() == nil
		}()
		MustPrepareAllInPlace(p, sqlsPointers...)
		return sqlsCopy, true
	}

	type testElement struct {
		sqls []string
		ok   bool
	}
	tests := []testElement{
		{[]string{validStmt0}, true},
		{[]string{validStmt0, validStmt1}, true},
		{[]string{invalidStmt}, false},
		{[]string{validStmt0, invalidStmt}, false},
		{[]string{invalidStmt, validStmt0}, false},
	}

	for _, conn := range preparers {
		for _, v := range tests {
			// Calc right stmts names
			var rStmts []string
			for i := range v.sqls {
				rStmts = append(rStmts, "stmt"+strconvh.FormatUint(stmtNum+uint(i)))
			}

			if stmts, ok := tFunc(conn, v.sqls); ok != v.ok || (ok && !reflect.DeepEqual(rStmts, stmts)) {
				t.Errorf("%v: expect %v '%v', got %v '%v'", v.sqls, v.ok, rStmts, ok, stmts)
			}
		}
	}
}
