package cmd

import (
	"fmt"
	"orf/object"
	"orf/repository"
	"strings"
)

func Log(commit string) error {
	repo, err := repository.FindRepo(".", true)
	if err != nil {
		return fmt.Errorf("error finding repo: %w", err)
	}

	fmt.Println("digraph orflog{")
	fmt.Println("  node[shape=rect]")

	// Start the log from the commit hash provided in args
	log(repo, object.FindObject(repo, commit, "commit", false), make(map[string]struct{}))

	fmt.Println("}")

	return nil
}

func log(repo *repository.Repo, hash string, seen map[string]struct{}) error {

	// If hash already exists, return
	if _, exists := seen[hash]; exists {
		return nil
	}

	// Add hash to set
	seen[hash] = struct{}{}

	commit, err := object.ReadObject(repo.Directory, hash)
	if err != nil {
		return fmt.Errorf("error reading commit: %w", err)
	}

	c, ok := commit.(*object.Commit)
	if !ok {
		return fmt.Errorf("unexpected type %T for Commit, expected *Commit", commit)
	}

	message, err := extractMessage(c)
	if err != nil {
		return err
	}
	fmt.Printf("  c_%s [label=\"%s: %s\"]\n", hash, hash[:8], message)

	// Ensure the format is commit
	if string(c.GetFormat()) != "commit" {
		return fmt.Errorf("no commit object found from hash %v", hash)
	}

	parents, exists := c.GetKVData().Get("parent")
	if !exists {
		// Check if initial commit has no parent
		return nil
	}

	// Recursively check all parents from commit
	var parentHashes []string
	switch v := parents.(type) {
	case string:
		parentHashes = []string{v}
	case []string:
		parentHashes = v
	default:
		return fmt.Errorf("unexpected type for parent: %v", v)
	}

	for _, parentHash := range parentHashes {
		fmt.Printf("  c_%s -> c_%s;\n", hash, parentHash)
		if err := log(repo, parentHash, seen); err != nil {
			return err
		}
	}

	return nil
}

// Extracts and formats the commit message.
func extractMessage(c *object.Commit) (string, error) {
	var message string
	if value, exists := c.GetKVData().Get("message"); exists {
		message = fmt.Sprintf("%s", value)
	}
	message = strings.TrimSpace(message)
	message = strings.ReplaceAll(message, "\\", "\\\\")
	message = strings.ReplaceAll(message, "\"", "\\\"")

	// Print only the first line of message
	if strings.Contains(message, "\n") {
		message = message[:strings.Index(message, "\n")]
	}

	return message, nil
}
