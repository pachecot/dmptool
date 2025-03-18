package list

import (
	"reflect"
	"testing"
)

func TestPartitionDigits(t *testing.T) {
	tests := []struct {
		input    string
		expected []string
	}{
		{"abc123def456", []string{"abc", "123", "def", "456"}},
		{"123abc456", []string{"123", "abc", "456"}},
		{"abc123", []string{"abc", "123"}},
		{"123", []string{"123"}},
		{"abc", []string{"abc"}},
		{"", []string{""}},
		{"123abc456def789", []string{"123", "abc", "456", "def", "789"}},
		{"abc123def", []string{"abc", "123", "def"}},
		{"123456", []string{"123456"}},
		{"abcDEF123", []string{"abcDEF", "123"}},
	}

	for _, test := range tests {
		result := partitionDigits(test.input)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("partitionDigits(%q) = %v; want %v", test.input, result, test.expected)
		}
	}
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		order    ordering
		a        []string
		b        []string
		expected int
	}{
		{
			name:     "Equal strings",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc"},
			b:        []string{"abc"},
			expected: 0,
		},
		{
			name:     "Ascending order, a < b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc"},
			b:        []string{"def"},
			expected: -1,
		},
		{
			name:     "Ascending order, a > b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"def"},
			b:        []string{"abc"},
			expected: 1,
		},
		{
			name:     "Descending order, a < b",
			order:    ordering{col: 0, desc: true},
			a:        []string{"abc"},
			b:        []string{"def"},
			expected: 1,
		},
		{
			name:     "Descending order, a > b",
			order:    ordering{col: 0, desc: true},
			a:        []string{"def"},
			b:        []string{"abc"},
			expected: -1,
		},
		{
			name:     "Numeric comparison, a < b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"123"},
			b:        []string{"456"},
			expected: -1,
		},
		{
			name:     "Numeric comparison, a > b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"456"},
			b:        []string{"123"},
			expected: 1,
		},
		{
			name:     "Mixed alphanumeric comparison, a < b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc123"},
			b:        []string{"abc456"},
			expected: -1,
		},
		{
			name:     "Mixed alphanumeric comparison, a > b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc456"},
			b:        []string{"abc123"},
			expected: 1,
		},
		{
			name:     "Different lengths, a < b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc"},
			b:        []string{"abc123"},
			expected: -1,
		},
		{
			name:     "Different lengths, a > b",
			order:    ordering{col: 0, desc: false},
			a:        []string{"abc123"},
			b:        []string{"abc"},
			expected: 1,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.order.compare(test.a, test.b)
			if result != test.expected {
				t.Errorf("compare(%v, %v) = %d; want %d", test.a, test.b, result, test.expected)
			}
		})
	}
}

func TestReorder(t *testing.T) {
	tests := []struct {
		name      string
		orders    []string
		fields    []string
		table     [][]string
		expected  [][]string
		expectErr error
	}{
		{
			name:      "Valid reorder ascending",
			orders:    []string{"acs name"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expected:  [][]string{{"2", "Alice"}, {"1", "Bob"}, {"3", "Charlie"}},
			expectErr: nil,
		},
		{
			name:      "Valid reorder descending",
			orders:    []string{"desc name"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expected:  [][]string{{"3", "Charlie"}, {"1", "Bob"}, {"2", "Alice"}},
			expectErr: nil,
		},
		{
			name:      "Unknown field",
			orders:    []string{"acs unknown"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expected:  [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expectErr: errOrderbyUnknownField,
		},
		{
			name:      "Unrecognized order direction",
			orders:    []string{"unknown name"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expected:  [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expectErr: errOrderbyUnrecognizedOrder,
		},
		{
			name:      "Parse failed",
			orders:    []string{"acs name extra"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expected:  [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Charlie"}},
			expectErr: errOrderbyParseFailed,
		},
		{
			name:      "Multiple orders",
			orders:    []string{"acs name", "desc id"},
			fields:    []string{"id", "name"},
			table:     [][]string{{"1", "Bob"}, {"2", "Alice"}, {"3", "Alice"}},
			expected:  [][]string{{"3", "Alice"}, {"2", "Alice"}, {"1", "Bob"}},
			expectErr: nil,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			tableCopy := make([][]string, len(test.table))
			for i := range test.table {
				tableCopy[i] = append([]string{}, test.table[i]...)
			}

			err := reorder(test.orders, test.fields, tableCopy)
			if err != test.expectErr {
				t.Errorf("reorder() error = %v; want %v", err, test.expectErr)
				return
			}

			if !reflect.DeepEqual(tableCopy, test.expected) {
				t.Errorf("reorder() = %v; want %v", tableCopy, test.expected)
			}
		})
	}
}
