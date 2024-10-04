package repository

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// mocks
func MockGetFile(path string, force bool, name string) (string, error) {
	return "", errors.New("file not found") // Simulate an error
}

func TestInitializeRepo(t *testing.T) {
	path := t.TempDir()
	repo, err := InitializeRepo(path, false)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestInitializeForceRepo(t *testing.T) {
	path := t.TempDir()
	repo, err := InitializeRepo(path, true)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	assert.Equal(t, path, repo.WorkTree)
	assert.Equal(t, filepath.Join(path, ".orf"), repo.Directory)
	assert.Nil(t, repo.Config)
}

// func TestInitializeRepo_ErrorGettingConfigPath(t *testing.T) {
// 	// Replace GetFile with the mock version
// 	originalGetFile := GetFile
// 	defer func() { GetFile = originalGetFile }() // Restore original function after test
// 	GetFile = MockGetFile

// 	// Call the function under test
// 	_, err := InitializeRepo("some/path", false)

// 	// Check if the error is as expected
// 	if err == nil {
// 		t.Fatal("expected an error but got none")
// 	}

// 	// Check if the error message matches the expected error
// 	if !errors.Is(err, errors.New("file not found")) {
// 		t.Errorf("expected error 'file not found', got: %v", err)
// 	}
// }

func TestCreateRepo(t *testing.T) {
	path := t.TempDir()
	repo, err := CreateRepo(path)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.DirExists(t, filepath.Join(path, ".orf", "branches"))
	assert.DirExists(t, filepath.Join(path, ".orf", "objects"))
	assert.DirExists(t, filepath.Join(path, ".orf", "refs", "tags"))
	assert.DirExists(t, filepath.Join(path, ".orf", "refs", "heads"))
	assert.FileExists(t, filepath.Join(path, ".orf", "description"))
	assert.FileExists(t, filepath.Join(path, ".orf", "HEAD"))
	assert.FileExists(t, filepath.Join(path, ".orf", "config"))
}

func TestFindRepo_NoRepository(t *testing.T) {
	// Test case: No repository found
	path := t.TempDir()
	repo, err := FindRepo(path, false)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestFindForceRepo_NoRepository(t *testing.T) {
	// Test case: Force search with no repository found
	repo, err := FindRepo("/non/existent/path", true)
	assert.NoError(t, err)
	assert.Nil(t, repo)
}

func TestFindRepo_RootRepository(t *testing.T) {
	// Test case: Repository found in the same directory
	path := t.TempDir()

	repoPath := filepath.Join(path, ".orf")
	err := os.MkdirAll(repoPath, os.ModePerm)
	assert.NoError(t, err)

	repo, err := FindRepo(path, false)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, path, repo.WorkTree)
	assert.Equal(t, repoPath, repo.Directory)
}

func TestFindRepo(t *testing.T) {
	// Test case: Repository found in a parent directory
	path := t.TempDir()

	repoPath := filepath.Join(path, ".orf")
	err := os.MkdirAll(repoPath, os.ModePerm)
	assert.NoError(t, err)

	childPath := filepath.Join(path, "child")
	err = os.MkdirAll(repoPath, os.ModePerm)
	assert.NoError(t, err)

	repo, err := FindRepo(childPath, false)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	assert.Equal(t, path, repo.WorkTree)
	assert.Equal(t, repoPath, repo.Directory)
}

func TestIsFile(t *testing.T) {
	file, err := os.CreateTemp("", "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	assert.True(t, isFile(file.Name()))
	assert.False(t, isFile("/non/existent/path"))
}

func TestIsDir(t *testing.T) {
	dir := t.TempDir()
	assert.True(t, isDir(dir))
	assert.False(t, isDir("/non/existent/path"))
}

func TestGetPath(t *testing.T) {
	workTree := "/home/user/repo"
	expected := filepath.Join(workTree, "subdir", "file")
	assert.Equal(t, expected, getPath(workTree, "subdir", "file"))
}

func TestGetFile(t *testing.T) {
	workTree := t.TempDir()
	filePath, err := GetFile(workTree, false, "subdir", "file")
	assert.Error(t, err)
	assert.Equal(t, "", filePath)
}

func TestGetForceFile(t *testing.T) {
	workTree := t.TempDir()
	filePath, err := GetFile(workTree, true, "subdir", "file")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(workTree, "subdir", "file"), filePath)
}

func TestGetDir(t *testing.T) {
	workTree := t.TempDir()
	dirPath, err := getDir(workTree, false, "subdir")
	assert.Error(t, err)
	assert.Equal(t, "", dirPath)
}

func TestGetNotDir(t *testing.T) {
	workTree := t.TempDir()
	file, err := os.CreateTemp(workTree, "testfile")
	assert.NoError(t, err)
	defer os.Remove(file.Name())

	dirPath, err := getDir(workTree, false, file.Name())
	assert.Error(t, err)
	assert.Equal(t, "", dirPath)
}

func TestGetForceDir(t *testing.T) {
	workTree := t.TempDir()
	dirPath, err := getDir(workTree, true, "subdir")
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(workTree, "subdir"), dirPath)
}

func TestReadConfig(t *testing.T) {
	workTree := t.TempDir()
	configPath := filepath.Join(workTree, ".orf")
	err := os.MkdirAll(configPath, os.ModePerm)
	assert.NoError(t, err)

	configContent := `[core]
	repositoryformatversion = 0`
	err = os.WriteFile(filepath.Join(configPath, "config"), []byte(configContent), 0644)
	assert.NoError(t, err)

	config, err := readConfig(configPath, false)
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, 0, config.Section("core").Key("repositoryformatversion").MustInt())
}
