package cmd

import (
	"fmt"
	"orf/kv"
	"orf/object"
	"orf/repository"
)

func Tag(name string, target string, willCreateTarget bool) error {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return fmt.Errorf("error finding repo: %v", err)
	}

	if name != "" {
		createTag(repo, name, target, willCreateTarget)
	} else {
		refs, err := listRefs(repo, "")
		if err != nil {
			return fmt.Errorf("error listing refs: %v", err)
		}
		tags, isFound := refs.Get("tags")
		if isFound {
			showRef(repo, tags.(*kv.OrderedMap), true, "tags")
		}
	}

	return nil
}

func createTag(repo *repository.Repo, name string, target string, willCreateTarget bool) {

	// Get the GitObject from the object reference
	hash := object.FindObject(repo, target, "commit", true)
	obj, err := object.ReadObject(repo.Directory, hash)
	if err != nil {
		fmt.Printf("error reading object: %v", err)
	}

	// Assert the object is of type *object.Commit
	commit, ok := obj.(*object.Commit)
	if !ok {
		fmt.Printf("unexpected type %T for commit, expected *Commit", obj)
	}

	if willCreateTarget {
		// Create tag object (commit)
		tag := object.CreateTag(commit.GetData())
		tag.SetKVData("object", hash)
		tag.SetKVData("type", "commit")
		tag.SetKVData("tag", name)
		tag.SetKVData("tagger", "orf <orf@example.com>")
		tag.SetKVData("message", "Tagging commit "+hash)

		tagHash, err := object.WriteObject(repo.Directory, tag)
		if err != nil {
			fmt.Printf("error writing tag object: %v", err)
		}
		createRef(repo, "tags"+name, tagHash)

	} else {
		createRef(repo, "tags"+name, hash)
	}
}
