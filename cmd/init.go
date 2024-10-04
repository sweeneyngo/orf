package cmd

import (
	"fmt"
	"orf/repository"
)

func Init(path string) error {
	_, err := repository.CreateRepo(path)
	if err != nil {
		return fmt.Errorf("error initializing repo: %v", err)
	}

	return nil
}
