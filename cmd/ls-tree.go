package cmd

import (
	"bytes"
	"fmt"
	"orf/object"
	"orf/repository"
	"path/filepath"
)

func ListTree(tree string, recursive bool) error {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return fmt.Errorf("error finding repo: %w", err)
	}

	// Start the log from the commit hash provided in args
	err = listTree(repo, object.FindObject(repo, tree, "tree", false), recursive, "")
	if err != nil {
		return err
	}

	return nil
}

// listTree recursively lists the tree objects in a orf repository.
// It prints the tree/commit/leaf information with the appropriate padding, type, and path.
func listTree(repo *repository.Repo, ref string, recursive bool, prefix string) error {

	hash := object.FindObject(repo, ref, "tree", false)
	tree, err := object.ReadObject(repo.Directory, hash)
	if err != nil {
		return fmt.Errorf("failed to read object %s: %v", ref, err)
	}

	// Assert the object is of type *object.Tree
	t, ok := tree.(*object.Tree)
	if !ok {
		return fmt.Errorf("unexpected type %T for tree, expected *Tree", tree)
	}

	for _, leaf := range t.Leaves {
		leafMode := getLeafMode(leaf)
		leafType, err := getLeafType(leafMode)
		if err != nil {
			return fmt.Errorf("invalid leaf mode %v: %v", leafMode, err)
		}

		// If not recursive or the leaf is not a tree, print its details
		if !recursive || leafType != "tree" {
			paddedMode := fmt.Sprintf("%06s", string(leafMode)) // Pad mode to 6 digits
			fmt.Printf("%s %s %s\t%s\n", paddedMode, leafType, leaf.Hash, filepath.Join(prefix, leaf.Path))
		} else {
			err = listTree(repo, leaf.Hash, recursive, filepath.Join(prefix, leaf.Path))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// getLeafMode extracts the mode prefix from a leaf, either 1 byte or 2 bytes.
func getLeafMode(leaf *object.Leaf) []byte {
	if len(leaf.Mode) == 5 {
		return leaf.Mode[0:1]
	}
	return leaf.Mode[0:2]
}

// getLeafType returns the type of the leaf based on its mode prefix.
func getLeafType(mode []byte) (string, error) {
	const (
		treeMode    = 0x04
		blobMode    = 0x10
		symlinkMode = 0x12
		commitMode  = 0x16
	)
	switch {
	case bytes.Equal(mode, []byte{treeMode}):
		return "tree", nil
	case bytes.Equal(mode, []byte{blobMode}), bytes.Equal(mode, []byte{symlinkMode}):
		return "blob", nil
	case bytes.Equal(mode, []byte{commitMode}):
		return "commit", nil
	default:
		return "", fmt.Errorf("unknown leaf mode %v", mode)
	}
}
