package repository

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-ini/ini"
)

type Repo struct {
	WorkTree  string
	Directory string
	Config    *ini.File
}

// CreateRepo creates a new repository at the specified path. It initializes the repository
// by calling initializeRepo with the force flag set to true. If the directory already exists
// and is not empty, it returns an error. It creates necessary directories and files for the
// repository, including branches, objects, refs, description, HEAD, and config files.
// It returns a pointer to the Repo struct if no errors have occurred.
func CreateRepo(path string) (*Repo, error) {

	repo, _ := initializeRepo(path, true)

	// Create directories
	directories := []string{
		"branches",
		"objects",
		filepath.Join("refs", "tags"),
		filepath.Join("refs", "heads"),
	}

	for _, directory := range directories {
		path := filepath.Join(repo.WorkTree, ".orf", directory)
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return nil, err
		}
	}

	// Create .orf/description
	descriptionPath := filepath.Join(repo.WorkTree, ".orf", "description")
	if err := os.WriteFile(descriptionPath, []byte("Unnamed repository; edit this file 'description' to name the repository.\n"), 0644); err != nil {
		return nil, err
	}

	// Create .orf/HEAD
	headPath := filepath.Join(repo.WorkTree, ".orf", "HEAD")
	if err := os.WriteFile(headPath, []byte("ref: refs/heads/master\n"), 0644); err != nil {
		return nil, err
	}

	// Create .orf/config
	configPath := filepath.Join(repo.WorkTree, ".orf", "config")
	configContent := strings.ReplaceAll(`[core]
		repositoryformatversion = 0
		filemode = false
		bare = false`, "\t", "")

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		return nil, err
	}

	return repo, nil
}

// FindRepo searches for a repository from a given path and moves up the directory tree.
// If a repository is found, it checks for validity then returns a pointer to the Repo struct.
// If the force flag is set to false, it will return an error if no repository is found.
// It returns a pointer to the repository once found.
func FindRepo(path string, force bool) (*Repo, error) {

	if isDir(filepath.Join(path, ".orf")) {
		return initializeRepo(path, false)
	}

	parentPath := filepath.Clean(filepath.Join(path, ".."))
	if parentPath == path {
		if !force {
			return nil, errors.New("no orf directory found")
		}
		return nil, nil
	}

	return FindRepo(parentPath, force)
}

// initializeRepo initializes a repository at the given path. If the force flag is set to true,
// it will create the repository even if the directory is not empty. It returns a pointer to the
// Repo struct if no errors have occurred.
func initializeRepo(path string, force bool) (*Repo, error) {

	directory := filepath.Join(path, ".orf")

	if !force && !isDir(directory) {
		return nil, fmt.Errorf("not a Orf repository: %s", path)
	}

	// Get configuration path
	configPath, err := GetFilePath(path, force, "config")
	if err != nil {
		return nil, fmt.Errorf("error getting config path: %w", err)
	}

	// Read configuration file .orf/config
	config, err := readConfig(configPath, force)
	if err != nil {
		return nil, fmt.Errorf("error reading config file %s: %w", configPath, err)
	}

	return &Repo{
		WorkTree:  path,
		Directory: directory,
		Config:    config,
	}, nil
}

// readConfig reads the configuration file from the given path and returns an ini.File object.
// If the force flag is set to false, it will return an error if the configuration file does not exist
// or if the repository format version is unsupported.
func readConfig(path string, force bool) (*ini.File, error) {
	file := filepath.Join(path, "config")
	if !force && !IsFile(file) {
		return nil, errors.New("no configuration file in path")
	}

	config, err := ini.Load(file)
	if !force && err != nil {
		return nil, err
	}

	if !force {
		version, err := config.Section("core").Key("repositoryformatversion").Int()
		if err != nil || version != 0 {
			return nil, fmt.Errorf("unsupported repositoryformatversion: %d", version)
		}
	}

	return config, nil
}

// GetFilePath returns the path to a file within the repository's work tree.
// It will create the directory structure leading to the file (barring the actual file), ensuring it exists.
func GetFilePath(WorkTree string, force bool, paths ...string) (string, error) {

	if _, err := GetDir(WorkTree, force, paths[:len(paths)-1]...); err != nil {
		return "", err
	}

	return getPath(WorkTree, paths...), nil
}

// GetDir ensures that the directory structure exists for the given path elements.
// If the force flag is true, it will create the directory structure if it does not exist.
func GetDir(WorkTree string, force bool, paths ...string) (string, error) {

	path := getPath(WorkTree, paths...)
	info, err := os.Stat(path)

	if err != nil {
		if force {
			if err := os.MkdirAll(path, os.ModePerm); err != nil {
				return "", err
			}
		} else {
			return "", err

		}
	}

	if err == nil && !info.IsDir() {
		return "", fmt.Errorf("not a directory: %s", path)
	}

	return path, nil
}

// getPath constructs a filepath by joining the workTree with variable path elements.
func getPath(WorkTree string, paths ...string) string {
	return filepath.Join(append([]string{WorkTree}, paths...)...)
}

// isFile checks if the given path corresponds to a file.
func IsFile(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return !info.IsDir()
}

// isDir checks if the given path corresponds to a directory.
func isDir(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return info.IsDir()
}
