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

type Object struct {
	format string
	size   uint32
	Data   []byte
}

// ReadObject reads an object hash (SHA-256) from a .orf repository and returns an Object.
// The type of the returned Object depends on the object associated with the given hash.
func ReadObject(directory string, hash string) (*Object, error) {

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
		c := CreateBlob(data[dataIndex+1:])
		return &c.Object, nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectFormat)
	}
}

func WriteObject(directory string, object *Object) (string, error) {

	header := []byte(object.format + " ")
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, object.size)

	result := append(header, length...)
	result = append(result, '\x00')
	result = append(result, object.Data...)

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
