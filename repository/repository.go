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

func InitializeRepo(path string, force bool) (*Repo, error) {

	directory := filepath.Join(path, ".orf")

	if !force && !isDir(directory) {
		return nil, fmt.Errorf("not a Orf repository: %s", path)
	}

	// Get configuration path
	configPath, err := GetFile(path, force, "config")
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

func CreateRepo(path string) (*Repo, error) {

	repo, err := InitializeRepo(path, true)

	if err != nil {
		return nil, errors.New("error creating repo")
	}

	if info, err := os.Stat(repo.WorkTree); err == nil {

		if !info.IsDir() {
			return nil, fmt.Errorf("not a directory: %s", repo.WorkTree)
		}

		if _, err := os.Stat(repo.Directory); err == nil {
			if entries, err := os.ReadDir(repo.Directory); err == nil && len(entries) > 0 {
				return nil, fmt.Errorf("%s is not empty", repo.Directory)
			}
		}
	} else {
		if err := os.MkdirAll(repo.WorkTree, os.ModePerm); err != nil {
			return nil, err
		}
	}

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

func FindRepo(path string, force bool) (*Repo, error) {

	if isDir(filepath.Join(path, ".orf")) {
		return InitializeRepo(path, true)
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

func GetFile(WorkTree string, force bool, paths ...string) (string, error) {

	if _, err := getDir(WorkTree, force, paths[:len(paths)-1]...); err != nil {
		return "", err
	}

	return getPath(WorkTree, paths...), nil
}

func isFile(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return !info.IsDir()
}

func isDir(path string) bool {
	info, err := os.Stat(path)

	if err != nil {
		return false
	}

	return info.IsDir()
}

func getPath(WorkTree string, paths ...string) string {
	return filepath.Join(append([]string{WorkTree}, paths...)...)
}

func getDir(WorkTree string, force bool, paths ...string) (string, error) {

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

func readConfig(path string, force bool) (*ini.File, error) {
	file := filepath.Join(path, "config")
	if !force && !isFile(file) {
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
