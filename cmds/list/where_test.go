package list

import "testing"

func TestIsLike(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		input    string
		match    string
		expected bool
	}{
		{"Case 1", "xxx", "%", true},
		{"Case 2", "xxxABC", "%abc", true},
		{"Case 3", "ABCxxx", "abc%", true},
		{"Case 4", "xxxABCxxx", "%abc%", true},
		{"Case 5", "xxx", "abc", false},
		{"Case 6", "ABCxxx", "%abc", false},
		{"Case 7", "xxxABC", "abc%", false},
		{"Case 8", "xxxABC", "abc", false},
		{"Case 9", "ABCDEF", "abc%def", true},
		{"Case 10", "ABCxxxDEF", "abc%def", true},
		{"Case 11", "xxABCxxxDEFxx", "%abc%def%", true},
		{"Case 12", "xxABxCxxxDEFxx", "%abc%def%", false},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := isLike(test.input, test.match)
			if result != test.expected {
				t.Errorf("Expected %v, but got %v", test.expected, result)
			}
		})
	}
}

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

func TestTokenize(t *testing.T) {

	// Test cases
	tests := []struct {
		name     string
		input    string
		expected []token
	}{
		{"Case 1", "name = John", []token{{kind: k_text, text: "name"}, {kind: k_eq, text: "="}, {kind: k_text, text: "John"}}},
		{"Case 1b", "name == John", []token{{kind: k_text, text: "name"}, {kind: k_eq, text: "=="}, {kind: k_text, text: "John"}}},
		{"Case 2", "name != John", []token{{kind: k_text, text: "name"}, {kind: k_ne, text: "!="}, {kind: k_text, text: "John"}}},
		{"Case 3", "name > John", []token{{kind: k_text, text: "name"}, {kind: k_gt, text: ">"}, {kind: k_text, text: "John"}}},
		{"Case 4", "name < John", []token{{kind: k_text, text: "name"}, {kind: k_lt, text: "<"}, {kind: k_text, text: "John"}}},
		{"Case 5", "name >= John", []token{{kind: k_text, text: "name"}, {kind: k_ge, text: ">="}, {kind: k_text, text: "John"}}},
		{"Case 6", "name <= John", []token{{kind: k_text, text: "name"}, {kind: k_le, text: "<="}, {kind: k_text, text: "John"}}},
		{"Case 7", "name like John", []token{{kind: k_text, text: "name"}, {kind: k_like, text: ""}, {kind: k_text, text: "John"}}},
		{"Case 8", "(A,B,C)", []token{
			{kind: k_paren_left, text: ""},
			{kind: k_text, text: "A"},
			{kind: k_comma, text: ""},
			{kind: k_text, text: "B"},
			{kind: k_comma, text: ""},
			{kind: k_text, text: "C"},
			{kind: k_paren_right, text: ""},
		}},
		{"Case 9", " X IN ( A, B, C ) ", []token{
			{kind: k_text, text: "X"},
			{kind: k_in, text: ""},
			{kind: k_paren_left, text: ""},
			{kind: k_text, text: "A"},
			{kind: k_comma, text: ""},
			{kind: k_text, text: "B"},
			{kind: k_comma, text: ""},
			{kind: k_text, text: "C"},
			{kind: k_paren_right, text: ""},
		}},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tks := tokenize(test.input)
			if len(test.expected) != len(tks) {
				t.Errorf("expected %d got %d", len(test.expected), len(tks))
			}
			for i := range test.expected {
				if test.expected[i].kind != tks[i].kind {
					t.Errorf("item %d Expected %v but got %v", i, test.expected[i].kind, tks[i].kind)
				}
				if test.expected[i].text != tks[i].text {
					t.Errorf("item %d Expected %v but got %v", i, test.expected[i].text, tks[i].text)
				}
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
	tks := tokenize(input)
	n, result := parseList(tks)
	if n != len(tks) {
		t.Errorf("did not read entire input, read %d of %d.", n, len(tks))
	}
	if len(result) != len(expect) {
		t.Logf("%v.", result)
		t.Errorf("expected %d items got %d.", len(expect), len(result))
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
