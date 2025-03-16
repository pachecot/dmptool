package list

import "testing"

func TestScan(t *testing.T) {

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
		{"Case 10", " X BETWEEN A AND B ", []token{
			{kind: k_text, text: "X"},
			{kind: k_between, text: ""},
			{kind: k_text, text: "A"},
			{kind: k_and, text: ""},
			{kind: k_text, text: "B"},
		}},
		{"Case 11", " X BETWEEN A AND B AND Y < C", []token{
			{kind: k_text, text: "X"},
			{kind: k_between, text: ""},
			{kind: k_text, text: "A"},
			{kind: k_and, text: ""},
			{kind: k_text, text: "B"},
			{kind: k_and, text: ""},
			{kind: k_text, text: "Y"},
			{kind: k_lt, text: "<"},
			{kind: k_text, text: "C"},
		}},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tks := scan(test.input)
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

func TestScanNumbers(t *testing.T) {

	// Test cases
	tests := []struct {
		name  string
		input string
		kinds []kind
	}{
		{"Case 1", "123", []kind{k_integer}},
		{"Case 2", "123.0", []kind{k_decimal}},
		{"Case 3", "123e123", []kind{k_decimal}},
		{"Case 4", "123.013e13", []kind{k_decimal}},
		{"Case 5", "123.013e13>123", []kind{k_decimal, k_gt, k_integer}},
		{"Case 6", "123.013e13AAA", []kind{k_text}},
		{"Case 7", "123 AAA", []kind{k_integer, k_text}},
	}

	// Run tests
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tks := scan(test.input)
			if len(tks) != len(test.kinds) {
				t.Errorf("Expected single token but got %d :: %v", len(tks), tks)
			}
			for i := range test.kinds {
				if tks[i].kind != test.kinds[i] {
					t.Errorf("test %d: expected %v , but got %v", i, test.kinds[i], tks[0].kind)
				}
			}
		})
	}
}
