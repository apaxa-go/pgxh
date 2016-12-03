package pgxh

import (
	"github.com/apaxa-go/helper/databaseh/sqlh"
	"reflect"
	"testing"
)

type Label struct {
	ID   int32
	Name string
}

func (l *Label) SQLScanInterface() []interface{} {
	return []interface{}{
		&l.ID,
		&l.Name,
	}
}

type Labels []*Label

func (l *Labels) SQLNewElement() sqlh.SingleScannable {
	e := &Label{}
	*l = append(*l, e)
	return e
}

func TestScanAll(t *testing.T) {
	type testElement struct {
		stmt      string
		threshold int
		err       bool
		r         Labels
	}
	tests := []testElement{
		{validStmt0, 1, false, []*Label{{2, "two"}, {3, "three"}}},
		{validStmt0, 2, false, []*Label{{3, "three"}}},
		{invalidStmt, 1, true, nil},
		{validStmt1, 1, true, nil},
	}

	for _, conn := range queryers {
		for _, v := range tests {
			var labels Labels
			if err := ScanAll(conn, v.stmt, &labels, v.threshold); (err != nil) != v.err {
				t.Errorf("expect error %v, got %v", v.err, err)
			} else if !v.err && !reflect.DeepEqual(v.r, labels) {
				t.Errorf("expect '%v', got '%v'", v.r, labels)
			}
		}
	}
}
