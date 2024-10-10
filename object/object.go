package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"orf/repository"
	"os"
	"path/filepath"
	"strconv"
)

// Defines the basic methods that any object type must implement (Blob, Commit, Tag, Tree).
type Object interface {
	GetFormat() string
	GetSize() uint32
	GetData() []byte
}

// Simplest form of Object, all objects derive from Base.
type Base struct {
	format string
	size   uint32
	data   []byte
}

func (base *Base) GetFormat() string {
	return base.format
}

func (base *Base) GetSize() uint32 {
	return base.size
}

func (base *Base) GetData() []byte {
	return base.data
}

// ReadObject reads an object hash (SHA-256) from a .orf repository and returns an Object.
// The type of the returned Object depends on the object associated with the given hash.
func ReadObject(directory string, hash string) (Object, error) {

	hashDir := hash[0:2]
	hashFile := hash[2:]

	path, err := repository.GetFilePath(directory, false, "objects", hashDir, hashFile)
	if err != nil {
		return nil, err
	}

	// Open file at path
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decompress file with zlib
	zlibReader, err := zlib.NewReader(file)
	if err != nil {
		return nil, err
	}
	defer zlibReader.Close()

	var rawData bytes.Buffer
	_, err = io.Copy(&rawData, zlibReader)
	if err != nil {
		return nil, err
	}

	// Get object format
	data := rawData.Bytes()
	formatIndex := bytes.IndexByte(data, ' ')
	if formatIndex == -1 {
		return nil, fmt.Errorf("invalid object format index")
	}
	objectFormat := string(data[:formatIndex])

	// Get object size
	startSizeIndex := formatIndex + 1
	endSizeIndex := formatIndex + 5

	size, err := strconv.Atoi(string(data[startSizeIndex:endSizeIndex]))
	if err != nil {
		return nil, fmt.Errorf("invalid size format: %v", err)
	}

	// Get object data
	dataIndex := bytes.IndexByte(data[endSizeIndex:], '\x00')
	if dataIndex == -1 {
		return nil, fmt.Errorf("invalid data format index")
	}

	// Convert data index to absolute index
	dataIndex = endSizeIndex + dataIndex

	fmt.Println(data, size, dataIndex)
	if size != len(data)-(dataIndex+1) {
		return nil, fmt.Errorf("object size mismatch")
	}

	// Create object based on format
	switch objectFormat {
	case "blob":
		return CreateBlob(data[dataIndex+1:]), nil
	case "commit":
		return CreateCommit(data[dataIndex+1:]), nil
	case "tree":
		return CreateTree(data[dataIndex+1:]), nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectFormat)
	}
}

// Writes an Object to a .orf repository and returns the SHA-256 hash of the written object.
// If the directory is not specified, it returns the hash without writing the object to the repository.
func WriteObject(directory string, object Object) (string, error) {

	header := []byte(object.GetFormat() + " ")
	length := fmt.Sprintf("%04d", object.GetSize())

	result := append(header, length...)
	result = append(result, '\x00')
	result = append(result, object.GetData()...)

	sha := sha256.Sum256(result)
	shaHex := hex.EncodeToString(sha[:])

	if directory == "" {
		// Return hex if no .orf path specified
		return shaHex, nil
	}

	dirPath := filepath.Join(directory, "objects", shaHex[:2])
	filePath := filepath.Join(directory, "objects", shaHex[:2], shaHex[2:])

	if err := os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return "", fmt.Errorf("no directory with path %v found: %w", dirPath, err)
	}

	// Check if file already exists
	if _, err := os.Stat(filePath); err == nil {
		return shaHex, nil
	}

	file, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("fail to create file: %w", err)
	}
	defer file.Close()

	// Compress to zlib
	writer := zlib.NewWriter(file)
	defer writer.Close()

	if _, err := writer.Write(result); err != nil {
		return "", fmt.Errorf("fail to compress file: %w", err)
	}

	return shaHex, nil
}

func FindObject(repo *repository.Repo, name string, format string, follow bool) string {
	return name
}
