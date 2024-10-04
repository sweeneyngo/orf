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
