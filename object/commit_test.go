package object

import (
	"testing"
)

func TestCreateCommit(t *testing.T) {
	data := []byte("Initial commit")
	commit := CreateCommit(data)

	if commit.GetFormat() != "commit" {
		t.Errorf("Expected format 'commit', got %s", commit.GetFormat())
	}

	if commit.GetSize() != uint32(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), commit.GetSize())
	}

	if string(commit.GetData()) != string(data) {
		t.Errorf("Expected data %s, got %s", data, commit.GetData())
	}

	if commit.GetKVData() != nil {
		t.Errorf("Expected kvData to be nil, got %v", commit.GetKVData())
	}
}
