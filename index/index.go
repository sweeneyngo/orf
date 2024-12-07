package index

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"orf/repository"
	"os"
	"path/filepath"
)

type IndexEntry struct {
	CTimeSec   uint64
	CTimeNsec  uint64
	MTimeSec   uint64
	MTimeNsec  uint64
	Dev        uint64
	Ino        uint64
	ModeType   uint32
	ModePerms  uint32
	Uid        uint32
	Gid        uint32
	Fsize      uint32
	Sha        string
	FlagsValid bool
	FlagStaged uint16
	Name       string
}

type Index struct {
	Entries []IndexEntry
	Version uint32
}

func CreateIndex(version uint32, entries []IndexEntry) *Index {
	return &Index{
		Version: version,
		Entries: entries,
	}
}

func ReadIndex(repo *repository.Repo) (*Index, error) {
	// Read index file
	indexFile, err := repository.GetFilePath(repo.WorkTree, false, "index")
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(indexFile); err != nil {
		return nil, err
	}

	// Open file
	file, err := os.Open(indexFile)
	if err != nil {
		return nil, err
	}

	// Get header, signature, version, count
	signature := make([]byte, 4)
	versionBytes := make([]byte, 4)
	countBytes := make([]byte, 4)

	content, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// Get first 4 bytes of header, assign to signature
	copy(signature, content[:4])
	if string(signature) != "DIRC" {
		return nil, fmt.Errorf("invalid signature in index file")
	}

	// Get next 4 bytes of header, assign to version, convert to big endian
	copy(versionBytes, content[4:8])
	version := binary.BigEndian.Uint32(versionBytes)
	if version != 2 {
		return nil, fmt.Errorf("invalid version in index file")
	}

	// Get next 4 bytes of header, assign to count, convert to big endian
	copy(countBytes, content[8:12])
	count := binary.BigEndian.Uint32(countBytes)

	entries := []IndexEntry{}
	entryBytes := content[12:]

	for i := uint32(0); i < count; i++ {

		entry := IndexEntry{}
		entryBytes = entryBytes[i:]

		// Read creation time, as unix timestamps (seconds since epoch, 1970-01-01 00:00:00 UTC)
		entry.CTimeSec = binary.BigEndian.Uint64(entryBytes[0:4])

		// Read creation time, as nanoseconds
		entry.CTimeNsec = binary.BigEndian.Uint64(entryBytes[4:8])

		// Read modification time, as unix timestamps (seconds since epoch, 1970-01-01 00:00:00 UTC)
		entry.MTimeSec = binary.BigEndian.Uint64(entryBytes[8:12])

		// Read modification time, as nanoseconds
		entry.MTimeNsec = binary.BigEndian.Uint64(entryBytes[12:16])

		// Read device number
		entry.Dev = binary.BigEndian.Uint64(entryBytes[16:20])

		// Read inode number
		entry.Ino = binary.BigEndian.Uint64(entryBytes[20:24])

		unused := binary.BigEndian.Uint32(entryBytes[24:26])
		if unused != 0 {
			return nil, fmt.Errorf("unused field in index entry is not zero")
		}

		mode := binary.BigEndian.Uint32(entryBytes[26:28])
		entry.ModeType = mode >> 12

		// Check if mode type is (regular file, symbolic link, gitlink)
		if entry.ModeType != 0b1000 && entry.ModeType != 0b1010 && entry.ModeType != 0b1110 {
			return nil, fmt.Errorf("invalid mode type in index entry")
		}
		entry.ModePerms = mode & 0b111111111

		// Read user id, group id, file size
		entry.Uid = binary.BigEndian.Uint32(entryBytes[28:32])
		entry.Gid = binary.BigEndian.Uint32(entryBytes[32:36])
		entry.Fsize = binary.BigEndian.Uint32(entryBytes[36:40])

		// Read SHA-1 hash
		entry.Sha = fmt.Sprintf("%040x", entryBytes[40:60])

		flags := binary.BigEndian.Uint16(entryBytes[60:62])
		entry.FlagsValid = flags&0b1000000000000000 != 0
		flagExtended := flags&0b0100000000000000 != 0

		if flagExtended {
			return nil, fmt.Errorf("extended flags not implemented")
		}

		entry.FlagStaged = flags & 0b0011000000000000
		// Length of name
		nameLength := flags & 0b0000111111111111

		i += 62

		if nameLength < 0xfff {
			// Check if file[idx + 62:idx + 62 + nameLength] is 0x00
			if entryBytes[62+nameLength] != 0x00 {
				return nil, fmt.Errorf("name in index entry is not null-terminated")
			}
			// Read name
			entry.Name = string(entryBytes[62 : 62+nameLength])
			i += uint32(nameLength) + 1
		} else {
			return nil, fmt.Errorf("extended name length not implemented")
		}

		i = uint32(8 * math.Ceil(float64(i)/8))

		// Append entry to entries
		entries = append(entries, entry)
	}

	return CreateIndex(version, entries), nil
}

