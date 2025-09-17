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

func TestDomainAnd(t *testing.T) {
	left := odoorpc.NewDomain().
		Equals("name", "test").
		Equals("active", true)
	right := odoorpc.NewDomain().GreaterThan("age", 18)
	got := left.And(right)
	want := []any{
		"&",
		[]any{"&", []any{"name", "=", "test"}, []any{"active", "=", true}},
		[]any{"age", ">", 18},
	}
	if !reflect.DeepEqual([]any(got), want) {
		t.Fatalf("unexpected And domain: %#v", got)
	}
}

func TestDomainOr(t *testing.T) {
	left := odoorpc.NewDomain().Equals("name", "test")
	right := odoorpc.NewDomain().Equals("email", "test@example.com")
	third := odoorpc.NewDomain().GreaterThanOrEqual("score", 80)
	got := left.Or(right, third)
	want := []any{
		"|",
		[]any{"|", []any{"name", "=", "test"}, []any{"email", "=", "test@example.com"}},
		[]any{"score", ">=", 80},
	}
	if !reflect.DeepEqual([]any(got), want) {
		t.Fatalf("unexpected Or domain: %#v", got)
	}
}

func TestDomainAndIgnoresEmpty(t *testing.T) {
	other := odoorpc.NewDomain().Equals("name", "test")
	got := odoorpc.NewDomain().And(other)
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
	}
	if !reflect.DeepEqual([]any(got), []any(want)) {
		t.Fatalf("unexpected domain when left empty: %#v", got)
	}
}
