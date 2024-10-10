package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
)

func CatObject(hash string, format string) (object.Object, error) {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return nil, fmt.Errorf("error finding repo: %v", err)
	}

	newObject, err := object.ReadObject(repo.Directory, object.FindObject(repo, hash, format, false))
	if err != nil {
		return nil, fmt.Errorf("error reading object: %v", err)
	}

	return newObject, nil
}
