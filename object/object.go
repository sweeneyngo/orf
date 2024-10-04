/*
Package object provides functionality for handling objects within a .orf repository.
Objects are fundamental units of storage in the repository, and they can be of various types such as blobs, commits, trees, and tags.

The Object interface defines the basic methods that any object type must implement:
- GetFormat() string: Returns the format/type of the object (e.g., "blob", "commit").
- GetSize() uint32: Returns the size of the object data.
- GetData() []byte: Returns the raw data of the object.

The Base struct is a concrete implementation of the Object interface and serves as a base type for other specific object types. It includes:
- format: A string representing the format/type of the object.
- size: A uint32 representing the size of the object data.
- data: A byte slice containing the raw data of the object.

Specific object types such as Commit, Blob, Tree, and Tag can embed the Base struct to inherit its fields and methods, while also adding their own specific fields and methods.

The ReadObject function reads an object from a .orf repository given its hash and returns an Object. The type of the returned Object depends on the format of the object associated with the given hash.

The WriteObject function writes an Object to a .orf repository and returns the SHA-256 hash of the written object. If the directory is not specified, it returns the hash without writing the object to the repository.
*/
package object

import (
	"bytes"
	"compress/zlib"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"orf/repository"
	"os"
	"path/filepath"
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

	path, err := repository.GetFile(directory, false, "objects", hashDir, hashFile)
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

	size := int(binary.BigEndian.Uint32(data[startSizeIndex:endSizeIndex]))

	// Get object data
	dataIndex := bytes.IndexByte(data[endSizeIndex:], '\x00')
	if dataIndex == -1 {
		return nil, fmt.Errorf("invalid data format index")
	}

	// Convert data index to absolute index
	dataIndex = endSizeIndex + dataIndex

	if size != len(data)-(dataIndex+1) {
		return nil, fmt.Errorf("object size mismatch")
	}

	// Create object based on format
	switch objectFormat {
	case "blob":
		return CreateBlob(data[dataIndex+1:]), nil
	case "commit":
		return CreateCommit(data[dataIndex+1:]), nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectFormat)
	}
}

func WriteObject(directory string, object Object) (string, error) {

	header := []byte(object.GetFormat() + " ")
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, object.GetSize())

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
