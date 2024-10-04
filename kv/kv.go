package kv

import (
	"bytes"
	"fmt"
	"strings"
)

type OrderedMap struct {
	data  map[string]interface{}
	order []string // Maintain order of key-value insertion
}

// Retrieves the value associated with the given key.
// It returns the value and a boolean indicating whether the key was found.
func (orderedMap *OrderedMap) Get(key string) (interface{}, bool) {
	value, exists := orderedMap.data[key]
	return value, exists
}

// Returns the data, regardless of order.
func (orderedMap *OrderedMap) GetData() map[string]interface{} {
	return orderedMap.data
}

// Returns the order of keys as they were inserted.
func (orderedMap *OrderedMap) GetOrder() []string {
	return orderedMap.order
}

// Parse takes a raw byte slice and a starting index, and parses it into a map of key-value pairs.
// The function handles continuation lines and ensures that values are correctly associated with their keys.
// If a key already exists in the map, the function appends the new value to a list of values for that key.
// If the map is nil, it initializes a new map before parsing.
func Parse(rawData []byte, startIndex int, dct *OrderedMap) (*OrderedMap, error) {

	if dct == nil {
		dct = &OrderedMap{
			data:  make(map[string]interface{}),
			order: []string{},
		}
	}

	spaceIndex := bytes.IndexByte(rawData[startIndex:], ' ')
	newLineIndex := bytes.IndexByte(rawData[startIndex:], '\n')

	if spaceIndex < 0 || newLineIndex < 0 || newLineIndex < spaceIndex {
		// Found a message (end of key-value pairs)
		if newLineIndex == startIndex {
			// Store message after the new line
			dct.Add("message", rawData[startIndex+1:])
			return dct, nil
		}
	}

	// Base case: found the key-value pair
	key := string(rawData[startIndex : startIndex+spaceIndex])

	// Move the end pointer to the value, handling continuation lines
	endIndex := startIndex
	for {
		endIndex = endIndex + bytes.IndexByte(rawData[endIndex:], '\n')
		// Check if it's the end of continuation line
		if rawData[endIndex+1] != ' ' {
			break
		}
	}

	value := string(bytes.TrimSpace(rawData[spaceIndex+1 : endIndex]))

	// Check if the key already exists in the map
	if existingValue, exists := dct.data[key]; exists {
		// If the existing value is a list, append; otherwise, make it a list
		switch v := existingValue.(type) {
		case []string:
			dct.data[key] = append(v, value)
		default:
			dct.data[key] = []string{fmt.Sprintf("%v", v), value}
		}
	} else {
		dct.Add(key, value)
	}

	// Find next key-value pair (advance to the next line)
	return Parse(rawData, endIndex+1, dct)
}

func Serialize(orderedMap *OrderedMap) []byte {

	var output []byte
	order := orderedMap.order
	data := orderedMap.data

	// Output fields
	for _, key := range order {
		// Skip the message (not key-value)
		if key == "message" {
			continue
		}

		// Normalize the value to a slice
		var values []string
		switch ty := data[key].(type) {
		case string:
			values = []string{ty}
		case []string:
			values = ty
		default:
			continue
		}

		for _, value := range values {
			value = strings.ReplaceAll(value, "\n", "\n ")
			output = append(output, []byte(fmt.Sprintf("%s %s\n", key, value))...)
		}
	}

	if msg, exists := data["message"]; exists {
		output = append(output, []byte("\n")...)
		output = append(output, []byte(fmt.Sprintf("%s", msg))...)
		output = append(output, []byte("\n")...)
	}

	return output
}

func (orderedMap *OrderedMap) Add(key string, value interface{}) {
	if _, exists := orderedMap.data[key]; !exists {
		orderedMap.order = append(orderedMap.order, key)
	}
	orderedMap.data[key] = value
}
