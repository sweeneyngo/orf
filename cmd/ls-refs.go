package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
)

func ListRefs() {

	repo, err := repository.FindRepo(".", true)
	if err != nil {
		fmt.Printf("error finding repo: %v", err)
	}
	refs, err := object.ListRefs(repo, "")
	if err != nil {
		fmt.Printf("error listing refs: %v", err)
	}
	object.ShowRef(repo, refs, true, "refs")
}
