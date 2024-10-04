package kv

import (
	"bytes"
	"os"
	"testing"
)

func TestOrderedMap_Get(t *testing.T) {
	orderedMap := &OrderedMap{
		data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		order: []string{"key1", "key2"},
	}

	value, exists := orderedMap.Get("key1")
	if !exists || value != "value1" {
		t.Errorf("Expected value1, got %v", value)
	}

	value, exists = orderedMap.Get("key3")
	if exists || value != nil {
		t.Errorf("Expected nil, got %v", value)
	}
}

func TestOrderedMap_GetData(t *testing.T) {
	orderedMap := &OrderedMap{
		data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		order: []string{"key1", "key2"},
	}

	data := orderedMap.GetData()
	if len(data) != 2 {
		t.Errorf("Expected 2 items, got %d", len(data))
	}
}

func TestOrderedMap_GetOrder(t *testing.T) {
	orderedMap := &OrderedMap{
		data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		order: []string{"key1", "key2"},
	}

	order := orderedMap.GetOrder()
	if len(order) != 2 || order[0] != "key1" || order[1] != "key2" {
		t.Errorf("Expected [key1, key2], got %v", order)
	}
}

func TestParse(t *testing.T) {
	rawData, err := os.ReadFile("test_data/commit.txt")
	if err != nil {
		t.Errorf("failed to read file: %v", err)
	}

	orderedMap, err := Parse(rawData, 0, nil)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(orderedMap.data) != 6 {
		t.Errorf("Expected 6 items, got %d", len(orderedMap.data))
	}

	expectedData := map[string]string{
		"tree":      "29ff16c9c14e2652b22f8b78bb08a5a07930c147",
		"parent":    "206941306e8a8af65b66eaaaea388a7ae24d49a0",
		"author":    "John Doe <johndoe@jd.oe> 1527025023 +0200",
		"committer": "John Doe <johndoe@jd.oe> 1527025044 +0200",
		"gpgsig": `-----BEGIN PGP SIGNATURE-----
 iQIzBAABCAAdFiEExwXquOM8bWb4Q2zVGxM2FxoLkGQFAlsEjZQACgkQGxM2FxoL
 kGQdcBAAqPP+ln4nGDd2gETXjvOpOxLzIMEw4A9gU6CzWzm+oB8mEIKyaH0UFIPh
 rNUZ1j7/ZGFNeBDtT55LPdPIQw4KKlcf6kC8MPWP3qSu3xHqx12C5zyai2duFZUU
 wqOt9iCFCscFQYqKs3xsHI+ncQb+PGjVZA8+jPw7nrPIkeSXQV2aZb1E68wa2YIL
 3eYgTUKz34cB6tAq9YwHnZpyPx8UJCZGkshpJmgtZ3mCbtQaO17LoihnqPn4UOMr
 V75R/7FjSuPLS8NaZF4wfi52btXMSxO/u7GuoJkzJscP3p4qtwe6Rl9dc1XC8P7k
 NIbGZ5Yg5cEPcfmhgXFOhQZkD0yxcJqBUcoFpnp2vu5XJl2E5I/quIyVxUXi6O6c
 /obspcvace4wy8uO0bdVhc4nJ+Rla4InVSJaUaBeiHTW8kReSFYyMmDCzLjGIu1q
 doU61OM3Zv1ptsLu3gUE6GU27iWYj2RWN3e3HE4Sbd89IFwLXNdSuM0ifDLZk7AQ
 WBhRhipCCgZhkj9g2NEk7jRVslti1NdN5zoQLaJNqSwO1MtxTmJ15Ksk3QP6kfLB
 Q52UWybBzpaP9HEd4XnR+HuQ4k2K0ns2KgNImsNvIyFwbpMUyUWLMPimaV1DWUXo
 5SBjDB/V/W2JBFR+XKHFJeFwYhj7DD/ocsGr4ZMx/lgc8rjIBkI=
 =lgTX
 -----END PGP SIGNATURE-----`,
	}

	for key, expectedValue := range expectedData {
		value, exists := orderedMap.Get(key)
		if !exists || value != expectedValue {
			t.Errorf("Expected %s for key %s, got %v", expectedValue, key, value)
		}
	}

	message, exists := orderedMap.Get("message")
	byteMessage, ok := message.([]byte)
	if !ok {
		t.Errorf("Expected []byte for key message, but got %T", message)
		return
	}
	expectedMessage := "Create first draft"
	if !exists || expectedMessage != string(byteMessage) {
		t.Errorf("Expected '%s' for key message, got '%v'", expectedMessage, string(byteMessage))
	}
}

func TestSerialize(t *testing.T) {
	orderedMap := &OrderedMap{
		data: map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
		},
		order: []string{"key1", "key2"},
	}

	expected := "key1 value1\nkey2 value2\n"
	output := Serialize(orderedMap)
	if !bytes.Equal(output, []byte(expected)) {
		t.Errorf("Expected %s, got %s", expected, string(output))
	}
}

func TestOrderedMap_Add(t *testing.T) {
	orderedMap := &OrderedMap{
		data:  make(map[string]interface{}),
		order: []string{},
	}

	orderedMap.Add("key1", "value1")
	if len(orderedMap.data) != 1 || orderedMap.data["key1"] != "value1" {
		t.Errorf("Unexpected data: %v", orderedMap.data)
	}

	orderedMap.Add("key2", "value2")
	if len(orderedMap.data) != 2 || orderedMap.data["key2"] != "value2" {
		t.Errorf("Unexpected data: %v", orderedMap.data)
	}

	orderedMap.Add("key1", "newValue1")
	if len(orderedMap.data) != 2 || orderedMap.data["key1"] != "newValue1" {
		t.Errorf("Unexpected data: %v", orderedMap.data)
	}
}
