package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
	"os"
)

func HashObject(path string, format string, store bool) (string, error) {

	directory := ""
	if store {
		// If true, store object in .orf
		repo, err := repository.FindRepo(path, true)
		if err != nil {
			return "", err
		}

		directory = repo.Directory
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("error reading file: %v", err)
	}

	hash, err := GetHash(data, format, directory)
	if err != nil {
		return "", fmt.Errorf("error getting hash: %v", err)
	}
	return hash, nil
}

func GetHash(data []byte, format string, directory string) (string, error) {

	var o object.Object
	switch format {
	case "blob":
		blob := object.CreateBlob(data)
		o = blob
	default:
		return "", fmt.Errorf("invalid format")
	}

	hash, err := object.WriteObject(directory, o)
	if err != nil {
		return "", fmt.Errorf("error writing object: %v", err)
	}

	return hash, nil
}
