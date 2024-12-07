package cmd

import (
	"fmt"
	"orf/index"
	"orf/object"
	"orf/repository"

	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func Status() error {
	repo, err := repository.FindRepo(".", false)
	if err != nil {
		return err
	}

	index, err := index.ReadIndex(repo)
	if err != nil {
		return err
	}

	fmt.Printf(GetBranch(repo))
	printIndexHead(repo, index)
	printIndexWorkTree(repo, index)

	return nil
}

func GetBranch(repo *repository.Repo) (string, error) {

	// Check if ref: refs/heads/ exists in HEAD file
	headFile, err := repository.GetFilePath(repo.WorkTree, false, "HEAD")
	if err != nil {
		return "", err
	}

	// Check if HEAD file starts with "ref: refs/heads/"
	file, err := os.Open(headFile)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// Read the file content
	var head string
	_, err = fmt.Fscanf(file, "%s", &head)
	if err != nil {
		return "", err
	}

	// Check if the HEAD file starts with "ref: refs/heads/"
	if strings.HasPrefix(head, "ref: refs/heads/") {
		return head[16:], nil // Return the branch name (removing the "ref: refs/heads/" part)
	}

	// If no branch reference is found, return an empty string or an error
	return "", fmt.Errorf("no active branch found")
}

func printIndexHead(repo *repository.Repo, index *index.Index) error {
	fmt.Println("Changes to be committed:")
	head, err := treeToDict(repo, "HEAD", "")

	if err != nil {
		return err
	}

	for _, entry := range index.Entries {
		if head[entry.Name] != "" {
			if head[entry.Name] != entry.Sha {
				fmt.Printf("  (modified) %s\n", entry.Name)
			}

			delete(head, entry.Name)
		} else {
			fmt.Printf("  (new file) %s\n", entry.Name)
		}
	}

	for _, entry := range head {
		fmt.Printf("  (deleted) %s\n", entry)
	}
	return nil
}

func printIndexWorkTree(repo *repository.Repo, index *index.Index) error {
	fmt.Println("Changes not staged for commit:")
	gitDirPrefix := repo.Directory + string(os.PathSeparator)

	// List all files in the working tree
	var allFiles []string

	// Walk the filesystem in the worktree directory
	err := filepath.Walk(repo.WorkTree, func(root string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip files in the .git directory or its subdirectories
		if root == repo.Directory || strings.HasPrefix(root, gitDirPrefix) {
			return filepath.SkipDir
		}

		// Only process files (not directories)
		if !info.IsDir() {
			relPath, err := filepath.Rel(repo.WorkTree, root)
			if err != nil {
				return err
			}
			allFiles = append(allFiles, relPath)
		}

		return nil
	})

	if err != nil {
		return err
	}

	// Traverse the index and compare the actual files
	for _, entry := range index.Entries {
		fullPath := filepath.Join(repo.WorkTree, entry.Name)

		// If the file is not in the working tree (deleted)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			fmt.Println("  deleted:", entry.Name)
		} else {
			// Compare file metadata (ctime and mtime)
			stat, err := os.Stat(fullPath)
			if err != nil {
				return err
			}

			// Assuming entry.ctime and entry.mtime are in the form of uint64 arrays
			ctime := entry.CTimeSec*1e9 + entry.CTimeNsec
			mtime := entry.MTimeSec*1e9 + entry.MTimeNsec

			// If the metadata is different, compare the file hashes
			if stat.Sys().(*syscall.Stat_t).Ctim.Nsec != int64(ctime) || stat.ModTime().UnixNano() != int64(mtime) {
				// Open the file and compute the hash
				data, err := os.ReadFile(fullPath)
				if err != nil {
					return err
				}
				newSha, err := GetHash(data, "blob", "")
				if err != nil {
					return err
				}

				// If the hashes are different, the file is modified
				if entry.Sha != newSha {
					fmt.Println("  modified:", entry.Name)
				}
			}
		}

		// Remove the file from allFiles as it exists in the index
		for i, f := range allFiles {
			if f == entry.Name {
				allFiles = append(allFiles[:i], allFiles[i+1:]...)
				break
			}
		}
	}
	return nil
}

func treeToDict(repo *repository.Repo, ref string, prefix string) (map[string]string, error) {

	ret := make(map[string]string)

	// Find the tree object
	treeHash := object.FindObject(repo, ref, "tree", false)
	if treeHash == "" {
		return nil, fmt.Errorf("failed to find tree object for ref: %s", ref)
	}

	// Read the tree object (it's expected to be a "tree" object)
	tree, err := object.ReadObject(repo.Directory, treeHash)
	if err != nil {
		return nil, fmt.Errorf("failed to read tree object: %v", err)
	}

	// Ensure the object is actually a tree
	t, ok := tree.(*object.Tree)
	if !ok {
		return nil, fmt.Errorf("unexpected type for tree: %T, expected *object.Tree", tree)
	}

	// Loop over the tree leaves (files or subtrees)
	for _, leaf := range t.Leaves {
		fullPath := filepath.Join(prefix, leaf.Path)

		// Check if the leaf is a subtree (i.e., directory)
		isSubtree := leaf.Mode[0] == 0x04 // This checks if the mode starts with '04', indicating a directory

		// If it's a directory (subtree), recurse; otherwise, add it to the map
		if isSubtree {
			// Recurse into the subtree and update the map with its contents
			subtree, err := treeToDict(repo, leaf.Hash, fullPath)
			if err != nil {
				return nil, fmt.Errorf("failed to process subtree %s: %v", fullPath, err)
			}
			// Merge the subtree map into the result map
			for k, v := range subtree {
				ret[k] = v
			}
		} else {
			ret[fullPath] = leaf.Hash
		}
	}

	return ret, nil
}
