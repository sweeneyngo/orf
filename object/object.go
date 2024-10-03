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
	"strconv"
)

type Object struct {
	format string
	size   uint32
	Data   []byte
}

// ReadObject reads an object hash (SHA-256) from a .orf repository and returns an Object.
// The type of the returned Object depends on the object associated with the given hash.
func ReadObject(workTree string, hash string) (*Object, error) {

	hashDir := hash[0:2]
	hashFile := hash[2:]

	path, err := repository.GetFile(workTree, false, "objects", hashDir, hashFile)
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
	sizeIndex := bytes.IndexByte(data, '\x00')
	if sizeIndex == -1 {
		return nil, fmt.Errorf("invalid object size index")
	}

	size, err := strconv.Atoi(string(data[formatIndex:sizeIndex]))
	if err != nil {
		return nil, fmt.Errorf("invalid object size: %v", err)
	}

	if size < 0 || size > int(^uint32(0)) {
		return nil, fmt.Errorf("value out of range for uint32 conversion")
	}

	if size != len(data)-(sizeIndex-1) {
		return nil, fmt.Errorf("object size mismatch")
	}

	// objectSize := uint32(size)

	switch objectFormat {
	// case "commit":
	// 	c := CreateCommit(data[sizeIndex+1:])
	// 	return c, nil
	// case "tree":
	// 	c := CreateTree(data[sizeIndex+1:])
	// 	return c, nil
	// case "tag":
	// 	c := CreateTag(data[sizeIndex+1:])
	// 	return c, nil
	// case "blob":
	// 	c := CreateBlob(data[sizeIndex+1:])
	// 	return c, nil
	default:
		return nil, fmt.Errorf("unknown object type: %s", objectFormat)
	}
}

func WriteObject(workTree string, object *Object) (string, error) {

	header := []byte(object.format + " ")
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, object.size)

	result := append(header, length...)
	result = append(result, 0)
	result = append(result, object.Data...)

	sha := sha256.Sum256(result)
	shaHex := hex.EncodeToString(sha[:])

	if workTree == "" {
		return "", fmt.Errorf("no orf worktree path found")
	}

	dirPath := filepath.Join(workTree, "objects", shaHex[:2])
	filePath := filepath.Join(workTree, "objects", shaHex[:2], shaHex[2:])

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
