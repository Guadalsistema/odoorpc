package odoorpc_test

import (
	"reflect"
	"testing"

	"github.com/Guadalsistema/odoorpc"
)

func TestParseDomainEmpty(t *testing.T) {
	got, err := odoorpc.ParseDomain("[]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty domain, got %#v", got)
	}
}

func TestParseDomainSingleCondition(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('name', '=', 'test')]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainMultipleConditions(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('name', '=', 'test'), ('active', '=', True)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
		[]any{"active", "=", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainBooleanFalse(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('cond', '=', False)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"cond", "=", false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainBooleanCaseInsensitive(t *testing.T) {
	tests := []struct {
		input string
		want  any
	}{
		{"[('a', '=', True)]", true},
		{"[('a', '=', true)]", true},
		{"[('a', '=', FALSE)]", false},
		{"[('a', '=', false)]", false},
	}
	for _, tt := range tests {
		got, err := odoorpc.ParseDomain(tt.input)
		if err != nil {
			t.Fatalf("unexpected error for %q: %v", tt.input, err)
		}
		tuple, ok := got[0].([]any)
		if !ok || len(tuple) != 3 {
			t.Fatalf("expected tuple for %q, got %#v", tt.input, got[0])
		}
		if tuple[2] != tt.want {
			t.Fatalf("for %q: got %#v, want %#v", tt.input, tuple[2], tt.want)
		}
	}
}

func TestParseDomainNone(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('field', '=', None)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"field", "=", nil},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainIntegers(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('age', '>', 18), ('id', '=', 42)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"age", ">", int64(18)},
		[]any{"id", "=", int64(42)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainNegativeNumber(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('balance', '<', -100)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"balance", "<", int64(-100)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainFloat(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('price', '>=', 99.99)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"price", ">=", 99.99},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainInOperator(t *testing.T) {
	got, err := odoorpc.ParseDomain("[('id', 'in', [1, 2, 3])]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"id", "in", []any{int64(1), int64(2), int64(3)}},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainOrOperator(t *testing.T) {
	got, err := odoorpc.ParseDomain("['|', ('name', '=', 'a'), ('name', '=', 'b')]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		[]any{"name", "=", "a"},
		[]any{"name", "=", "b"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainAndOperator(t *testing.T) {
	got, err := odoorpc.ParseDomain("['&', ('x', '=', 1), ('y', '=', 2)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"&",
		[]any{"x", "=", int64(1)},
		[]any{"y", "=", int64(2)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainNotOperator(t *testing.T) {
	got, err := odoorpc.ParseDomain("['!', ('active', '=', False)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"!",
		[]any{"active", "=", false},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainComplexNested(t *testing.T) {
	input := "['|', '|', ('state', '=', 'draft'), ('state', '=', 'sent'), ('state', '=', 'confirmed')]"
	got, err := odoorpc.ParseDomain(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		"|",
		[]any{"state", "=", "draft"},
		[]any{"state", "=", "sent"},
		[]any{"state", "=", "confirmed"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainDoubleQuotedStrings(t *testing.T) {
	got, err := odoorpc.ParseDomain(`[("name", "=", "test")]`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainWhitespace(t *testing.T) {
	got, err := odoorpc.ParseDomain(`  [  ( 'name' ,  '=' ,  'test' )  ]  `)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		[]any{"name", "=", "test"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainEmptyString(t *testing.T) {
	got, err := odoorpc.ParseDomain("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != 0 {
		t.Fatalf("expected empty domain, got %#v", got)
	}
}

func TestParseDomainErrorMissingBracket(t *testing.T) {
	_, err := odoorpc.ParseDomain("[('name', '=', 'test')")
	if err == nil {
		t.Fatal("expected error for missing bracket")
	}
}

func TestParseDomainErrorInvalidStart(t *testing.T) {
	_, err := odoorpc.ParseDomain("('name', '=', 'test')")
	if err == nil {
		t.Fatal("expected error for invalid start")
	}
}

func TestParseDomainErrorUnknownKeyword(t *testing.T) {
	_, err := odoorpc.ParseDomain("[('name', '=', Invalid)]")
	if err == nil {
		t.Fatal("expected error for unknown keyword")
	}
}

func TestParseDomainAndConditions(t *testing.T) {
	// Explicit AND with two conditions
	got, err := odoorpc.ParseDomain("['&', ('name', '=', 'test'), ('active', '=', true)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"&",
		[]any{"name", "=", "test"},
		[]any{"active", "=", true},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainAndMultipleConditions(t *testing.T) {
	// Multiple AND: '&', '&', cond1, cond2, cond3
	got, err := odoorpc.ParseDomain("['&', '&', ('a', '=', 1), ('b', '=', 2), ('c', '=', 3)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"&",
		"&",
		[]any{"a", "=", int64(1)},
		[]any{"b", "=", int64(2)},
		[]any{"c", "=", int64(3)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainOrConditions(t *testing.T) {
	// OR with two conditions
	got, err := odoorpc.ParseDomain("['|', ('state', '=', 'draft'), ('state', '=', 'sent')]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		[]any{"state", "=", "draft"},
		[]any{"state", "=", "sent"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainOrMultipleConditions(t *testing.T) {
	// Multiple OR: '|', '|', cond1, cond2, cond3
	got, err := odoorpc.ParseDomain("['|', '|', ('x', '=', 1), ('y', '=', 2), ('z', '=', 3)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		"|",
		[]any{"x", "=", int64(1)},
		[]any{"y", "=", int64(2)},
		[]any{"z", "=", int64(3)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainAndOrCombined(t *testing.T) {
	// Combined: (a AND b) OR c => '|', '&', a, b, c
	got, err := odoorpc.ParseDomain("['|', '&', ('a', '=', 1), ('b', '=', 2), ('c', '=', 3)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		"&",
		[]any{"a", "=", int64(1)},
		[]any{"b", "=", int64(2)},
		[]any{"c", "=", int64(3)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainOrAndCombined(t *testing.T) {
	// Combined: a OR (b AND c) => '|', a, '&', b, c
	got, err := odoorpc.ParseDomain("['|', ('a', '=', 1), '&', ('b', '=', 2), ('c', '=', 3)]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		[]any{"a", "=", int64(1)},
		"&",
		[]any{"b", "=", int64(2)},
		[]any{"c", "=", int64(3)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainNotWithOr(t *testing.T) {
	// NOT combined with OR: '|', '!', cond1, cond2
	got, err := odoorpc.ParseDomain("['|', '!', ('active', '=', false), ('state', '=', 'done')]")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"|",
		"!",
		[]any{"active", "=", false},
		[]any{"state", "=", "done"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestParseDomainComplexAndOr(t *testing.T) {
	// Complex: (a OR b) AND (c OR d) => '&', '|', a, b, '|', c, d
	input := "['&', '|', ('a', '=', 1), ('b', '=', 2), '|', ('c', '=', 3), ('d', '=', 4)]"
	got, err := odoorpc.ParseDomain(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := odoorpc.Domain{
		"&",
		"|",
		[]any{"a", "=", int64(1)},
		[]any{"b", "=", int64(2)},
		"|",
		[]any{"c", "=", int64(3)},
		[]any{"d", "=", int64(4)},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}
