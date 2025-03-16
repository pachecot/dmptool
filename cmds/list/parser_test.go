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
		{"Case 1", "name = John", "name", "=", "John"},
		{"Case 1b", "name == John", "name", "=", "John"},
		{"Case 2", "name != John", "name", "!=", "John"},
		{"Case 3", "name > John", "name", ">", "John"},
		{"Case 4", "name < John", "name", "<", "John"},
		{"Case 5", "name >= John", "name", ">=", "John"},
		{"Case 6", "name <= John", "name", "<=", "John"},
		{"Case 7", "name like John", "name", "like", "John"},
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
			"(Name = A) OR (Name = B)",
			binOp{
				kind: k_or,
				lv: binOp{
					kind: k_eq,
					lv: token{
						kind: k_text,
						text: "Name",
					},
					rv: token{
						kind: k_text,
						text: "A",
					},
				},
				rv: binOp{
					kind: k_eq,
					lv: token{
						kind: k_text,
						text: "Name",
					},
					rv: token{
						kind: k_text,
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
	input := "(A,B,C,D)"
	expect := []expression{
		token{kind: k_text, text: "A"},
		token{kind: k_text, text: "B"},
		token{kind: k_text, text: "C"},
		token{kind: k_text, text: "D"},
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
