package cmd

import (
	"fmt"
	"orf/index"
	"orf/repository"
)

func ListFiles(isVerbose bool) error {

	// TODO: Implement 8.3 ls-files command

	repo, err := repository.FindRepo(".", false)
	if err != nil {
		return err
	}

	index, err := index.ReadIndex(repo)

	if err != nil {
		return err
	}

	for _, entry := range index.Entries {
		// Print entry
		fmt.Printf("%s\n", entry.Name)
		if isVerbose {
			fmt.Printf("  mode_type: %b\n", entry.ModeType)
			fmt.Printf("  mode: %o\n", entry.ModePerms)
			fmt.Printf("  size: %d\n", entry.Fsize)
			fmt.Printf("  sha(blob): %s\n", entry.Sha)
			fmt.Printf("  ctime: %d.%d\n", entry.CTimeSec, entry.CTimeNsec)
			fmt.Printf("  mtime: %d.%d\n", entry.MTimeSec, entry.MTimeNsec)
			fmt.Printf("  device: %d\n", entry.Dev)
			fmt.Printf("  inode: %d\n", entry.Ino)
			fmt.Printf("  uid: %d\n", entry.Uid)
			fmt.Printf("  gid: %d\n", entry.Gid)
			fmt.Printf("  flags_valid: %t\n", entry.FlagsValid)
			fmt.Printf("  flag_staged: %d\n", entry.FlagStaged)
		}

	}
	return nil
}
