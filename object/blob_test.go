package object

import (
	"testing"
)

func TestCreateBlob(t *testing.T) {
	data := []byte("Hello, world!")
	blob := CreateBlob(data)

	if blob.GetFormat() != "blob" {
		t.Errorf("Expected format 'commit', got %s", blob.GetFormat())
	}

	if blob.GetSize() != uint32(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), blob.GetSize())
	}

	if string(blob.GetData()) != string(data) {
		t.Errorf("Expected data %s, got %s", data, blob.GetData())
	}
}
