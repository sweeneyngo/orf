package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
)

func RevParse(refType string, name string) error {

	var format string
	if refType != "" {
		format = refType
	} else {
		format = ""
	}

	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return fmt.Errorf("error finding repo: %v", err)
	}

	fmt.Print(object.FindObject(repo, name, format, true))
	return nil
}
