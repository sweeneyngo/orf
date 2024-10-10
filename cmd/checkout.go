package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
	"os"
	"path/filepath"
)

// Checkout checks out a commit into the specified path.
func Checkout(hash string, path string) error {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return fmt.Errorf("error finding repo: %w", err)
	}

	obj, err := getObject(repo, hash)
	if err != nil {
		return err
	}

	var treeObj *object.Tree
	switch v := obj.(type) {
	case *object.Commit:
		treeObj, err = getTreeFromCommit(v, repo)
		if err != nil {
			return err
		}
	case *object.Tree:
		treeObj = v
	default:
		return fmt.Errorf("unexpected object type: %T", obj)
	}

	// Ensure the target directory exists and is empty
	if err := prepareDirectory(path); err != nil {
		return err
	}

	// Checkout the tree into the path
	return checkoutTree(repo, treeObj, path)
}

// getObject reads the object by its hash, returning either a Commit object or a base.
func getObject(repo *repository.Repo, hash string) (interface{}, error) {
	obj, err := object.ReadObject(repo.Directory, object.FindObject(repo, hash, "commit", false))
	if err != nil {
		return nil, err
	}

	// Check if it's a commit object
	if obj.GetFormat() == "commit" {
		c, ok := obj.(*object.Commit)
		if !ok {
			return nil, fmt.Errorf("failed to parse commit object")
		}
		return c, nil
	}

	return obj, nil
}

// prepareDirectory ensures the target path exists and is empty.
func prepareDirectory(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", path, err)
			}
		} else {
			return fmt.Errorf("failed to stat path %s: %v", path, err)
		}
	}

	// Check if it's a directory and it's empty
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", path)
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("failed to read directory %s: %v", path, err)
	}
	if len(files) > 0 {
		return fmt.Errorf("directory is not empty: %s", path)
	}

	return nil
}

// getTreeFromCommit retrieves the tree object associated with a commit.
func getTreeFromCommit(commit *object.Commit, repo *repository.Repo) (*object.Tree, error) {

	tree, _ := commit.GetKVData().Get("tree")

	var treeHash string
	switch v := tree.(type) {
	case string:
		treeHash = v
	case []byte:
		treeHash = string(v)
	default:
		return nil, fmt.Errorf("unexpected type %T for 'tree'", v)
	}

	obj, err := object.ReadObject(repo.Directory, treeHash)
	if err != nil {
		return nil, err
	}

	if obj.GetFormat() != "tree" {
		return nil, fmt.Errorf("expected tree object, got %s", obj.GetFormat())
	}

	t, ok := obj.(*object.Tree)
	if !ok {
		return nil, fmt.Errorf("incorrect object format for tree")
	}

	return t, nil
}

// checkoutTree recursively checks out the tree into the specified directory.
func checkoutTree(repo *repository.Repo, tree *object.Tree, path string) error {
	for _, item := range tree.GetLeaves() {
		obj, err := object.ReadObject(repo.Directory, item.GetHash())
		if err != nil {
			return fmt.Errorf("failed to read object %s: %v", item.GetHash(), err)
		}

		dest := filepath.Join(path, item.GetPath())

		// Process tree (directory) or blob (file)
		switch obj.GetFormat() {
		case "tree":
			if err := os.MkdirAll(dest, os.ModePerm); err != nil {
				return fmt.Errorf("failed to create directory %s: %v", dest, err)
			}
			if err := checkoutTree(repo, obj.(*object.Tree), dest); err != nil {
				return err
			}
		case "blob":
			if err := writeData(dest, obj.GetData()); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported object format %s", obj.GetFormat())
		}
	}

	return nil
}

// writeData writes the given data to the specified file path.
func writeData(dest string, data []byte) error {
	f, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %v", dest, err)
	}
	defer f.Close()

	_, err = f.Write(data)
	if err != nil {
		return fmt.Errorf("failed to write data to file %s: %v", dest, err)
	}

	return nil
}
