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
