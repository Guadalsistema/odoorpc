package odoorpc_test

import (
	"reflect"
	"testing"

	"github.com/Guadalsistema/odoorpc"
)

func TestDomainBuilder(t *testing.T) {
	d := odoorpc.NewDomain().
		Equals("name", "test").
		In("ids", []int64{1, 2}).
		ChildOf("parent_id", 10).
		LessThan("age", 30).
		GreaterThan("score", 50).
		LessThanOrEqual("level", 3).
		GreaterThanOrEqual("rank", 4)
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
		[]any{"ids", "in", []any{int64(1), int64(2)}},
		[]any{"parent_id", "child_of", 10},
		[]any{"age", "<", 30},
		[]any{"score", ">", 50},
		[]any{"level", "<=", 3},
		[]any{"rank", ">=", 4},
	}
	if !reflect.DeepEqual([]any(d), []any(want)) {
		t.Fatalf("unexpected domain: %#v", d)
	}
}
