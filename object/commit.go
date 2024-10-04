package object

import "orf/kv"

type Commit struct {
	Object
	kvData *kv.OrderedMap
}

func CreateCommit(data []byte) *Commit {
	return &Commit{
		Object: Object{
			format: "commit",          // Set the format for a blob
			size:   uint32(len(data)), // Set the size based on the data length
			Data:   data,              // Set the data directly in the Object
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
