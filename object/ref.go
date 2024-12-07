package object

import (
	"fmt"
	"orf/kv"
	"orf/repository"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func ListRefs(repo *repository.Repo, path string) (*kv.OrderedMap, error) {

	if path == "" {
		newPath, err := repository.GetDir(repo.Directory, true, "refs")
		if err != nil {
			return nil, err
		}
		path = newPath
	}

	output := kv.CreateOrderedMap()

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		return files[i].Name() < files[j].Name()
	})

	for _, file := range files {
		newPath := filepath.Join(path, file.Name())

		if file.IsDir() {
			refs, err := ListRefs(repo, newPath)
			if err != nil {
				return nil, err
			}
			output.Add(file.Name(), refs)
		} else {
			data, err := resolveRef(repo, newPath)
			if err != nil {
				return nil, fmt.Errorf("error resolving ref: %v", err)
			}
			output.Add(file.Name(), data)
		}
	}

	return output, nil
}

func ShowRef(repo *repository.Repo, refs *kv.OrderedMap, withHash bool, prefix string) {

	order := refs.GetOrder()
	for _, k := range order {
		v, _ := refs.Get(k)

		if refStr, ok := v.(string); ok {
			if withHash {
				// Print the reference with the hash
				fmt.Printf("%s %s%s\n", refStr, prefix, k)
			} else {
				fmt.Printf("%s%s\n", prefix, k)
			}
		} else if nestedMap, ok := v.(*kv.OrderedMap); ok {
			// If the value is another OrderedMap (a nested map), recurse into it
			newPrefix := prefix
			if newPrefix != "" {
				newPrefix += "/" // Add separator for nested levels
			}
			newPrefix += k
			ShowRef(repo, nestedMap, withHash, newPrefix)
		} else {
			fmt.Println("Unknown type:", v)
		}
	}
}

func CreateRef(repo *repository.Repo, refName string, target string) error {

	refPath, err := repository.GetFilePath(repo.Directory, true, "refs", refName)
	if err != nil {
		return err
	}

	// Open file in refs/refName, then write hash with newline
	err = os.WriteFile(refPath, []byte(target+"\n"), 0644)
	if err != nil {
		return err
	}

	return nil
}

func resolveRef(repo *repository.Repo, ref string) (string, error) {

	path, err := repository.GetFilePath(repo.Directory, true, ref)
	if err != nil {
		return "", err
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return "", err
	}

	if !fileInfo.Mode().IsRegular() {
		// If it's not a regular file, return nil
		return "", nil
	}

	// Open the file for reading
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	var data string
	_, err = fmt.Fscanf(file, "%s", &data)
	if err != nil && err.Error() != "EOF" {
		return "", err
	}

	// Drop final \n, similar to `data[:-1]` in Python
	if len(data) > 0 && data[len(data)-1] == '\n' {
		data = data[:len(data)-1]
	}

	// If the content starts with "ref: ", resolve it, otherwise return the data as is
	if strings.HasPrefix(data, "ref: ") {

		data, err = resolveRef(repo, data[5:])
		if err != nil {
			return "", err
		}

		return data, nil
	}

	return data, nil
}
