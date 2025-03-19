package list

import "testing"

func TestParseWhere(t *testing.T) {

	// Test cases
	tests := []struct {
		name  string
		input string
		key   string
		op    string
		value string
	}{
		{"Case 1", "name = 'John'", "name", "=", "John"},
		{"Case 1b", "name == 'John'", "name", "=", "John"},
		{"Case 2", "name != 'John'", "name", "!=", "John"},
		{"Case 3", "name > 'John'", "name", ">", "John"},
		{"Case 4", "name < 'John'", "name", "<", "John"},
		{"Case 5", "name >= 'John'", "name", ">=", "John"},
		{"Case 6", "name <= 'John'", "name", "<=", "John"},
		{"Case 7", "name like 'John'", "name", "like", "John"},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := parseWhere(test.input)
			op, ok := exp.(binOp)
			if !ok {
				t.Errorf("Expected binOp.")
			}
			if tk, ok := op.lv.(token); !ok || tk.text != test.key {
				t.Errorf("Expected %v , but got %v", test.key, tk.text)
			}
			if tk := op.kind; tk.String() != test.op {
				t.Errorf("Expected %v , but got %s", test.op, tk)
			}
			if tk, ok := op.rv.(token); !ok || tk.text != test.value {
				t.Errorf("Expected %v , but got %v", test.value, tk.text)
			}
		})
	}
}

func TestParse(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		input    string
		expected expression
	}{
		{"Case 1",
			"(Name = 'A') OR (Name = 'B')",
			binOp{
				kind: k_or,
				lv: binOp{
					kind: k_eq,
					lv: token{
						kind: k_field,
						text: "Name",
					},
					rv: token{
						kind: k_string,
						text: "A",
					},
				},
				rv: binOp{
					kind: k_eq,
					lv: token{
						kind: k_field,
						text: "Name",
					},
					rv: token{
						kind: k_string,
						text: "B",
					},
				},
			},
		},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exp := parseWhere(test.input)
			op, ok := exp.(binOp)
			if !ok {
				t.Errorf("Expected binOp.")
			}
			_, ok = op.lv.(binOp)
			if !ok {
				t.Errorf("Expected binOp.")
			}
			_, ok = op.rv.(binOp)
			if !ok {
				t.Errorf("Expected binOp.")
			}

		})
	}
}

func TestParseList(t *testing.T) {
	input := "('A','B','C','D')"
	expect := []expression{
		token{kind: k_string, text: "A"},
		token{kind: k_string, text: "B"},
		token{kind: k_string, text: "C"},
		token{kind: k_string, text: "D"},
	}
	tks := scan(input)
	n, result := parseList(tks)
	if n != len(tks) {
		t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
	}
	if len(result) != len(expect) {
		t.Logf("%v.", result)
		t.Errorf("expected %d items got %d.", len(expect), len(result))
	}
}

func TestParseBetween(t *testing.T) {
	input := "X BETWEEN 'A' AND 'B'"
	expect := betweenOp{
		kind:  k_between,
		test:  token{kind: k_field, text: "X"},
		begin: token{kind: k_string, text: "A"},
		end:   token{kind: k_string, text: "B"},
	}
	tks := scan(input)
	n, result := parse(tks)
	if n != len(tks) {
		t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
	}
	rop, ok := result.(betweenOp)
	if !ok {
		t.Errorf("expected between got %#v", result)
	}
	if expect.String() != rop.String() {
		t.Errorf("expected %q items got %q.", expect, rop)
	}
}

func TestParseBetweenAnd(t *testing.T) {
	input := "X BETWEEN 'A' AND 'B' AND Y < 'C'"
	expect := binOp{
		kind: k_and,
		lv: betweenOp{
			kind:  k_between,
			test:  token{kind: k_field, text: "X"},
			begin: token{kind: k_string, text: "A"},
			end:   token{kind: k_string, text: "B"},
		},
		rv: binOp{
			kind: k_lt,
			lv:   token{kind: k_field, text: "Y"},
			rv:   token{kind: k_string, text: "C"},
		},
	}
	tks := scan(input)
	n, result := parse(tks)
	if n != len(tks) {
		t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
	}
	resultOp, ok := result.(binOp)
	if !ok {
		t.Errorf("expected between got %#v", result)
	}
	if expect.String() != resultOp.String() {
		t.Errorf("expected %q items got %q.", expect, resultOp)
	}
	resultL := resultOp.lv.(betweenOp)
	expectL := expect.lv.(betweenOp)
	if expectL.String() != resultL.String() {
		t.Errorf("expected %q items got %q.", expectL, resultL)
	}
}

func TestParseAndAnd(t *testing.T) {
	input := "X = 'A' AND Y > 'B' AND Z < 'C'"
	expect := binOp{
		kind: k_and,
		lv: binOp{
			kind: k_eq,
			lv:   token{kind: k_field, text: "X"},
			rv:   token{kind: k_string, text: "A"},
		},
		rv: binOp{
			kind: k_and,
			lv: binOp{
				kind: k_gt,
				lv:   token{kind: k_field, text: "Y"},
				rv:   token{kind: k_string, text: "B"},
			},
			rv: binOp{
				kind: k_lt,
				lv:   token{kind: k_field, text: "Z"},
				rv:   token{kind: k_string, text: "C"},
			},
		},
	}
	tks := scan(input)
	n, result := parse(tks)
	if n != len(tks) {
		t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
	}
	resultOp, ok := result.(binOp)
	if !ok {
		t.Errorf("expected binOp got %#v", result)
	}
	if expect.String() != resultOp.String() {
		t.Errorf("expected %q items got %q.", expect, resultOp)
	}
	resultR, ok1 := resultOp.rv.(binOp)
	expectR, ok2 := expect.rv.(binOp)
	if !ok1 || !ok2 || expectR.String() != resultR.String() {
		t.Errorf("expected %q items got %q.", expectR, resultR)
	}
}

