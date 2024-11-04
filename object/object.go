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
	"regexp"
	"strconv"
	"strings"
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
	shaList, err := ResolveObject(repo, name)
	if err != nil {
		return ""
	}

	if len(shaList) == 0 {
		fmt.Printf("No such reference %s.", name)
		return ""
	}

	if len(shaList) > 1 {
		fmt.Printf("Ambiguous reference %s: Candidates are:\n - %s.", name, shaList)
		return ""
	}

	sha := shaList[0]
	if format == "" {
		return sha
	}

	for {
		obj, err := ReadObject(repo.Directory, sha)
		if err != nil {
			fmt.Printf("Error reading object %s: %v", sha, err)
			return ""
		}

		if string(obj.GetFormat()) == format {
			fmt.Printf("Found %s object %s.", format, sha)
			return sha
		}

		if !follow {
			fmt.Printf("Object %s is a %s, not a %s.", sha, obj.GetFormat(), format)
			return ""
		}

		// Follow tags
		if string(obj.GetFormat()) == "tag" {

			// Assert tag
			tag, ok := obj.(*Tag)
			if !ok {
				fmt.Printf("Object %s is not a tag", sha)
				return ""
			}

			value, found := tag.GetKVData().Get("object")

			if !found {
				fmt.Printf("Tag %s does not have an object", sha)
				return ""
			}

			// Assert string
			valueStr, ok := value.(string)
			if !ok {
				fmt.Printf("Tag %s object is not a string", sha)
				return ""
			}

			sha = valueStr

		} else if string(obj.GetFormat()) == "commit" && format == "tree" {

			// Assert commit
			commit, ok := obj.(*Commit)
			if !ok {
				fmt.Printf("Object %s is not a commit", sha)
				return ""
			}

			value, found := commit.GetKVData().Get("tree")

			if !found {
				fmt.Printf("Commit %s does not have an tree", sha)
				return ""
			}

			// Assert string
			valueStr, ok := value.(string)
			if !ok {
				fmt.Printf("Commit %s tree is not a string", sha)
				return ""
			}

			sha = valueStr

		} else {
			return ""
		}
	}
}

func ResolveObject(repo *repository.Repo, name string) ([]string, error) {

	var candidates []string
	hashRE := regexp.MustCompile(`^[0-9A-Fa-f]{4,40}$`)

	// If the name is empty, return nil.
	if strings.TrimSpace(name) == "" {
		return nil, nil
	}

	// Head is nonambiguous
	if name == "HEAD" {
		if headRef, found := resolveRef(repo, "HEAD"); found != nil {
			candidates = append(candidates, headRef)
		}
		return candidates, nil
	}

	// If it's a hex string, try for a hash.
	if hashRE.MatchString(name) {
		name = strings.ToLower(name)
		prefix := name[:2]
		path, err := repository.GetDir(repo.WorkTree, false, "objects", prefix)

		if err != nil {
			return nil, err
		}

		// Check if the directory exists
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return nil, nil
		}

		rem := name[2:]
		files, err := os.ReadDir(path)
		if err != nil {
			return nil, err
		}

		for _, f := range files {
			if strings.HasPrefix(f.Name(), rem) {
				candidates = append(candidates, prefix+f.Name())
			}
		}
	}

	// Try for references.
	if asTag, found := resolveRef(repo, "refs/tags/"+name); found != nil {
		candidates = append(candidates, asTag)
	}

	if asBranch, found := resolveRef(repo, "refs/heads/"+name); found != nil {
		candidates = append(candidates, asBranch)
	}

	return candidates, nil
}
