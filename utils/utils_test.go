package utils

import (
	"testing"
)

func TestFindIndex(t *testing.T) {
	tests := []struct {
		rawData    []byte
		startIndex int
		target     byte
		expected   int
	}{
		{[]byte{1, 2, 3, 4, 5}, 0, 3, 2},
		{[]byte{1, 2, 3, 4, 5}, 2, 4, 3},
		{[]byte{1, 2, 3, 4, 5}, 0, 6, -1},
		{[]byte{1, 2, 3, 4, 5}, 5, 1, -1},
		{[]byte{}, 0, 1, -1},
	}

	for _, test := range tests {
		result := FindIndex(test.rawData, test.startIndex, test.target)
		if result != test.expected {
			t.Errorf("FindIndex(%v, %d, %d) = %d; expected %d", test.rawData, test.startIndex, test.target, result, test.expected)
		}
	}
}
func TestPrepend(t *testing.T) {
	tests := []struct {
		rawData  []byte
		values   []byte
		expected []byte
	}{
		{[]byte{1, 2, 3}, []byte{0}, []byte{0, 1, 2, 3}},
		{[]byte{1, 2, 3}, []byte{4, 5}, []byte{4, 5, 1, 2, 3}},
		{[]byte{}, []byte{1, 2, 3}, []byte{1, 2, 3}},
		{[]byte{1, 2, 3}, []byte{}, []byte{1, 2, 3}},
	}

	for _, test := range tests {
		result := Prepend(test.rawData, test.values...)
		if !equal(result, test.expected) {
			t.Errorf("Prepend(%v, %v) = %v; expected %v", test.rawData, test.values, result, test.expected)
		}
	}
}

func TestAppend(t *testing.T) {
	tests := []struct {
		rawData  []byte
		values   []byte
		expected []byte
	}{
		{[]byte{1, 2, 3}, []byte{4}, []byte{1, 2, 3, 4}},
		{[]byte{1, 2, 3}, []byte{4, 5}, []byte{1, 2, 3, 4, 5}},
		{[]byte{}, []byte{1, 2, 3}, []byte{1, 2, 3}},
		{[]byte{1, 2, 3}, []byte{}, []byte{1, 2, 3}},
	}

	for _, test := range tests {
		result := Append(test.rawData, test.values...)
		if !equal(result, test.expected) {
			t.Errorf("Append(%v, %v) = %v; expected %v", test.rawData, test.values, result, test.expected)
		}
	}
}

func equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
