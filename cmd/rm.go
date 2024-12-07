package cmd

import (
	"fmt"
	"orf/index"
	"orf/repository"
	"orf/utils"
	"os"
	"path/filepath"
	"strings"
)

func Remove(paths []string) error {
	repo, err := repository.FindRepo(".", false)
	if err != nil {
		return err
	}

	if err := RemovePaths(repo, paths, true, false); err != nil {
		return err
	}

	return nil
}

func RemovePaths(repo *repository.Repo, paths []string, delete bool, skipMissing bool) error {

	indx, err := index.ReadIndex(repo)
	if err != nil {
		return fmt.Errorf("failed to read index: %v", err)
	}

	// Define the worktree path (directory where the repository's files are stored)
	worktree := repo.WorkTree + string(os.PathSeparator)

	// Make paths absolute and check if they are within the worktree
	var abspaths []string
	for _, path := range paths {
		abspath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path for %v: %v", path, err)
		}

		if strings.HasPrefix(abspath, worktree) {
			abspaths = append(abspaths, abspath)
		} else {
			return fmt.Errorf("cannot remove paths outside of worktree: %v", paths)
		}
	}

	// Lists to keep track of entries and paths to be removed
	var keptEntries []index.IndexEntry
	var remove []string

	// Iterate over the index entries and separate the ones to remove and to keep
	for _, e := range indx.Entries {
		fullPath := filepath.Join(repo.WorkTree, e.Name)

		// If the entry is in the list of paths to remove
		if utils.Contains(abspaths, fullPath) {
			remove = append(remove, fullPath)
			// Remove it from abspaths (so we know all paths were accounted for)
			abspaths = utils.RemoveFromSlice(abspaths, fullPath)
		} else {
			keptEntries = append(keptEntries, e) // Keep entry
		}
	}

	// If there are remaining paths in abspaths, it means they are not in the index
	if len(abspaths) > 0 && !skipMissing {
		return fmt.Errorf("cannot remove paths not in the index: %v", abspaths)
	}

	// Physically delete the paths from the filesystem, if the delete flag is true
	if delete {
		for _, path := range remove {
			if err := os.Remove(path); err != nil {
				return fmt.Errorf("failed to remove file %v: %v", path, err)
			}
		}
	}

	// Update the index entries and write the updated index back to the file
	indx.Entries = keptEntries
	if err := indx.WriteIndex(repo); err != nil {
		return fmt.Errorf("failed to write updated index: %v", err)
	}

	return nil
}
