package object

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"orf/utils"
	"sort"
)

// Represents a tree object, with leaves representing all Leaf objects.
type Tree struct {
	Base
	leaves []*Leaf
}

type Leaf struct {
	mode []byte
	path string
	hash string
}

func CreateTree(data []byte) *Tree {
	return &Tree{
		Base: Base{
			format: "tree",            // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			data:   data,              // Set the data directly in the Object
		},
		leaves: []*Leaf{},
	}
}

func (tree *Tree) GetFormat() string {
	return tree.format
}

func (tree *Tree) GetSize() uint32 {
	return tree.size
}

func (tree *Tree) GetData() []byte {
	return tree.data
}

func (tree *Tree) GetLeaves() []*Leaf {
	return tree.leaves
}

func (leaf *Leaf) GetMode() []byte {
	return leaf.mode
}

func (leaf *Leaf) GetPath() string {
	return leaf.path
}

func (leaf *Leaf) GetHash() string {
	return leaf.hash
}

func (tree *Tree) Serialize() ([]byte, error) {
	sort.Sort(ByPath(tree.leaves))
	output := []byte{}

	for _, leaf := range tree.leaves {
		output = utils.Append(output, leaf.mode...)
		output = utils.Append(output, ' ')
		output = utils.Append(output, []byte(leaf.path)...)
		output = utils.Append(output, '\x00')

		hashBytes, err := convertHexToBytes(leaf.hash)
		if err != nil {
			return nil, err
		}

		output = utils.Append(output, hashBytes...)
	}

	return output, nil
}

func (tree *Tree) Deserialize(data []byte) error {
	leaves, err := parseTree(data)
	if err != nil {
		return err
	}

	tree.leaves = leaves
	return nil
}

// parseTree parses the byte data and constructs a list of Leaves (mode, path, hash).
func parseTree(rawData []byte) ([]*Leaf, error) {
	startIndex := 0
	maxLength := len(rawData)

	var tree []*Leaf
	var leaf *Leaf
	var err error

	for startIndex < maxLength {
		startIndex, leaf, err = parseLeaf(rawData, startIndex)
		if err != nil {
			return nil, err
		}
		tree = append(tree, leaf)
	}
	return tree, nil
}

func parseLeaf(rawData []byte, startIndex int) (int, *Leaf, error) {

	modeIndex := utils.FindIndex(rawData, startIndex, ' ')
	if modeIndex-startIndex != 5 && modeIndex-startIndex != 6 {
		return -1, nil, fmt.Errorf("error parsing mode, incorrect num bytes")
	}

	mode := rawData[startIndex:modeIndex]
	if len(mode) == 5 {
		mode = utils.Prepend(rawData, ' ')
	}

	pathIndex := utils.FindIndex(rawData, startIndex, '\x00')
	path := rawData[modeIndex+1 : pathIndex]

	hash := hex.EncodeToString(rawData[pathIndex+1 : pathIndex+21])

	return pathIndex + 21, &Leaf{
		mode: mode,
		path: string(path),
		hash: hash,
	}, nil
}

// ByPath is a sort.Interface that follows these custom rules:
// Directories (that is, tree entries) are sorted with a final / added.
// It matters, because directories are sorted after files, and therefore is less than files.
type ByPath []*Leaf

func (p ByPath) Len() int { return len(p) }
func (p ByPath) Less(i, j int) bool {

	if isModeDirectory(p[i].mode) && !isModeDirectory(p[j].mode) {
		return true
	}

	if !isModeDirectory(p[i].mode) && isModeDirectory(p[j].mode) {
		return false
	}

	// If both are the same mode, sort by path
	return p[i].path < p[j].path
}

func (p ByPath) Swap(i, j int) { p[i], p[j] = p[j], p[i] }

// isModeDirectory checks if a mode (byte array) is a directory or a file ("10").
func isModeDirectory(mode []byte) bool {
	return bytes.HasPrefix(mode, []byte("10"))
}

// convertHexToBytes converts a hash string (in hex format) to a byte format, truncated by 20 bytes.
func convertHexToBytes(hash string) ([]byte, error) {
	result, err := hex.DecodeString(hash)
	if err != nil {
		return nil, err
	}

	if len(result) > 20 {
		result = result[:20]
	}

	return result, nil
}