func (index *Index) WriteIndex(repo *repository.Repo) error {
	// Open the index file for writing
	indexFilePath := filepath.Join(repo.Directory, "index")
	f, err := os.Create(indexFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// HEADER
	// Write magic bytes
	if _, err := f.Write([]byte("DIRC")); err != nil {
		return err
	}

	// Write the version number (4 bytes, big-endian)
	if err := binary.Write(f, binary.BigEndian, index.Version); err != nil {
		return err
	}

	// Write the number of entries (4 bytes, big-endian)
	if err := binary.Write(f, binary.BigEndian, uint32(len(index.Entries))); err != nil {
		return err
	}

	// ENTRIES
	idx := 0
	for _, e := range index.Entries {
		// Write ctime (2 * uint32)
		if err := binary.Write(f, binary.BigEndian, e.CTimeSec); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, e.CTimeNsec); err != nil {
			return err
		}

		// Write mtime (2 * uint32)
		if err := binary.Write(f, binary.BigEndian, e.MTimeSec); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, e.MTimeNsec); err != nil {
			return err
		}

		// Write dev, ino, mode, uid, gid, fsize (all uint32)
		if err := binary.Write(f, binary.BigEndian, e.Dev); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, e.Ino); err != nil {
			return err
		}

		// Write mode (combine type and permissions into a single uint32)
		mode := byte(e.ModeType<<12) | byte(e.ModePerms)
		if err := binary.Write(f, binary.BigEndian, mode); err != nil {
			return err
		}

		if err := binary.Write(f, binary.BigEndian, e.Uid); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, e.Gid); err != nil {
			return err
		}

		if err := binary.Write(f, binary.BigEndian, e.Fsize); err != nil {
			return err
		}

		// Convert SHA from hex string to bytes (20 bytes)
		shaBytes, err := hex.DecodeString(e.Sha)
		if err != nil {
			return fmt.Errorf("invalid sha: %v", err)
		}
		if _, err := f.Write(shaBytes); err != nil {
			return err
		}

		// Calculate flag_assume_valid (0x1 << 15 if true, else 0)
		flagAssumeValid := uint16(0)
		if e.FlagsValid {
			flagAssumeValid = 1 << 15
		}

		// Calculate name length (0xFFF max)
		nameBytes := []byte(e.Name)
		nameLen := len(nameBytes)
		if nameLen >= 0xFFF {
			nameLen = 0xFFF
		}

		// Merge flags and name length into 2 bytes
		flagsAndLen := flagAssumeValid | e.FlagStaged | uint16(nameLen)
		if err := binary.Write(f, binary.BigEndian, flagsAndLen); err != nil {
			return err
		}

		// Write the name and a final 0x00 byte
		if _, err := f.Write(nameBytes); err != nil {
			return err
		}
		if err := binary.Write(f, binary.BigEndian, byte(0)); err != nil {
			return err
		}

		// Update the index for padding
		idx += 62 + len(nameBytes) + 1

		// Padding: Ensure entry is aligned on an 8-byte boundary
		if idx%8 != 0 {
			pad := 8 - (idx % 8)
			if err := binary.Write(f, binary.BigEndian, make([]byte, pad)); err != nil {
				return err
			}
			idx += pad
		}
	}

	return nil
}
