package object

import (
	"bytes"
	"encoding/hex"
	"sort"
	"testing"
)

func TestCreateTree(t *testing.T) {
	data := []byte("test data")
	tree := CreateTree(data)

	if tree.GetFormat() != "tree" {
		t.Errorf("Expected format 'tree', got %s", tree.GetFormat())
	}

	if tree.GetSize() != uint32(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), tree.GetSize())
	}

	if !bytes.Equal(tree.GetData(), data) {
		t.Errorf("Expected data %s, got %s", data, tree.GetData())
	}

	if len(tree.GetLeaves()) != 0 {
		t.Errorf("Expected 0 leaves, got %d", len(tree.GetLeaves()))
	}
}

func TestTreeSerialization(t *testing.T) {
	tree := CreateTree([]byte{})
	leaf := &Leaf{
		mode: []byte("100644"),
		path: "file.txt",
		hash: "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391",
	}
	tree.leaves = append(tree.leaves, leaf)

	serialized, err := tree.Serialize()
	if err != nil {
		t.Fatalf("Serialization failed: %v", err)
	}

	expected := append([]byte("100644 file.txt\x00"), hexToBytes(t, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")...)
	if !bytes.Equal(serialized, expected) {
		t.Errorf("Expected serialized data %x, got %x", expected, serialized)
	}
}

func TestTreeDeserialization(t *testing.T) {
	data := append([]byte("100644 file.txt\x00"), hexToBytes(t, "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391")...)
	tree := CreateTree([]byte{})

	err := tree.Deserialize(data)
	if err != nil {
		t.Fatalf("Deserialization failed: %v", err)
	}

	if len(tree.GetLeaves()) != 1 {
		t.Fatalf("Expected 1 leaf, got %d", len(tree.GetLeaves()))
	}

	leaf := tree.GetLeaves()[0]
	if string(leaf.GetMode()) != "100644" {
		t.Errorf("Expected mode '100644', got %s", leaf.GetMode())
	}

	if leaf.GetPath() != "file.txt" {
		t.Errorf("Expected path 'file.txt', got %s", leaf.GetPath())
	}

	if leaf.GetHash() != "e69de29bb2d1d6434b8b29ae775ad8c2e48c5391" {
		t.Errorf("Expected hash 'e69de29bb2d1d6434b8b29ae775ad8c2e48c5391', got %s", leaf.GetHash())
	}
}

func TestByPathSorting(t *testing.T) {
	leaves := []*Leaf{
		{mode: []byte("100644"), path: "b.txt"},
		{mode: []byte("100644"), path: "a.txt"},
		{mode: []byte("40000"), path: "dir"},
	}

	sort.Sort(ByPath(leaves))

	expectedPaths := []string{"a.txt", "b.txt", "dir"}
	for i, leaf := range leaves {
		if leaf.GetPath() != expectedPaths[i] {
			t.Errorf("Expected path %s, got %s", expectedPaths[i], leaf.GetPath())
		}
	}
}

func hexToBytes(t *testing.T, hexStr string) []byte {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		t.Fatalf("Failed to decode hex string: %v", err)
	}
	return bytes
}
