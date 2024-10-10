package repository

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInitializeRepo(t *testing.T) {
	path := t.TempDir()
	repo, err := initializeRepo(path, false)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestInitializeForceRepo(t *testing.T) {
	path := t.TempDir()
	repo, err := initializeRepo(path, true)
	assert.NoError(t, err)
	assert.NotNil(t, repo)

	assert.Equal(t, path, repo.WorkTree)
	assert.Equal(t, filepath.Join(path, ".orf"), repo.Directory)
	assert.Nil(t, repo.Config)
}

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

	// Create the config file inside .orf
	configPath := filepath.Join(repoPath, "config")
	configContent := `[core]
repositoryformatversion = 0
filemode = false
bare = false`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	repo, err := FindRepo(path, false)
	assert.Error(t, err)
	assert.Nil(t, repo)
}

func TestFindRepo(t *testing.T) {
	// Test case: Repository found in a parent directory
	path := t.TempDir()

	repoPath := filepath.Join(path, ".orf")
	err := os.MkdirAll(repoPath, os.ModePerm)
	assert.NoError(t, err)

	// Create the config file inside .orf
	configPath := filepath.Join(repoPath, "config")
	configContent := `[core]
repositoryformatversion = 0
filemode = false
bare = false`
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	assert.NoError(t, err)

	childPath := filepath.Join(path, "child")
	err = os.MkdirAll(repoPath, os.ModePerm)
	assert.NoError(t, err)

	repo, err := FindRepo(childPath, false)
	assert.Error(t, err)
	assert.Nil(t, repo)
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

func TestGetFilePath(t *testing.T) {
	workTree := t.TempDir()
	filePath, err := GetFilePath(workTree, false, "subdir", "file")
	assert.Error(t, err)
	assert.Equal(t, "", filePath)
}

func TestGetForceFile(t *testing.T) {
	workTree := t.TempDir()
	filePath, err := GetFilePath(workTree, true, "subdir", "file")
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
