package object

import "orf/kv"

// Represents a commit object, with key-value data forming the commit message.
type Commit struct {
	Base
	kvData *kv.OrderedMap
}

func (commit *Commit) GetFormat() string {
	return commit.format
}

func (commit *Commit) GetSize() uint32 {
	return commit.size
}

func (commit *Commit) GetData() []byte {
	return commit.data
}

func (commit *Commit) GetKVData() *kv.OrderedMap {
	return commit.kvData
}

// Creates a new Commit object with empty key-value data.
func CreateCommit(data []byte) *Commit {
	return &Commit{
		Base: Base{
			format: "commit",          // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			data:   data,              // Set the data directly in the Object
		},
		kvData: nil,
	}
}

func (commit *Commit) Serialize() []byte {
	return kv.Serialize(commit.kvData)
}

func (commit *Commit) Deserialize(data []byte) error {
	kvData, err := kv.Parse(data, 0, commit.kvData)
	if err != nil {
		return err
	}

	commit.kvData = kvData
	return nil
}