func TestParseList2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []expression
	}{
		{
			name:     "Empty list",
			input:    "()",
			expected: []expression{},
		},
		{
			name:  "Single item list",
			input: "('A')",
			expected: []expression{
				token{kind: k_string, text: "A"},
			},
		},
		{
			name:  "Multiple items list",
			input: "('A', 'B', 'C')",
			expected: []expression{
				token{kind: k_string, text: "A"},
				token{kind: k_string, text: "B"},
				token{kind: k_string, text: "C"},
			},
		},
		{
			name:  "List with nested expressions",
			input: "((1 AND 2), 'B', 'C')",
			expected: []expression{
				binOp{
					kind: k_and,
					lv:   token{kind: k_integer, text: "1"},
					rv:   token{kind: k_integer, text: "2"},
				},
				token{kind: k_string, text: "B"},
				token{kind: k_string, text: "C"},
			},
		},
		{
			name:  "List with trailing comma",
			input: "('A', 'B',)",
			expected: []expression{
				token{kind: k_string, text: "A"},
				token{kind: k_string, text: "B"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tks := scan(test.input)
			n, result := parseList(tks)
			if n != len(tks) {
				t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
			}
			if len(result) != len(test.expected) {
				t.Errorf("expected %d items, got %d.", len(test.expected), len(result))
			}
			for i, exp := range test.expected {
				if result[i].String() != exp.String() {
					t.Errorf("expected %q, got %q at index %d.", exp, result[i], i)
				}
			}
		})
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected expression
	}{
		{
			name:  "Simple equality",
			input: "X = 'A'",
			expected: binOp{
				kind: k_eq,
				lv:   token{kind: k_field, text: "X"},
				rv:   token{kind: k_string, text: "A"},
			},
		},
		{
			name:  "Simple inequality",
			input: "X != 'A'",
			expected: binOp{
				kind: k_ne,
				lv:   token{kind: k_field, text: "X"},
				rv:   token{kind: k_string, text: "A"},
			},
		},
		{
			name:  "Parentheses grouping",
			input: "(X = 'A')",
			expected: binOp{
				kind: k_eq,
				lv:   token{kind: k_field, text: "X"},
				rv:   token{kind: k_string, text: "A"},
			},
		},
		{
			name:  "Logical AND",
			input: "X = 'A' AND Y = 'B'",
			expected: binOp{
				kind: k_and,
				lv: binOp{
					kind: k_eq,
					lv:   token{kind: k_field, text: "X"},
					rv:   token{kind: k_string, text: "A"},
				},
				rv: binOp{
					kind: k_eq,
					lv:   token{kind: k_field, text: "Y"},
					rv:   token{kind: k_string, text: "B"},
				},
			},
		},
		{
			name:  "Logical OR",
			input: "X = 'A' OR Y = 'B'",
			expected: binOp{
				kind: k_or,
				lv: binOp{
					kind: k_eq,
					lv:   token{kind: k_field, text: "X"},
					rv:   token{kind: k_string, text: "A"},
				},
				rv: binOp{
					kind: k_eq,
					lv:   token{kind: k_field, text: "Y"},
					rv:   token{kind: k_string, text: "B"},
				},
			},
		},
		{
			name:  "IN operator",
			input: "X IN ('A', 'B', 'C')",
			expected: inOp{
				lv: token{kind: k_field, text: "X"},
				items: []expression{
					token{kind: k_string, text: "A"},
					token{kind: k_string, text: "B"},
					token{kind: k_string, text: "C"},
				},
			},
		},
		{
			name:  "BETWEEN operator",
			input: "X BETWEEN 'A' AND 'B'",
			expected: betweenOp{
				kind:  k_between,
				test:  token{kind: k_field, text: "X"},
				begin: token{kind: k_string, text: "A"},
				end:   token{kind: k_string, text: "B"},
			},
		},
		{
			name:  "LIKE operator",
			input: "X LIKE 'pattern%'",
			expected: binOp{
				kind: k_like,
				lv:   token{kind: k_field, text: "X"},
				rv:   token{kind: k_string, text: "pattern%"},
			},
		},
		{
			name:  "ISNULL operator",
			input: "X ISNULL",
			expected: uniOp{
				kind: k_is_null,
				rv:   token{kind: k_field, text: "X"},
			},
		},
		{
			name:  "ISNOTNULL operator",
			input: "X ISNOTNULL",
			expected: uniOp{
				kind: k_is_not_null,
				rv:   token{kind: k_field, text: "X"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tks := scan(test.input)
			n, result := parse(tks)
			if n != len(tks) {
				t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
			}
			if result.String() != test.expected.String() {
				t.Errorf("expected %q, got %q.", test.expected, result)
			}
		})
	}
}
