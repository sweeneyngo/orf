package cmd

import (
	"fmt"
	"orf/index"
	"orf/object"
	"orf/repository"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/go-ini/ini"
)

func Commit(message string) error {
	// Find the repository
	repo, err := repository.FindRepo(".", false)
	if err != nil {
		return err
	}

	// Read the index
	index, err := index.ReadIndex(repo)
	if err != nil {
		return err
	}

	// Create the tree and get the SHA for the root tree
	tree, err := TreeFromIndex(repo, index.Entries)
	if err != nil {
		return err
	}

	parent := object.FindObject(repo, "HEAD", "commit", false)

	// Get the author from the git config (simulated here)
	config, err := readGitConfig()
	if err != nil {
		return err
	}

	author := getUserGitConfig(config)

	// Create the commit
	commit, err := WriteCommit(repo, tree, parent, author, time.Now(), message)
	if err != nil {
		return err
	}

	// Get the active branch
	activeBranch, err := GetBranch(repo)
	if err != nil {
		return err
	}

	// Update the HEAD or the active branch reference
	if activeBranch != "" {
		// If on a branch, update the ref
		refPath, err := repository.GetFilePath(repo.WorkTree, false, "refs/heads/"+activeBranch)
		if err != nil {
			return err
		}

		err = os.WriteFile(refPath, []byte(commit+"\n"), 0644)
		if err != nil {
			return err
		}

	} else {
		// If not on a branch, update HEAD itself
		headPath, err := repository.GetFilePath(repo.WorkTree, false, "HEAD")
		if err != nil {
			return err
		}
		err = os.WriteFile(headPath, []byte(commit+"\n"), 0644)
		if err != nil {
			return err
		}
	}

	fmt.Println("Commit successful!")
	return nil
}

func WriteCommit(repo *repository.Repo, tree, parent, author string, timestamp time.Time, message string) (string, error) {
	commit := object.CreateCommit(nil)

	// Add tree and parent to the commit
	commit.GetKVData().Add("tree", []byte(tree))

	if parent != "" {
		commit.GetKVData().Add("parent", []byte(parent))
	}

	// Format the timestamp and timezone
	offset := int(timestamp.UTC().Sub(timestamp).Seconds())
	hours := offset / 3600
	minutes := (offset % 3600) / 60
	tz := fmt.Sprintf("%+03d%02d", hours, minutes)

	// Format the author and committer string
	authorString := fmt.Sprintf("%s %d %s", author, timestamp.Unix(), tz)
	commit.GetKVData().Add("author", []byte(authorString))
	commit.GetKVData().Add("committer", []byte(authorString))

	// Add the commit message
	commit.GetKVData().Add("]", []byte(authorString))

	// Simulate writing the commit object to the object store and return a SHA hash
	// In a real implementation, this would interact with the Git object store.
	return "dummy-sha-commit", nil
}

func TreeFromIndex(repo *repository.Repo, indexEntries []index.IndexEntry) (string, error) {
	contents := make(map[string][]interface{}) // Directory -> list of entries (files or trees)

	// Initialize the contents map with the root directory
	contents[""] = []interface{}{}

	// Iterate over index entries to categorize them into directories
	for _, entry := range indexEntries {
		dirname := filepath.Dir(entry.Name)

		// Ensure every directory in the path exists in contents
		key := dirname
		for key != "" {
			if _, exists := contents[key]; !exists {
				contents[key] = []interface{}{}
			}
			key = filepath.Dir(key)
		}

		// Add the entry to the directory list
		contents[dirname] = append(contents[dirname], entry)
	}

	// Sort directories by length in descending order to process child dirs first
	var sortedPaths []string
	for path := range contents {
		sortedPaths = append(sortedPaths, path)
	}
	sort.Slice(sortedPaths, func(i, j int) bool {
		return len(sortedPaths[i]) > len(sortedPaths[j])
	})

	var sha string

	// Process each directory path, starting from the deepest directory
	for _, path := range sortedPaths {
		tree := &object.Tree{}

		// Add entries (files or subdirectories) to the tree
		for _, entry := range contents[path] {
			switch e := entry.(type) {
			case index.IndexEntry: // Regular file entry
				// Convert mode to octal format (simplified)
				leafMode := fmt.Sprintf("%02o%04o", e.ModeType, e.ModePerms)
				leaf := &object.Leaf{
					Mode: []byte(leafMode),
					Path: filepath.Base(e.Name),
					Hash: e.Sha,
				}
				tree.Leaves = append(tree.Leaves, leaf)
			default:
				// Handle other types, e.g., subtrees (not implemented here)
			}
		}

		// Write the tree to the object store and get its SHA
		var err error
		sha, err = object.WriteObject(repo.Directory, tree)
		if err != nil {
			return "", err
		}

		// Add the new tree hash to the parent directory
		parent := filepath.Dir(path)
		base := filepath.Base(path)
		contents[parent] = append(contents[parent], struct {
			Name string
			Sha  string
		}{base, sha})
	}

	return sha, nil
}

func readGitConfig() (*ini.File, error) {
	// Get the XDG_CONFIG_HOME environment variable or default to ~/.config
	xdgConfigHome := os.Getenv("XDG_CONFIG_HOME")
	if xdgConfigHome == "" {
		// Default to ~/.config if XDG_CONFIG_HOME is not set
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("unable to get user home directory: %v", err)
		}
		xdgConfigHome = filepath.Join(homeDir, ".config")
	}

	// Construct the list of config files to check
	configFiles := []string{
		filepath.Join(xdgConfigHome, "git", "config"),
		filepath.Join(os.Getenv("HOME"), ".gitconfig"),
	}

	// Read the configuration files using the ini package
	var config *ini.File
	for _, configFile := range configFiles {
		// Check if the file exists before attempting to read
		if _, err := os.Stat(configFile); err == nil {
			cfg, err := ini.Load(configFile)
			if err != nil {
				return nil, fmt.Errorf("unable to read config file %v: %v", configFile, err)
			}
			config = cfg
			break
		}
	}

	if config == nil {
		return nil, fmt.Errorf("no valid git configuration files found")
	}

	return config, nil
}

func getUserGitConfig(config *ini.File) string {
	// Check if the "user" section exists
	userSection := config.Section("user")
	if userSection != nil {
		// Get the "name" and "email" fields from the user section
		name := userSection.Key("name").String()
		email := userSection.Key("email").String()
		if name != "" && email != "" {
			// Format and return the user identity
			return fmt.Sprintf("%s <%s>", name, email)
		}
	}

	// If the "user" section or fields are missing, return an empty string
	return ""
}
