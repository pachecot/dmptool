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
			if tk, ok := op.lv.(token); !ok || tk.p != test.key {
				t.Errorf("Expected %v , but got %v", test.key, tk.p)
			}
			if tk := op.kind; tk.String() != test.op {
				t.Errorf("Expected %v , but got %s", test.op, tk)
			}
			if tk, ok := op.rv.(token); !ok || tk.p != test.value {
				t.Errorf("Expected %v , but got %v", test.value, tk.p)
			}
		})
	}

}
