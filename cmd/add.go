package cmd

import (
	"fmt"
	"orf/index"
	"orf/repository"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

func Add(paths []string) error {
	// Find the repository
	repo, err := repository.FindRepo(".", false)
	if err != nil {
		return err
	}

	// Add the file to the index
	if err := add(repo, paths, true, false); err != nil {
		return err
	}

	return nil
}

func add(repo *repository.Repo, paths []string, delete bool, skipMissing bool) error {
	// First, remove the paths from the index if they exist
	if err := RemovePaths(repo, paths, delete, skipMissing); err != nil {
		return err
	}

	worktree := repo.WorkTree + string(os.PathSeparator)

	// Convert paths to pairs (absolute, relative to worktree)
	var cleanPaths []struct {
		Abspath string
		Relpath string
	}
	for _, path := range paths {
		abspath, err := filepath.Abs(path)
		if err != nil {
			return fmt.Errorf("failed to get absolute path: %v", err)
		}

		// Check if the file exists and is within the worktree
		if !strings.HasPrefix(abspath, worktree) || !repository.IsFile(abspath) {
			return fmt.Errorf("not a file, or outside the worktree: %v", path)
		}

		relpath, err := filepath.Rel(worktree, abspath)
		if err != nil {
			return fmt.Errorf("failed to get relative path: %v", err)
		}
		cleanPaths = append(cleanPaths, struct {
			Abspath string
			Relpath string
		}{Abspath: abspath, Relpath: relpath})
	}

	// Read the current index
	indx, err := index.ReadIndex(repo)
	if err != nil {
		return fmt.Errorf("failed to read index: %v", err)
	}

	// Add each file to the index
	for _, pathInfo := range cleanPaths {
		abspath := pathInfo.Abspath
		relpath := pathInfo.Relpath

		// Compute the SHA of the file
		sha, err := HashObject(abspath, "blob", false)
		if err != nil {
			return fmt.Errorf("failed to hash file %v: %v", abspath, err)
		}

		// Get file metadata
		stat, err := os.Stat(abspath)
		if err != nil {
			return fmt.Errorf("failed to stat file %v: %v", abspath, err)
		}

		ctime := stat.ModTime().Unix()
		ctimeNs := int64(stat.ModTime().Nanosecond())
		mtime := stat.ModTime().Unix()
		mtimeNs := int64(stat.ModTime().Nanosecond())

		// Create the index entry
		entry := index.IndexEntry{
			Name:       relpath,
			Sha:        sha,
			CTimeSec:   uint64(ctime),
			CTimeNsec:  uint64(ctimeNs),
			MTimeSec:   uint64(mtime),
			MTimeNsec:  uint64(mtimeNs),
			Dev:        uint64(stat.Sys().(*syscall.Stat_t).Dev),
			Ino:        uint64(stat.Sys().(*syscall.Stat_t).Ino),
			ModeType:   0b1000, // Regular file mode
			ModePerms:  0o644,  // Default permissions
			Uid:        uint32(stat.Sys().(*syscall.Stat_t).Uid),
			Gid:        uint32(stat.Sys().(*syscall.Stat_t).Gid),
			Fsize:      uint32(stat.Size()),
			FlagsValid: false,
			FlagStaged: 0,
		}

		// Append the new entry to the index entries
		indx.Entries = append(indx.Entries, entry)
	}

	// Write the updated index back to the index file
	if err := indx.WriteIndex(repo); err != nil {
		return fmt.Errorf("failed to write index: %v", err)
	}

	return nil
}
