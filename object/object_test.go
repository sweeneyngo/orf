package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

type MockRepository struct{}

func (m *MockRepository) GetFile(directory string, create bool, paths ...string) (string, error) {
	return filepath.Join(directory, filepath.Join(paths...)), nil
}

func TestBaseObject(t *testing.T) {
	data := []byte("test data")
	base := &Base{
		format: "blob",
		size:   uint32(len(data)),
		data:   data,
	}

	if base.GetFormat() != "blob" {
		t.Errorf("Expected format 'blob', got %s", base.GetFormat())
	}

	if base.GetSize() != uint32(len(data)) {
		t.Errorf("Expected size %d, got %d", len(data), base.GetSize())
	}

	if !bytes.Equal(base.GetData(), data) {
		t.Errorf("Expected data %s, got %s", data, base.GetData())
	}
}

func TestReadObject(t *testing.T) {
	// Setup
	directory := t.TempDir()
	data := []byte("blob 0009\x00test data")
	hash := sha256.Sum256(data)
	hashHex := hex.EncodeToString(hash[:])
	hashDir := hashHex[:2]
	hashFile := hashHex[2:]

	// Create object file
	objectPath := filepath.Join(directory, "objects", hashDir, hashFile)
	if err := os.MkdirAll(filepath.Dir(objectPath), os.ModePerm); err != nil {
		t.Fatalf("Failed to create directories: %v", err)
	}

	var compressedData bytes.Buffer
	writer := zlib.NewWriter(&compressedData)
	if _, err := writer.Write(data); err != nil {
		t.Fatalf("Failed to write compressed data: %v", err)
	}
	writer.Close()

	if err := os.WriteFile(objectPath, compressedData.Bytes(), os.ModePerm); err != nil {
		t.Fatalf("Failed to write object file: %v", err)
	}

	// Test
	obj, err := ReadObject(directory, hashHex)
	if err != nil {
		t.Fatalf("ReadObject failed: %v", err)
	}

	if obj.GetFormat() != "blob" {
		t.Errorf("Expected format 'blob', got %s", obj.GetFormat())
	}

	if obj.GetSize() != 9 {
		t.Errorf("Expected size 9, got %d", obj.GetSize())
	}

	if !bytes.Equal(obj.GetData(), []byte("test data")) {
		t.Errorf("Expected data 'test data', got %s", obj.GetData())
	}
}

func TestWriteObject(t *testing.T) {
	// Setup
	directory := t.TempDir()
	data := []byte("test data")
	base := &Base{
		format: "blob",
		size:   uint32(len(data)),
		data:   data,
	}

	// Test
	hashHex, err := WriteObject(directory, base)
	if err != nil {
		t.Fatalf("WriteObject failed: %v", err)
	}

	hashDir := hashHex[:2]
	hashFile := hashHex[2:]
	objectPath := filepath.Join(directory, "objects", hashDir, hashFile)

	if _, err := os.Stat(objectPath); os.IsNotExist(err) {
		t.Fatalf("Expected object file to exist, but it does not")
	}

	// Verify content
	file, err := os.Open(objectPath)
	if err != nil {
		t.Fatalf("Failed to open object file: %v", err)
	}
	defer file.Close()

	zlibReader, err := zlib.NewReader(file)
	if err != nil {
		t.Fatalf("Failed to create zlib reader: %v", err)
	}
	defer zlibReader.Close()

	var rawData bytes.Buffer
	if _, err := io.Copy(&rawData, zlibReader); err != nil {
		t.Fatalf("Failed to read compressed data: %v", err)
	}

	expectedData := append([]byte("blob 0009\x00"), data...)
	if !bytes.Equal(rawData.Bytes(), expectedData) {
		t.Errorf("Expected data %s, got %s", expectedData, rawData.Bytes())
	}
}
